package authcontroller

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/middle/ware"
)

type Controller struct {
	service  *auth.Service
	config   config.AuthConfig
	attempts *auth.Limiter
}

type request struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Token       string `json:"token"`
}

func New(service *auth.Service, cfg config.AuthConfig) *Controller {
	return &Controller{service: service, config: cfg, attempts: auth.NewLimiter(50, 15*time.Minute)}
}

func (c *Controller) Handler(action func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		if r.Method == http.MethodPost {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if !c.attempts.Allow(ip) {
				w.Header().Set("Retry-After", "900")
				payload.WriteError(ware.LoggerFromContext(ctx), w, errors.E("auth", http.StatusTooManyRequests, errors.M{"message": {"too many attempts; try again later"}}))
				return
			}
		}
		if err := action(w, r); err != nil {
			payload.WriteError(ware.LoggerFromContext(ctx), w, err)
		}
	})
}

func read(r *http.Request) (request, error) {
	var input request
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return input, errors.E("auth[read]", http.StatusUnsupportedMediaType)
	}
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, errors.E("auth[read]", http.StatusBadRequest)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return input, errors.E("auth[read]", http.StatusBadRequest)
	}
	return input, nil
}

func (c *Controller) CSRF(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(c.config.CSRFCookieName())
	token := ""
	if err == nil && auth.ValidToken(cookie.Value) {
		token = cookie.Value
	} else {
		token, err = auth.RandomToken()
		if err != nil {
			return errors.E("auth[csrf]", err)
		}
	}
	http.SetCookie(w, &http.Cookie{Name: c.config.CSRFCookieName(), Value: token, Path: "/", HttpOnly: true, Secure: c.config.CookieSecure, SameSite: http.SameSiteLaxMode})
	return payload.Write(w, &struct {
		Data struct {
			Token string `json:"csrf_token"`
		} `json:"data"`
	}{Data: struct {
		Token string `json:"csrf_token"`
	}{Token: token}})
}

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) error {
	in, err := read(r)
	if err != nil {
		return err
	}
	if err := c.service.Register(r.Context(), in.Email, in.Password, in.DisplayName); err != nil {
		return err
	}
	return accepted(w, "If this address can be registered, a verification email will be sent.")
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) error {
	in, err := read(r)
	if err != nil {
		return err
	}
	session, err := c.service.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		return err
	}
	if cookie, err := r.Cookie(c.config.SessionCookieName()); err == nil {
		if err := c.service.Logout(r.Context(), cookie.Value); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{Name: c.config.SessionCookieName(), Value: session.Token, Path: "/", HttpOnly: true, Secure: c.config.CookieSecure, SameSite: http.SameSiteLaxMode, Expires: session.ExpiresAt, MaxAge: int(c.config.SessionTTL.Seconds())})
	return payload.Write(w, &struct {
		Data auth.User `json:"data"`
	}{Data: session.User})
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(c.config.SessionCookieName()); err == nil {
		if err := c.service.Logout(r.Context(), cookie.Value); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{Name: c.config.SessionCookieName(), Path: "/", HttpOnly: true, Secure: c.config.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (c *Controller) Me(w http.ResponseWriter, r *http.Request) error {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		return errors.E("auth[me]", http.StatusUnauthorized)
	}
	return payload.Write(w, &struct {
		Data auth.User `json:"data"`
	}{Data: user})
}

func (c *Controller) VerifyEmail(w http.ResponseWriter, r *http.Request) error {
	in, err := read(r)
	if err != nil {
		return err
	}
	if err := c.service.VerifyEmail(r.Context(), in.Token); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (c *Controller) ResendVerification(w http.ResponseWriter, r *http.Request) error {
	return c.requestEmail(w, r, "verify_email")
}
func (c *Controller) ForgotPassword(w http.ResponseWriter, r *http.Request) error {
	return c.requestEmail(w, r, "reset_password")
}

func (c *Controller) requestEmail(w http.ResponseWriter, r *http.Request, purpose string) error {
	in, err := read(r)
	if err != nil {
		return err
	}
	if err := c.service.RequestEmail(r.Context(), in.Email, purpose); err != nil {
		return err
	}
	return accepted(w, "If the account is eligible, an email will be sent.")
}

func (c *Controller) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	in, err := read(r)
	if err != nil {
		return err
	}
	if err := c.service.ResetPassword(r.Context(), in.Token, in.Password); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type acceptedResponse struct {
	Message string `json:"message"`
}

func (*acceptedResponse) StatusCode() int { return http.StatusAccepted }
func accepted(w http.ResponseWriter, message string) error {
	return payload.Write(w, &acceptedResponse{Message: message})
}
