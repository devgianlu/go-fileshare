package http

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/gofiber/fiber/v2"
	"io/fs"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type indexViewData struct {
	User           *fileshare.User
	Files          []fs.DirEntry
	FilesPrefixURL string
}

func (s *httpServer) handleIndex(ctx *fiber.Ctx) error {
	user := fileshare.UserFromContext(ctx)

	var files []fs.DirEntry
	if user != nil {
		var err error
		files, err = s.storage.ReadDir(".", user)
		if err != nil {
			return err
		}
	}

	return ctx.Render("index", &indexViewData{
		User:           user,
		Files:          files,
		FilesPrefixURL: "/",
	})
}

type filesViewData struct {
	Files          []fs.DirEntry
	FilesPrefixURL string
}

func (s *httpServer) handleFiles(ctx *fiber.Ctx) error {
	user := fileshare.UserFromContext(ctx)
	if user == nil {
		return newHttpError(http.StatusForbidden, "cannot see files", fmt.Errorf("unauthenticated users cannot see files"))
	}

	var paths []string
	for i := 1; true; i++ {
		path := ctx.Params(fmt.Sprintf("*%d", i))
		if len(path) == 0 {
			break
		}

		paths = append(paths, path)
	}

	var dir string
	if len(paths) > 0 {
		dir = filepath.Join(paths...)
	} else {
		dir = "."
	}

	files, err := s.storage.ReadDir(dir, user)
	if err != nil {
		return err
	}

	return ctx.Render("files", &filesViewData{
		Files:          files,
		FilesPrefixURL: filepath.Clean(fmt.Sprintf("/%s", dir)) + "/",
	})
}

func (s *httpServer) handleDownload(ctx *fiber.Ctx) error {
	user := fileshare.UserFromContext(ctx)
	if user == nil {
		return newHttpError(http.StatusForbidden, "cannot download files", fmt.Errorf("unauthenticated users cannot download files"))
	}

	var paths []string
	for i := 1; true; i++ {
		path := ctx.Params(fmt.Sprintf("*%d", i))
		if len(path) == 0 {
			break
		}

		paths = append(paths, path)
	}

	var dir string
	if len(paths) > 0 {
		dir = filepath.Join(paths...)
	} else {
		dir = "."
	}

	file, err := s.storage.OpenFile(dir, user)
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}

	if fileInfo.IsDir() {
		_ = file.Close()

		// TODO: download tar.gz / zip archive for directory
		return newHttpError(fiber.StatusNotImplemented, "directory download not implemented yet", fmt.Errorf("TODO download dir"))
	} else {
		ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", strconv.Quote(fileInfo.Name())))

		if fileInfo.Size() >= math.MaxInt {
			// download file chunked
			return ctx.SendStream(file)
		} else {
			return ctx.SendStream(file, int(fileInfo.Size()))
		}
	}
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
