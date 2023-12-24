package http

import (
	"errors"
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/gofiber/fiber/v2"
	"strings"
)

const authTokenCookieName = "token"

func (s *httpServer) getUser(authHeader string, authCookie string) (*fileshare.User, error) {
	var token string
	if len(authHeader) > 0 {
		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 {
			return nil, fmt.Errorf("invalid authorization header")
		} else if authParts[0] != "Bearer" {
			return nil, fmt.Errorf("unsupported authorization header: %s", authParts[0])
		}

		token = authParts[1]
	} else if len(authCookie) > 0 {
		token = authCookie
	} else {
		return nil, nil
	}

	nickname, err := s.auth.GetUser(token)
	if errors.Is(err, fileshare.ErrAuthMalformed) {
		return nil, newHttpError(fiber.StatusBadRequest, "malformed bearer token", err)
	} else if errors.Is(err, fileshare.ErrAuthInvalid) {
		return nil, newHttpError(fiber.StatusUnauthorized, "invalid bearer token", err)
	} else if err != nil {
		return nil, fmt.Errorf("failed authenticating: %w", err)
	}

	user, err := s.users.GetUser(nickname)
	if err != nil {
		return nil, fmt.Errorf("failed authenticating: %w", err)
	} else if user == nil {
		return nil, newHttpError(fiber.StatusForbidden, "unknown user", fmt.Errorf("no user for nickname %s", nickname))
	}

	return user, nil
}

func (s *httpServer) newAuthHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if user, err := s.getUser(ctx.Get("Authorization"), ctx.Cookies(authTokenCookieName)); err != nil {
			return err
		} else if user != nil {
			fileshare.SetContextWithUser(ctx, user)
		}

		return ctx.Next()
	}
}
