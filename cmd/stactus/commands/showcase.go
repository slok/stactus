package commands

import (
	"context"

	"github.com/alecthomas/kingpin/v2"
)

type ShowcaseCommand struct {
	Cmd *kingpin.CmdClause
}

// NewShowcaseCommand returns the vault command.
func NewShowcaseCommand(app *kingpin.Application) ShowcaseCommand {
	cmd := app.Command("showcase", "Showcase related commands, these will be used to create stactus theme and example showcases nothing useful for regular stactus users.")
	c := ShowcaseCommand{Cmd: cmd}

	return c
}

func (c ShowcaseCommand) Name() string { return c.Cmd.FullCommand() }
func (c ShowcaseCommand) Run(ctx context.Context) error {
	return nil
}
