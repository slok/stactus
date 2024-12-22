package minimalistic

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/common"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

var (
	//go:embed all:static
	staticFs embed.FS
	//go:embed all:templates
	templatesFs embed.FS
)

type Generator struct {
	fileManager utilfs.FileManager
	renderer    common.ThemeRenderer
	outPath     string
}

type GeneratorConfig struct {
	FileManager   utilfs.FileManager
	OutPath       string
	Logger        log.Logger
	ThemeRenderer *common.ThemeRenderer
}

func (c *GeneratorConfig) defaults() error {
	if c.FileManager == nil {
		c.FileManager = utilfs.StdFileManager
	}

	if c.OutPath == "" {
		return fmt.Errorf("out path is required")
	}

	// Ensure correct out path.
	c.OutPath = strings.TrimSpace(c.OutPath)
	c.OutPath = strings.TrimSuffix(c.OutPath, "/")
	c.OutPath = c.OutPath + "/"

	if c.ThemeRenderer == nil {
		rend, err := common.NewThemeRenderer(staticFs, templatesFs)
		if err != nil {
			return fmt.Errorf("could not create theme renderer: %w", err)
		}
		c.ThemeRenderer = rend
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}

	return nil
}

// NewGenerator returns a base theme generator, this is the simplest theme, it can be used as a base (hence the name)
// to create new themes on top instead of creating a new one from scratch.
func NewGenerator(config GeneratorConfig) (*Generator, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	g := &Generator{
		fileManager: config.FileManager,
		renderer:    *config.ThemeRenderer,
		outPath:     config.OutPath,
	}

	return g, nil
}

func (g Generator) CreateUI(ctx context.Context, ui model.UI) error {
	// Ensure correct out path.
	siteURL := strings.TrimSpace(ui.Settings.URL)
	siteURL = strings.TrimSuffix(siteURL, "/")

	tplCommonData := tplCommonData{
		BrandTitle:            ui.Settings.Name,
		URLPrefix:             siteURL,
		PrometheusMetricsPath: conventions.PrometheusMetricsPathName,
		AtomHistoryFeedPath:   conventions.IRHistoryAtomFeedPathName,
	}

	err := g.genStatic(ctx)
	if err != nil {
		return fmt.Errorf("could not generate static files: %w", err)
	}

	err = g.genSinglePage(ctx, ui, tplCommonData)
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
func (g Generator) genSinglePage(ctx context.Context, ui model.UI, tplCommon tplCommonData) error {
	data := map[string]any{} // TODO(slok).

	// Render index dashboard.
	index, err := g.renderer.Render(ctx, "page_index", data)
	if err != nil {
		return fmt.Errorf("could not render index: %w", err)
	}

	err = g.fileManager.WriteFile(ctx, conventions.IndexFilePath(g.outPath), []byte(index))
	if err != nil {
		return fmt.Errorf("could not write index: %w", err)
	}

	return nil
}

type tplCommonData struct {
	URLPrefix             string
	BrandTitle            string
	PrometheusMetricsPath string
	AtomHistoryFeedPath   string
}
