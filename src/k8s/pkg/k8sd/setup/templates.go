package setup

import (
	"embed"
	"fmt"
	"path/filepath"
	"text/template"
)

var (
	//go:embed embed
	templates embed.FS
)

func mustTemplate(parts ...string) *template.Template {
	path := filepath.Join(append([]string{"embed"}, parts...)...)
	b, err := templates.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("invalid template %s: %w", path, err))
	}
	return template.Must(template.New(path).Parse(string(b)))
}
