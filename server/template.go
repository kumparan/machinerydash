package server

import (
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

type htmlTemplate struct {
	templates *template.Template
}

func (t *htmlTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
