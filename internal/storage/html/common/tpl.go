package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/slok/stactus/internal/conventions"
)

const (
	ThemeDirStatic    = "static"
	ThemeDirTemplates = "templates"
)

// ThemeRenderer knows how to render different themes.
type ThemeRenderer struct {
	tpls        *template.Template
	staticFiles map[string]string
}

// NewOSFSThemeRenderer returns a theme rendered based on real OS FS. The directory
// must have `static“ and `templates“ directories.
func NewOSFSThemeRenderer(dir string) (*ThemeRenderer, error) {
	d := os.DirFS(dir)
	return NewThemeRenderer(d, d)
}

// NewThemeRenderer returns a theme rendered based on FS abstractions (OS, memory, S3...).
// Each of the directories must have `static“ and `templates“ directories accordingly.
func NewThemeRenderer(staticFS fs.FS, templatesFS fs.FS) (*ThemeRenderer, error) {
	// Check we have the required directories.
	_, err := fs.Stat(templatesFS, ThemeDirTemplates)
	if err != nil {
		return nil, fmt.Errorf("the required %q directory is missing: %w", ThemeDirTemplates, err)
	}
	_, err = fs.Stat(staticFS, ThemeDirStatic)
	if err != nil {
		return nil, fmt.Errorf("the required %q directory is missing: %w", ThemeDirStatic, err)
	}

	// Discover all template directories to parse.
	templatePaths := []string{}
	tplfs, err := fs.Sub(templatesFS, ThemeDirTemplates)
	if err != nil {
		return nil, fmt.Errorf("could not get %q sub dir: %w", ThemeDirTemplates, err)
	}
	err = fs.WalkDir(tplfs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Directories and non HTML files don't need to be handled.
		extension := strings.ToLower(filepath.Ext(path))
		if d.IsDir() || (extension != ".html" && extension != ".tpl") {
			return nil
		}

		templatePaths = append(templatePaths, filepath.Join(ThemeDirTemplates, path))

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
	stfs, err := fs.Sub(staticFS, ThemeDirStatic)
	if err != nil {
		return nil, fmt.Errorf("could not get %q sub dir: %w", ThemeDirStatic, err)
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

		staticFiles[filepath.Join(conventions.StaticFilesURLPrefix, path)] = string(data)

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
