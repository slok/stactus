package commands

import (
	"context"
	"fmt"

	"github.com/alecthomas/kingpin/v2"

	"github.com/slok/stactus/internal/info"
)

type VersionCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand
}

// NewVersionCommand returns the version command.
func NewVersionCommand(rootConfig *RootCommand, app *kingpin.Application) VersionCommand {
	cmd := app.Command("version", "Shows version.")
	c := VersionCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	return c
}

func (v VersionCommand) Name() string { return v.cmd.FullCommand() }
func (v VersionCommand) Run(_ context.Context) error {
	fmt.Fprint(v.rootConfig.Stdout, info.Version)
	return nil
}
