package http

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

type httpError struct {
	statusCode int
	message    string
	err        error
}

func (e *httpError) Error() string {
	return e.err.Error()
}

func (e *httpError) Unwrap() error {
	return e.err
}

func newHttpError(statusCode int, message string, err error) error {
	return &httpError{statusCode, message, err}
}

func asHttpError(err error) (bool, int, string) {
	var httpErr *httpError
	if !errors.As(err, &httpErr) {
		return false, 0, ""
	}

	return true, httpErr.statusCode, httpErr.message
}

func newErrorHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		err := ctx.Next()
		if err == nil {
			return nil
		}

		if ok, statusCode, message := asHttpError(err); ok {
			// set status code and message header
			ctx.Status(statusCode)
			ctx.Set("X-Error-Message", message)

			// return the error for the logger to see, we'll stop it in the error handler
			return err
		}

		// unhandled error, let it propagate
		ctx.Status(fiber.StatusInternalServerError)
		return err
	}
}
