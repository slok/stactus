package commands

import (
	"context"
	"fmt"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/dev"
	"github.com/slok/stactus/internal/storage"
	htmlsimple "github.com/slok/stactus/internal/storage/html/themes/simple"
)

type GeneretaCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	outPath     string
	siteURL     string
	devFixtures bool
}

// NewGeneretaCommand returns a generator with the github status page theme.
func NewGeneretaCommand(rootConfig *RootCommand, app *kingpin.Application) *GeneretaCommand {
	cmd := app.Command("generate", "Generates the static pages.")
	c := &GeneretaCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("out", "The directory where all the generated files will be written.").Default("./out").StringVar(&c.outPath)
	cmd.Flag("site-url", "The site base url.").Default("").StringVar(&c.siteURL)
	cmd.Flag("dev-fixtures", "If enabled it will load development fixtures.").BoolVar(&c.devFixtures)

	return c
}

func (c *GeneretaCommand) Name() string { return c.cmd.FullCommand() }
func (c *GeneretaCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	// Setup repository.
	repo := unifiedRepository{}

	// TODO(slok): Select theme.
	theme := "simple"
	switch theme {
	case "simple":
		repo.UICreator, err = htmlsimple.NewGenerator(htmlsimple.GeneratorConfig{
			OutPath: c.outPath,
			SiteURL: c.siteURL,
			Logger:  logger,
			ThemeCustomization: htmlsimple.ThemeCustomization{
				BrandTitle: "Stactus",
				BrandURL:   "https://github.com/slok/stactus",
			},
		})
		if err != nil {
			return fmt.Errorf("could not create html generator: %w", err)
		}
	default:
		return fmt.Errorf("unknown theme")
	}

	if c.devFixtures {
		devRepo := dev.NewAutogeneratedRepository()
		repo.SystemGetter = devRepo
		repo.IncidentReportGetter = devRepo
	} else {
		return fmt.Errorf("in development, for now only allow using with dev fixtures")
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
			SystemGetter: repo,
			IRGetter:     repo,
			UICreator:    repo,
			Logger:       logger,
		})
		if err != nil {
			return fmt.Errorf("could not create generation service: %w", err)
		}

		g.Add(
			func() error {
				_, err := genService.Generate(ctx, appgenerate.GenerateReq{})
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
	storage.SystemGetter
	storage.IncidentReportGetter
	storage.UICreator
}
