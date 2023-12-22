package http

import (
	"github.com/gofiber/fiber/v2"
)

func (s *httpServer) handleIndex(ctx *fiber.Ctx) error {
	return ctx.Render("index", &fiber.Map{})
}
