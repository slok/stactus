package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sirupsen/logrus"

	"github.com/slok/stactus/cmd/stactus/commands"
	"github.com/slok/stactus/internal/info"
	"github.com/slok/stactus/internal/log"
	loglogrus "github.com/slok/stactus/internal/log/logrus"
)

// Run runs the main application.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	app := kingpin.New("stactus", "Stactus status page.")
	app.DefaultEnvars()
	rootCmd := commands.NewRootCommand(app)

	// Setup commands (registers flags).
	generateCmd := commands.NewGeneretaCommand(rootCmd, app)
	showcaseCmd := commands.NewShowcaseCommand(rootCmd, app)
	serveCmd := commands.NewServeCommand(rootCmd, app)
	migrateCmd := commands.NewMigrateCommand(app)
	migrateStatusPageCmd := commands.NewMigrateStatusPageCommand(rootCmd, migrateCmd)
	versionCmd := commands.NewVersionCommand(rootCmd, app)

	cmds := map[string]commands.Command{
		generateCmd.Name():          generateCmd,
		showcaseCmd.Name():          showcaseCmd,
		serveCmd.Name():             serveCmd,
		migrateCmd.Name():           migrateCmd,
		migrateStatusPageCmd.Name(): migrateStatusPageCmd,
		versionCmd.Name():           versionCmd,
	}

	// Parse commandline.
	cmdName, err := app.Parse(args[1:])
	if err != nil {
		return fmt.Errorf("invalid command configuration: %w", err)
	}

	// Set standard input/output.
	rootCmd.Stdin = stdin
	rootCmd.Stdout = stdout
	rootCmd.Stderr = stderr

	// Set logger.
	rootCmd.Logger = getLogger(ctx, *rootCmd)

	// New context to control the shutdown from the root command.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Listen for shutdown signals, when signal received, stop main context to start the graceful shutdown.
	signalCtx, signalCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer signalCancel()
	go func() {
		<-signalCtx.Done()
		rootCmd.Logger.Infof("Termination signal received, waiting %s before starting app shutdown...", rootCmd.ShutdownWaitDuration)
		time.Sleep(rootCmd.ShutdownWaitDuration)
		cancel() // Stop the app.
	}()

	err = cmds[cmdName].Run(ctx)

	return err
}

// getLogger returns the application logger.
func getLogger(_ context.Context, config commands.RootCommand) log.Logger {
	if config.NoLog {
		return log.Noop
	}

	// If not logger disabled use logrus logger.
	logrusLog := logrus.New()
	logrusLog.Out = config.Stderr // By default logger goes to stderr (so it can split stdout prints).
	logrusLogEntry := logrus.NewEntry(logrusLog)

	if config.Debug {
		logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
	}

	// Log format.
	switch config.LoggerType {
	case commands.LoggerTypeDefault:
		logrusLogEntry.Logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   !config.NoColor,
			DisableColors: config.NoColor,
		})
	case commands.LoggerTypeJSON:
		logrusLogEntry.Logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger := loglogrus.NewLogrus(logrusLogEntry).WithValues(log.Kv{
		"version": info.Version,
	})

	logger.Debugf("Debug level is enabled") // Will log only when debug enabled.

	return logger
}

func main() {
	ctx := context.Background()
	err := Run(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
