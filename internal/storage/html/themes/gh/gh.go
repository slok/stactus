package gh

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/util"
	utilerror "github.com/slok/stactus/internal/util/error"
)

var (
	//go:embed all:static
	StaticFs embed.FS
	//go:embed all:templates
	TemplatesFs embed.FS
	// GH theme renderer.
	renderer = utilerror.Must(util.NewThemeRenderer(StaticFs, TemplatesFs))
)

type Generator struct {
	fileManager        util.FileManager
	renderer           util.ThemeRenderer
	outPath            string
	themeCustomization ThemeCustomization
}

type GeneratorConfig struct {
	FileManager        util.FileManager
	OutPath            string
	Logger             log.Logger
	ThemeCustomization ThemeCustomization
}

type ThemeCustomization struct {
	BrandTitle     string
	BrandURL       string
	BannerImageURL string
	LogoURL        string
}

func (c *ThemeCustomization) defaults() error {
	if c.BrandTitle == "" {
		return fmt.Errorf("Brand title is required")
	}
	if c.BrandURL == "" {
		return fmt.Errorf("Brand URL is required")
	}
	if c.BannerImageURL == "" {
		return fmt.Errorf("Banner image is required")
	}
	if c.LogoURL == "" {
		return fmt.Errorf("Logo URL is required")
	}

	return nil
}

func (c *GeneratorConfig) defaults() error {
	if c.FileManager == nil {
		c.FileManager = util.StdFileManager
	}

	if c.OutPath == "" {
		return fmt.Errorf("out path is required")
	}

	// Ensure correct out path.
	c.OutPath = strings.TrimSpace(c.OutPath)
	c.OutPath = strings.TrimSuffix(c.OutPath, "/")
	c.OutPath = c.OutPath + "/"

	if c.Logger == nil {
		c.Logger = log.Noop
	}

	err := c.ThemeCustomization.defaults()
	if err != nil {
		return fmt.Errorf("invalid theme customization: %w", err)
	}

	return nil

}

func NewGenerator(config GeneratorConfig) (*Generator, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Generator{
		fileManager:        config.FileManager,
		renderer:           *renderer,
		outPath:            config.OutPath,
		themeCustomization: config.ThemeCustomization,
	}, nil
}

func (g Generator) CreateUI(ctx context.Context, ui model.UI) error {
	err := g.genStatic(ctx)
	if err != nil {
		return fmt.Errorf("could not generate static files: %w", err)
	}

	err = g.genDashboard(ctx, ui)
	if err != nil {
		return fmt.Errorf("could not generate dashboard: %w", err)
	}

	return nil
}

// genStatic will generate the static files  (static CSS, JS...).
func (g Generator) genStatic(ctx context.Context) error {
	files, err := g.renderer.Statics(ctx)
	if err != nil {
		return fmt.Errorf("could not get static files: %w", err)
	}

	for path, f := range files {
		err := g.fileManager.WriteFile(ctx, g.outPath+path, []byte(f))
		if err != nil {
			return fmt.Errorf("could not write %q static file: %w", path, err)
		}
	}

	return nil
}

// genDashboard will generate the dashboard related files.
func (g Generator) genDashboard(ctx context.Context, ui model.UI) error {
	const (
		fileNameIndex = "index.html"
	)

	type indexData struct {
		BrandTitle     string
		BrandURL       string
		BannerImageURL string
		LogoURL        string
	}

	data := indexData{
		BrandTitle:     g.themeCustomization.BrandTitle,
		BrandURL:       g.themeCustomization.BrandURL,
		BannerImageURL: g.themeCustomization.BannerImageURL,
		LogoURL:        g.themeCustomization.LogoURL,
	}

	// Render index dashboard.
	index, err := g.renderer.Render(ctx, "page_index", data)
	if err != nil {
		return fmt.Errorf("could not render index: %w", err)
	}

	err = g.fileManager.WriteFile(ctx, g.outPath+fileNameIndex, []byte(index))
	if err != nil {
		return fmt.Errorf("could not write index: %w", err)
	}

	return nil
}
