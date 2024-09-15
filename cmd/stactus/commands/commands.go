package commands

import (
	"context"
	"io"
	"time"

	"github.com/alecthomas/kingpin/v2"

	"github.com/slok/stactus/internal/log"
)

const (
	// LoggerTypeDefault is the logger default type.
	LoggerTypeDefault = "default"
	// LoggerTypeJSON is the logger json type.
	LoggerTypeJSON = "json"
)

// Command represents an application command, all commands that want to be executed
// should implement and setup on main.
type Command interface {
	Name() string
	Run(ctx context.Context) error
}

// RootCommand represents the root command configuration and global configuration
// for all the commands.
type RootCommand struct {
	// Global flags.
	Debug                bool
	NoLog                bool
	NoColor              bool
	LoggerType           string
	ShutdownWaitDuration time.Duration

	// Global instances.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Logger log.Logger
}

// NewRootCommand initializes the main root configuration.
func NewRootCommand(app *kingpin.Application) *RootCommand {
	c := &RootCommand{}

	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Debug)
	app.Flag("no-log", "Disable logger.").BoolVar(&c.NoLog)
	app.Flag("no-color", "Disable logger color.").BoolVar(&c.NoColor)
	app.Flag("logger", "Selects the logger type.").Default(LoggerTypeDefault).EnumVar(&c.LoggerType, LoggerTypeDefault, LoggerTypeJSON)
	app.Flag("shutdown-wait-duration", "After the graceful shutdown, the duration that the app will wait before exiting (e.g: Used to give time to external components like network, to update).").Default("0s").DurationVar(&c.ShutdownWaitDuration)

	return c
}
