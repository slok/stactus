package commands

import (
	"context"

	"github.com/alecthomas/kingpin/v2"
)

type MigrateCommand struct {
	Cmd *kingpin.CmdClause
}

// NewMigrateCommand returns the vault command.
func NewMigrateCommand(app *kingpin.Application) MigrateCommand {
	cmd := app.Command("migrate", "Migrate related commands.")
	c := MigrateCommand{Cmd: cmd}

	return c
}

func (c MigrateCommand) Name() string { return c.Cmd.FullCommand() }
func (c MigrateCommand) Run(ctx context.Context) error {
	return nil
}
