package http

import (
	"github.com/devgianlu/go-fileshare"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"time"
)

func newLogger() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()
		end := time.Now()

		entry := Log(ctx).WithField("latency", end.Sub(start).String())
		if err != nil {
			entry.WithError(err).Error()
		} else {
			entry.Info()
		}

		return err
	}
}

func Log(ctx *fiber.Ctx) *logrus.Entry {
	user := fileshare.UserFromContext(ctx)

	entry := logrus.WithField("method", ctx.Method()).
		WithField("path", ctx.Path()).
		WithField("status", ctx.Response().StatusCode()).
		WithField("auth", user != nil)

	return entry
}
