package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

const (
	StaticURLPrefix = "static"
)

// ThemeRenderer knows how to render different themes.
type ThemeRenderer struct {
	tpls        *template.Template
	staticFiles map[string]string
}

func NewThemeRenderer(staticFS fs.FS, templatesFS fs.FS) (*ThemeRenderer, error) {
	// Discover all template directories to parse.
	templatePaths := []string{}
	err := fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Directories and non HTML files don't need to be handled.
		extension := strings.ToLower(filepath.Ext(path))
		if d.IsDir() || (extension != ".html" && extension != ".tpl") {
			return nil
		}

		templatePaths = append(templatePaths, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not discover template paths: %w", err)
	}

	// Parse all templates.
	templates, err := template.New("base").Funcs(sprig.FuncMap()).ParseFS(templatesFS, templatePaths...)
	if err != nil {
		return nil, fmt.Errorf("could not parse templates: %w", err)
	}

	// Discover all static files.
	staticFiles := map[string]string{}
	stfs, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("could not get subFS on static files: %w", err)
	}
	err = fs.WalkDir(stfs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories.
		if d.IsDir() {
			return nil
		}

		f, err := stfs.Open(path)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		staticFiles[StaticURLPrefix+"/"+path] = string(data)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not load static files: %w", err)
	}

	return &ThemeRenderer{
		tpls:        templates,
		staticFiles: staticFiles,
	}, nil
}

// Render will render theme templates.
func (t *ThemeRenderer) Render(ctx context.Context, tplName string, data any) (string, error) {
	var b bytes.Buffer
	err := t.tpls.ExecuteTemplate(&b, tplName, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// Render will render templates.
func (t *ThemeRenderer) Statics(ctx context.Context) (map[string]string, error) {
	return t.staticFiles, nil
}
