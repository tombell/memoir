package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tombell/middle"
	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/middleware"
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/controllers/artworkcontroller"
	"github.com/tombell/memoir/internal/controllers/authcontroller"
	"github.com/tombell/memoir/internal/controllers/searchcontroller"
	"github.com/tombell/memoir/internal/controllers/tracklistscontroller"
	"github.com/tombell/memoir/internal/controllers/trackscontroller"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// routes configures the routes of the application.
func routes(
	logger *slog.Logger,
	router *http.ServeMux,
	config *config.Config,
	tracklistStore *trackliststore.Store,
	trackStore *trackstore.Store,
	artworkStore *artworkstore.Store,
	accounts *auth.Service,
) {
	base := middle.Use(
		varyOrigin,
		ware.Logger(logger),
		ware.RequestID(uuid.NewString),
		ware.RequestLogging(),
		ware.CORS(ware.CORSOptions{
			AllowedOrigins:   []string{config.Auth.Origin},
			AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
		}),
	)
	account := middle.Use(base, middleware.CSRF(config.Auth), ware.Recovery())
	signedIn := middle.Use(account, middleware.Session(accounts, config.Auth, false))
	authController := authcontroller.New(accounts, config.Auth)
	router.Handle("GET /auth/csrf", account(authController.Handler(authController.CSRF)))
	router.Handle("POST /auth/register", account(authController.Handler(authController.Register)))
	router.Handle("POST /auth/login", account(authController.Handler(authController.Login)))
	router.Handle("POST /auth/logout", account(authController.Handler(authController.Logout)))
	router.Handle("GET /auth/me", signedIn(authController.Handler(authController.Me)))
	router.Handle("POST /auth/verify-email", account(authController.Handler(authController.VerifyEmail)))
	router.Handle("POST /auth/resend-verification", account(authController.Handler(authController.ResendVerification)))
	router.Handle("POST /auth/forgot-password", account(authController.Handler(authController.ForgotPassword)))
	router.Handle("POST /auth/reset-password", account(authController.Handler(authController.ResetPassword)))

	api := middle.Use(
		base,
		ware.Recovery(),
	)

	authorized := middle.Use(
		base,
		ware.Recovery(),
		middleware.CSRF(config.Auth),
		middleware.Session(accounts, config.Auth, true),
	)

	router.Handle("GET /tracklists", api(rw(tracklistscontroller.Index(trackStore, tracklistStore))))
	router.Handle("GET /tracklists/{id}", api(rw(tracklistscontroller.Show(tracklistStore))))
	router.Handle("POST /tracklists", authorized(rw(tracklistscontroller.Create(tracklistStore))))
	router.Handle("PATCH /tracklists/{id}", authorized(rw(tracklistscontroller.Update(tracklistStore))))
	router.Handle("DELETE /tracklists/{id}", authorized(rw(tracklistscontroller.Delete(tracklistStore))))

	router.Handle("GET /tracks/{id}", api(rw(trackscontroller.Show(trackStore))))

	router.Handle("POST /artwork", authorized(rw(artworkcontroller.Create(artworkStore))))

	router.Handle("GET /tracks/search", api(rw(searchcontroller.Tracks(trackStore))))
	// router.Handle("GET /search/tracklists", api(rw(searchcontroller.Tracklists(trackStore))))

	// TODO: maybe filter on tracks index endpoint
	router.Handle("GET /tracks/mostplayed", api(rw(trackscontroller.MostPlayed(trackStore))))

	router.Handle("OPTIONS /{path...}", api(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO: validate CORS options?
		w.WriteHeader(http.StatusOK)
	})))

	router.Handle("/{path...}", api(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})))
}

func varyOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		next.ServeHTTP(w, r)
	})
}
