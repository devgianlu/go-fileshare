package fileshare

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

type contextKey int

const (
	userContextKey = contextKey(iota + 1)
)

func SetContextWithUser(ctx *fiber.Ctx, user *User) {
	parent := ctx.UserContext()
	if parent == nil {
		parent = context.Background()
	}

	ctx.SetUserContext(context.WithValue(parent, userContextKey, user))
}

func UserFromContext(ctx *fiber.Ctx) *User {
	parent := ctx.UserContext()
	if parent == nil {
		parent = context.Background()
	}

	user, _ := parent.Value(userContextKey).(*User)
	return user
}
