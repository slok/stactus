package commands

import (
	"context"
	"fmt"
	"path"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"
	"github.com/slok/stactus/internal/storage/atlassianstatuspage"
	"golang.org/x/sync/errgroup"
)

type ShowcaseMigrateCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	outPath string
}

// NewShowcaseMigrateCommand returns a generator with the github status page theme.
func NewShowcaseMigrateCommand(rootConfig *RootCommand, app ShowcaseCommand) *ShowcaseMigrateCommand {
	cmd := app.Cmd.Command("migrate", "Will migrate the showcase to stactus files.")
	c := &ShowcaseMigrateCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("out", "The directory where all the generated files will be written.").Required().Short('o').StringVar(&c.outPath)

	return c
}

func (c *ShowcaseMigrateCommand) Name() string { return c.cmd.FullCommand() }
func (c *ShowcaseMigrateCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	// Prepare run entrypoints.
	var g run.Group

	// Upper layer context handler.
	{
		g.Add(
			func() error {
				<-ctx.Done()
				return nil
			},
			func(err error) {
				cancel(err)
			},
		)
	}

	// Add static pages generation.
	g.Add(
		func() error {
			// Generate all themes and clients status pages.
			group, ctx := errgroup.WithContext(ctx)
			group.SetLimit(5)

			for _, client := range showcaseStatusPageClients {
				group.Go(func() error {
					client := client

					logger.Infof("Migrating %s example", client.Name)

					// Setup repositories.
					repo, err := atlassianstatuspage.NewStatusPageRepository(client.URL)
					if err != nil {
						return fmt.Errorf("could not create repository for %q: %w", client.Name, err)
					}

					outPath := path.Join(c.outPath, client.Path)
					err = migrateStatusPageRepository(ctx, logger, repo, outPath)
					if err != nil {
						return fmt.Errorf("could not migrate %q example: %w", client.Name, err)
					}

					return nil
				})
			}

			return group.Wait()
		},
		func(err error) {},
	)

	return g.Run()
}
