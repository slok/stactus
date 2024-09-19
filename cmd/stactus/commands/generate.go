package commands

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/dev"
	"github.com/slok/stactus/internal/storage"
	htmlbase "github.com/slok/stactus/internal/storage/html/themes/base"
	htmlsimple "github.com/slok/stactus/internal/storage/html/themes/simple"
	"github.com/slok/stactus/internal/storage/iofs"
	"github.com/slok/stactus/internal/storage/prometheus"
)

const (
	themeBase   = "base"
	themeSimple = "simple"
)

type GeneretaCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	stactusFilePath string
	outPath         string
	theme           string
	siteURL         string
	devFixtures     bool
}

const (
	defaultStactusFile = "stactus.yaml"
)

// NewGeneretaCommand returns a generator with the github status page theme.
func NewGeneretaCommand(rootConfig *RootCommand, app *kingpin.Application) *GeneretaCommand {
	cmd := app.Command("generate", "Generates the static pages.")
	c := &GeneretaCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("stactus-file", "The path ot the stactus file.").Short('i').Default(defaultStactusFile).StringVar(&c.stactusFilePath)
	cmd.Flag("out", "The directory where all the generated files will be written.").Required().Short('o').StringVar(&c.outPath)
	cmd.Flag("site-url", "The site base url, if set it will override the one on the stactus configuration.").StringVar(&c.siteURL)
	cmd.Flag("dev-fixtures", "If enabled it will load development fixtures.").BoolVar(&c.devFixtures)
	cmd.Flag("theme", "Select the theme to render").Default(themeSimple).EnumVar(&c.theme, themeBase, themeSimple)

	return c
}

func (c *GeneretaCommand) Name() string { return c.cmd.FullCommand() }
func (c *GeneretaCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	// Open stactus file.
	stactusFileData, err := os.ReadFile(c.stactusFilePath)
	if err != nil {
		return fmt.Errorf("could not load stactus file: %w", err)
	}

	// Setup repository.
	repo := unifiedRepository{}

	if c.devFixtures {
		devRepo := dev.NewAutogeneratedRepository()
		repo.SystemGetter = devRepo
		repo.IncidentReportGetter = devRepo
	} else {
		d := path.Dir(c.stactusFilePath)
		rootFS := os.DirFS(d)
		incidentsFS, err := fs.Sub(rootFS, "incidents")
		if err != nil {
			return fmt.Errorf("incidents directory missing on at the same level of the stactus file: %w", err)
		}
		roRepo, err := iofs.NewReadRepository(ctx, iofs.ReadRepositoryConfig{
			IncidentsFS:     incidentsFS,
			StactusFileData: string(stactusFileData),
			Logger:          logger,
		})
		if err != nil {
			return fmt.Errorf("could not load data: %w", err)
		}

		repo.SystemGetter = roRepo
		repo.IncidentReportGetter = roRepo
		repo.StatusPageSettingsGetter = roRepo
	}

	switch c.theme {
	case themeBase:
		repo.UICreator, err = htmlbase.NewGenerator(htmlbase.GeneratorConfig{
			OutPath: c.outPath,
			Logger:  logger,
		})
		if err != nil {
			return fmt.Errorf("could not create html generator: %w", err)
		}
	case themeSimple:
		repo.UICreator, err = htmlsimple.NewGenerator(htmlsimple.GeneratorConfig{
			OutPath: c.outPath,
			Logger:  logger,
		})
		if err != nil {
			return fmt.Errorf("could not create html generator: %w", err)
		}
	default:
		return fmt.Errorf("unknown theme")
	}

	repo.PromMetricsCreator, err = prometheus.NewFSRepository(prometheus.RepositoryConfig{
		MetricsFilePath: filepath.Join(c.outPath, conventions.PrometheusMetricsPathName),
	})
	if err != nil {
		return fmt.Errorf("could not create prometheus metrics creator: %w", err)
	}

	// Prepare run entrypoints.
	var g run.Group

	// Upper layer context handler.
	{
		g.Add(
			func() error {
				<-ctx.Done()
				logger.Infof("Context cancelled...")
				return nil
			},
			func(err error) {
				cancel(err)
			},
		)
	}

	// Upper layer context handler.
	{
		genService, err := appgenerate.NewService(appgenerate.ServiceConfig{
			SettingsGetter:     repo,
			SystemGetter:       repo,
			IRGetter:           repo,
			UICreator:          repo,
			PromMetricsCreator: repo,
			Logger:             logger,
		})
		if err != nil {
			return fmt.Errorf("could not create generation service: %w", err)
		}

		g.Add(
			func() error {
				_, err := genService.Generate(ctx, appgenerate.GenerateReq{OverrideSiteURL: c.siteURL})
				if err != nil {
					return fmt.Errorf("generation failed: %w", err)
				}

				return nil
			},
			func(err error) {},
		)
	}

	return g.Run()
}

// unifiedRepository is a helper type to manage all repository as a single instance.
type unifiedRepository struct {
	storage.StatusPageSettingsGetter
	storage.SystemGetter
	storage.IncidentReportGetter
	storage.UICreator
	storage.PromMetricsCreator
}
