package api

import (
	"context"

	"github.com/riportdev/riport/share/logger"
)

type userCtxKeyType string

const userCtxKey userCtxKeyType = "user"

// WithUser returns a copy of a given context that contains a given username.
func WithUser(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, userCtxKey, username)
}

// GetUser returns a username from a given context.
func GetUser(ctx context.Context, log *logger.Logger) string {
	userValue := ctx.Value(userCtxKey)
	user, ok := userValue.(string)
	if !ok {
		log.Errorf("incorrect type: expected string, actual %T: %v", userValue, userValue)
		return ""
	}
	return user
}
