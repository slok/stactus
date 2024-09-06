package commands

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/slok/stactus/internal/info"
	"github.com/slok/stactus/internal/log"
	metrics "github.com/slok/stactus/internal/metrics/prometheus"
)

type ServerCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	appServerAddress    string
	statusServerAddress string
	healthCheckPath     string
	metricsPath         string
}

// NewServerCommand returns the server command.
func NewServerCommand(rootConfig *RootCommand, app *kingpin.Application) *ServerCommand {
	cmd := app.Command("server", "Executes server.")
	c := &ServerCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("app-listen-address", "Application listen address.").Default(":8080").StringVar(&c.appServerAddress)
	cmd.Flag("status-listen-address", "Application status listen address.").Default(":8081").StringVar(&c.statusServerAddress)
	cmd.Flag("health-check-path", "Health check path.").Default("/status").StringVar(&c.healthCheckPath)
	cmd.Flag("metrics-path", "Metrics path.").Default("/metrics").StringVar(&c.metricsPath)

	return c
}

func (c ServerCommand) Name() string { return c.cmd.FullCommand() }
func (c ServerCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	// Set up metrics with default metrics recorder.
	promRegistry := prometheus.DefaultRegisterer
	metricsRecorder := metrics.NewRecorder(promRegistry)
	info.MeasureInfo(metricsRecorder)

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

	// HTTP admin server.
	{
		logger := logger.WithValues(log.Kv{"addr": c.statusServerAddress, "http-server": "status"})
		mux := http.NewServeMux()

		// Pprof.
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		// Health checks.
		mux.HandleFunc(c.healthCheckPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))

		// Metrics.
		mux.Handle(c.metricsPath, promhttp.Handler())

		// Create server.
		server := http.Server{
			Addr:    c.statusServerAddress,
			Handler: mux,
		}

		g.Add(
			func() error {
				logger.Infof("HTTP server listening...")
				return server.ListenAndServe()
			},
			func(_ error) {
				logger.Infof("Start draining connections")
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				err := server.Shutdown(ctx)
				if err != nil {
					logger.Errorf("Error while shutting down the server: %s", err)
				} else {
					logger.Infof("Server stopped")
				}
			},
		)
	}

	return g.Run()
}
