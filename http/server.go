package http

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/devgianlu/go-fileshare/html"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"
)

type httpServer struct {
	port int

	log *logrus.Entry
	app *fiber.App

	auth fileshare.AuthProvider
}

func NewHTTPServer(port int, auth fileshare.AuthProvider) fileshare.HttpServer {
	s := httpServer{}
	s.log = logrus.WithField("module", "http")
	s.port = port
	s.auth = auth

	s.app = fiber.New(fiber.Config{
		Views: html.NewEngine(),
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// prevent our errors from propagating, they have already been handled
			if ok, _, _ := asHttpError(err); ok {
				return nil
			}

			return err
		},
	})
	s.app.Use(
		newLogger(),       // logs requests
		newErrorHandler(), // handles custom errors
		recover.New(recover.Config{EnableStackTrace: true}), // handles panics
		s.newAuthHandler(), // handles authentication
	)

	s.app.Get("/", s.handleIndex)

	return &s
}

func (s *httpServer) ListenForever() error {
	return s.app.Listen(fmt.Sprintf("0.0.0.0:%d", s.port))
}
