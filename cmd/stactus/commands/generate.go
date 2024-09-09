package commands

import (
	"context"
	"fmt"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage"
	htmlgh "github.com/slok/stactus/internal/storage/html/themes/gh"
	storagememory "github.com/slok/stactus/internal/storage/memory"
)

type GeneretaCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	outPath       string
	appConfigPath string
}

// NewGeneretaCommand returns a generator with the github status page theme.
func NewGeneretaCommand(rootConfig *RootCommand, app *kingpin.Application) *GeneretaCommand {
	cmd := app.Command("generate", "Generates the static pages.")
	c := &GeneretaCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("out", "The directory where all the generated files will be written.").Default("./out").StringVar(&c.outPath)
	cmd.Flag("configuration", "The app configuration file path.").Default("./stactus.yaml").StringVar(&c.appConfigPath)

	return c
}

func (c *GeneretaCommand) Name() string { return c.cmd.FullCommand() }
func (c *GeneretaCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	// Setup repository.
	repo := unifiedRepository{}

	repo.UICreator, err = htmlgh.NewGenerator(htmlgh.GeneratorConfig{
		OutPath: c.outPath,
		Logger:  logger,
		ThemeCustomization: htmlgh.ThemeCustomization{
			BrandTitle:     "Github",
			BrandURL:       "https://github.com",
			BannerImageURL: "https://user-images.githubusercontent.com/19292210/60553863-044dd200-9cea-11e9-987e-7db84449f215.png",
			LogoURL:        "https://raw.githubusercontent.com/gilbarbara/logos/main/logos/github-icon.svg",
		},
	})
	if err != nil {
		return fmt.Errorf("could not create html generator: %w", err)
	}

	repo.SystemGetter = storagememory.NewRepository([]model.System{
		{ID: "test-1", Name: "Test 1", Description: "System 1 is the Test 1 system"},
		{ID: "test-2", Name: "Test 2", Description: "System 2 is the Test 1 system"},
		{ID: "test-3", Name: "Test 3", Description: "System 3 is the Test 1 system"},
	}, []model.IncidentReport{})

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
