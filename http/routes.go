package http

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/gofiber/fiber/v2"
	"time"
)

func (s *httpServer) handleIndex(ctx *fiber.Ctx) error {
	user := fileshare.UserFromContext(ctx)
	return ctx.Render("index", &fiber.Map{
		"User": user,
	})
}

func (s *httpServer) handleLogin(ctx *fiber.Ctx) error {
	if user := fileshare.UserFromContext(ctx); user != nil {
		return ctx.Redirect("/")
	}

	return ctx.Render("login", &fiber.Map{})
}

type loginBody struct {
	Nickname string `schema:"nickname,required"`
}

func (s *httpServer) handlePostLogin(ctx *fiber.Ctx) error {
	var body loginBody
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	// TODO: implement some sort of authentication
	user, err := s.users.GetUser(body.Nickname)
	if err != nil {
		return err
	} else if user == nil {
		return newHttpError(fiber.StatusForbidden, "unknown user", fmt.Errorf("no user for nickname %s", body.Nickname))
	}

	token, err := s.auth.GetToken(user.Nickname)
	if err != nil {
		return err
	}

	ctx.Cookie(&fiber.Cookie{Name: authTokenCookieName, Value: token, HTTPOnly: true, Expires: time.Now().Add(7 * 24 * time.Hour)})
	return ctx.Redirect("/")
}

func (s *httpServer) handleLogout(ctx *fiber.Ctx) error {
	fileshare.SetContextWithUser(ctx, nil)

	ctx.ClearCookie(authTokenCookieName)
	return ctx.Redirect("/")
}
