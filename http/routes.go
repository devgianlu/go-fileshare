package http

import (
	"errors"
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"github.com/devgianlu/go-fileshare/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"io"
	"io/fs"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type indexViewData struct {
	User              *fileshare.User
	Files             []fs.DirEntry
	FilesPrefixURL    string
	FilesCanWriteHere bool
}

func (s *httpServer) handleIndex(ctx *fiber.Ctx) error {
	user := fileshare.UserFromContext(ctx)

	var canWrite bool
	var files []fs.DirEntry
	if user != nil {
		canWrite = s.storage.CanWrite(".", user)

		var err error
		files, err = s.storage.ReadDir(".", user)
		if err != nil {
			return err
		}
	}

	return ctx.Render("index", &indexViewData{
		User:              user,
		Files:             files,
		FilesPrefixURL:    "/",
		FilesCanWriteHere: canWrite,
	})
}

type filesViewData struct {
	Files             []fs.DirEntry
	FilesPrefixURL    string
	FilesCanWriteHere bool
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
		Files:             files,
		FilesPrefixURL:    filepath.Clean(fmt.Sprintf("/%s", dir)) + "/",
		FilesCanWriteHere: s.storage.CanWrite(dir, user),
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

	var path string
	if len(paths) > 0 {
		path = filepath.Join(paths...)
	} else {
		path = "."
	}

	// open file for stats and eventually reading
	file, stat, err := s.storage.OpenFile(path, user)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		// fix root archive name
		name := stat.Name()
		if name == "." {
			name = "files"
		}

		ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", strconv.Quote(name+".tar.gz")))

		return compressFolderToArchive(s.storage, user, path, ctx)
	} else {
		ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", strconv.Quote(stat.Name())))

		if stat.Size() >= math.MaxInt {
			// download file chunked
			return ctx.SendStream(file)
		} else {
			return ctx.SendStream(file, int(stat.Size()))
		}
	}
}

func (s *httpServer) handleUpload(ctx *fiber.Ctx) error {
	user := fileshare.UserFromContext(ctx)
	if user == nil {
		return newHttpError(http.StatusForbidden, "cannot upload files", fmt.Errorf("unauthenticated users cannot upload files"))
	}

	var paths []string
	for i := 1; true; i++ {
		path := ctx.Params(fmt.Sprintf("*%d", i))
		if len(path) == 0 {
			break
		}

		paths = append(paths, path)
	}

	var path string
	if len(paths) > 0 {
		path = filepath.Join(paths...)
	} else {
		path = "."
	}

	form, err := ctx.MultipartForm()
	if errors.Is(err, fasthttp.ErrNoMultipartForm) {
		return newHttpError(http.StatusBadRequest, "missing form", err)
	}

	formFiles, ok := form.File["file"]
	if !ok {
		return newHttpError(http.StatusBadRequest, "missing files", err)
	}

	for _, formFile := range formFiles {
		uploadFile, err := formFile.Open()
		if err != nil {
			return err
		}

		localFile, err := s.storage.CreateFile(filepath.Join(path, formFile.Filename), user)
		if err != nil {
			_ = uploadFile.Close()
			return err
		}

		if _, err := io.Copy(localFile, uploadFile); err != nil {
			_ = localFile.Close()
			_ = uploadFile.Close()
			return err
		}

		_ = localFile.Close()
		_ = uploadFile.Close()
	}

	return ctx.Redirect("/files/" + strings.Join(paths, "/"))
}

type loginViewData struct {
	PasswordAuth bool
	GithubAuth   bool
}

func (s *httpServer) handleLogin(ctx *fiber.Ctx) error {
	if user := fileshare.UserFromContext(ctx); user != nil && !user.Anonymous() {
		return ctx.Redirect("/")
	}

	_, passwordOk := s.auth[auth.AuthProviderTypePassword]
	_, githubOk := s.auth[auth.AuthProviderTypeGithub]

	return ctx.Render("login", &loginViewData{
		PasswordAuth: passwordOk,
		GithubAuth:   githubOk,
	})
}

type loginBody struct {
	Provider string `schema:"provider,required"`

	// Only for "passwd" provider
	Nickname string `schema:"nickname"`
	Password string `schema:"password"`
}

func (s *httpServer) handlePostLogin(ctx *fiber.Ctx) error {
	var body loginBody
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	provider, ok := s.auth[body.Provider]
	if !ok {
		return newHttpError(fiber.StatusBadRequest, "invalid auth provider", fmt.Errorf("unknown auth provider: %s", body.Provider))
	}

	switch provider := provider.(type) {
	case fileshare.OAuth2AuthProvider:
		url, err := provider.Callback()
		if err != nil {
			return err
		}

		return ctx.Redirect(url)
	default:
		var providerPayload any
		switch body.Provider {
		case auth.AuthProviderTypePassword:
			providerPayload = auth.PasswordAuthProviderPayload{Nickname: body.Nickname, Password: body.Password}
		default:
			panic("provider not implemented")
		}

		nickname, err := provider.Authenticate(providerPayload)
		if err != nil {
			return newHttpError(fiber.StatusUnauthorized, "invalid auth credentials", err)
		}

		// if we get here, authentication is good
		user, err := s.users.GetUser(nickname)
		if err != nil {
			return err
		} else if user == nil {
			return newHttpError(fiber.StatusForbidden, "unknown user", fmt.Errorf("no user for nickname %s", body.Nickname))
		}

		token, err := s.tokens.GetToken(nickname)
		if err != nil {
			return err
		}

		ctx.Cookie(&fiber.Cookie{Name: authTokenCookieName, Value: token, HTTPOnly: true, Expires: time.Now().Add(7 * 24 * time.Hour)})
		return ctx.Redirect("/")
	}
}

func (s *httpServer) handleOauthLoginCallback(ctx *fiber.Ctx) error {
	code, state := ctx.Query("code"), ctx.Query("state")
	if len(code) == 0 || len(state) == 0 {
		return newHttpError(fiber.StatusBadRequest, "missing code or state", fmt.Errorf("missing code (%s) or state (%s)", code, state))
	}

	providerKey := ctx.Params("provider")
	provider_, ok := s.auth[providerKey]
	if !ok {
		return newHttpError(fiber.StatusBadRequest, "provider not available", fmt.Errorf("auth provider %s is not configured", providerKey))
	}

	provider, ok := provider_.(fileshare.OAuth2AuthProvider)
	if !ok {
		return newHttpError(fiber.StatusBadRequest, "provider not available", fmt.Errorf("auth provider %s is not oauth2", providerKey))
	}

	nickname, err := provider.Authenticate(fileshare.OAuth2ProviderPayload{Code: code, State: state})
	if err != nil {
		return newHttpError(fiber.StatusUnauthorized, "invalid auth credentials", err)
	}

	// if we get here, authentication is good
	user, err := s.users.GetUser(nickname)
	if err != nil {
		return err
	} else if user == nil {
		return newHttpError(fiber.StatusForbidden, "unknown user", fmt.Errorf("no user for nickname %s", nickname))
	}

	token, err := s.tokens.GetToken(nickname)
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
