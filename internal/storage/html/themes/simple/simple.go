package simple

import (
	"embed"
	"fmt"

	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/storage/html/common"
	"github.com/slok/stactus/internal/storage/html/themes/base"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

var (
	//go:embed all:static
	staticFs embed.FS
	//go:embed all:templates
	templatesFs embed.FS
)

type GeneratorConfig struct {
	FileManager        utilfs.FileManager
	OutPath            string
	Logger             log.Logger
	ThemeCustomization ThemeCustomization
}

type ThemeCustomization struct {
	HistoryIRPerPage int
}

type Generator struct {
	base.Generator
}

// NewGenerator returns a simple theme using the base theme as the base.
func NewGenerator(config GeneratorConfig) (*Generator, error) {
	rend, err := common.NewThemeRenderer(staticFs, templatesFs)
	if err != nil {
		return nil, fmt.Errorf("could not create theme renderer: %w", err)
	}

	gen, err := base.NewGenerator(base.GeneratorConfig{
		FileManager: config.FileManager,
		OutPath:     config.OutPath,
		Logger:      config.Logger,
		Renderer:    rend,
		ThemeCustomization: base.ThemeCustomization{
			HistoryIRPerPage: config.ThemeCustomization.HistoryIRPerPage,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create UI generator: %w", err)
	}

	return &Generator{Generator: *gen}, nil
}
