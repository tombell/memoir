package auth

import (
	"context"
	"github.com/tombell/memoir/internal/database"
)

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
}

func userModel(row database.User) User {
	return User{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, EmailVerified: row.EmailVerifiedAt.Valid}
}

type userContextKey struct{}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey{}).(User)
	return user, ok
}
