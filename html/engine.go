package html

import (
	"embed"
	"fmt"
	"github.com/gofiber/template/html/v2"
	"io/fs"
	"net/http"
)

//go:embed templates/*.tmpl
var embeddedTemplatesFS embed.FS

func NewEngine() *html.Engine {
	f, err := fs.Sub(embeddedTemplatesFS, "templates")
	if err != nil {
		panic(fmt.Sprintf("cannot load templates filesystem: %v", err))
	}

	return html.NewFileSystem(http.FS(f), ".tmpl")
}
