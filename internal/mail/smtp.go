// Package mail sends plain text account emails over SMTP.
package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/tombell/memoir/internal/config"
)

type SMTP struct{ config config.SMTPConfig }

func New(cfg config.SMTPConfig) *SMTP { return &SMTP{config: cfg} }

func (s *SMTP) Send(ctx context.Context, to, subject, body string) error {
	if strings.ContainsAny(to+subject+s.config.From, "\r\n") {
		return fmt.Errorf("invalid email header")
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid recipient")
	}
	host, _, err := net.SplitHostPort(s.config.Address)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", s.config.Address)
	if err != nil {
		return err
	}
	defer conn.Close()
	rawConn := conn
	stop := context.AfterFunc(ctx, func() { _ = rawConn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	if s.config.TLSMode == "tls" {
		secured := tls.Client(conn, tlsConfig)
		if err := secured.HandshakeContext(ctx); err != nil {
			return err
		}
		conn = secured
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	switch s.config.TLSMode {
	case "starttls":
		if supported, _ := client.Extension("STARTTLS"); !supported {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	case "tls", "none":
	default:
		return fmt.Errorf("invalid SMTP TLS mode")
	}
	if s.config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.config.Username, s.config.Password, host)); err != nil {
			return err
		}
	}
	if err := client.Mail(s.config.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n", s.config.From, to, subject, strings.ReplaceAll(body, "\n", "\r\n"))
	if _, err := fmt.Fprint(w, message); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
