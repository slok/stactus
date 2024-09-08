package commands

import (
	"context"
	"fmt"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/info"
	"github.com/slok/stactus/internal/storage"
)

type GeneretaCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	outPath       string
	appConfigPath string
}

// NewGeneretaCommand returns the generate command.
func NewGeneretaCommand(rootConfig *RootCommand, app *kingpin.Application) GeneretaCommand {
	cmd := app.Command("generate", "Generates the static pages.")
	c := GeneretaCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("out", "The directory where all the generated files will be written.").Default("./gen").StringVar(&c.outPath)
	cmd.Flag("configuration", "The app configuration file path.").Default("./stactus.yaml").StringVar(&c.appConfigPath)

	return c
}

func (c GeneretaCommand) Name() string { return c.cmd.FullCommand() }
func (c GeneretaCommand) Run(ctx context.Context) (err error) {
	fmt.Fprint(c.rootConfig.Stdout, info.Version)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	// Setup repository.
	repo := unifiedRepository{}
	_ = repo

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

	return nil
}

// unifiedRepository is a helper type to manage all repository as a single instance.
type unifiedRepository struct {
	storage.SystemGetter
	storage.IncidentReportGetter
	storage.UICreator
}
