package contextlib

import (
	"context"

	"github.com/antonpriyma/otus-highload/internal/app/models"
)

type userKey struct{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userKey{}, models.UserID(userID))
}

func GetUserID(ctx context.Context) (models.UserID, bool) {
	userID, ok := ctx.Value(userKey{}).(models.UserID)
	return userID, ok
}
