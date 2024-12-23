package commands

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing/fstest"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/storage"
	"github.com/slok/stactus/internal/storage/feed"
	htmlcommon "github.com/slok/stactus/internal/storage/html/common"
	htmlsimple "github.com/slok/stactus/internal/storage/html/themes/simple"
	themesimple "github.com/slok/stactus/internal/storage/html/themes/simple"
	"github.com/slok/stactus/internal/storage/iofs"
	"github.com/slok/stactus/internal/storage/prometheus"
)

type ServeCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	stactusFilePath string
	listenAddress   string
}

// NewServeCommand returns a generator with the github status page theme.
func NewServeCommand(rootConfig *RootCommand, app *kingpin.Application) *ServeCommand {
	cmd := app.Command("serve", "Server that serves the generated status pages and auto reloads.")
	c := &ServeCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("stactus-file", "The path ot the stactus file.").Short('i').Default(defaultStactusFile).StringVar(&c.stactusFilePath)
	cmd.Flag("listen-address", "The address where the server will be listening.").Default(":8080").StringVar(&c.listenAddress)

	return c
}

func (c *ServeCommand) Name() string { return c.cmd.FullCommand() }
func (c *ServeCommand) Run(ctx context.Context) (err error) {
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
				logger.Infof("Context cancelled...")
				return nil
			},
			func(err error) {
				cancel(err)
			},
		)
	}

	// Development server.
	{
		// Open stactus file.
		stactusFileData, err := os.ReadFile(c.stactusFilePath)
		if err != nil {
			return fmt.Errorf("could not load stactus file: %w", err)
		}

		// Setup repository.
		d := path.Dir(c.stactusFilePath)
		rootFS := os.DirFS(d)
		incidentsFS, err := fs.Sub(rootFS, "incidents")
		if err != nil {
			return fmt.Errorf("incidents directory missing on at the same level of the stactus file: %w", err)
		}
		roRepo, err := iofs.NewReadRepository(ctx, iofs.ReadRepositoryConfig{
			IncidentsFS:     incidentsFS,
			StactusFileData: string(stactusFileData),
			Logger:          logger,
		})
		if err != nil {
			return fmt.Errorf("could not load data: %w", err)
		}

		// Override templates if required.
		settings, err := roRepo.GetStatusPageSettings(ctx)
		if err != nil {
			return fmt.Errorf("could not retrieve page status settings: %w", err)
		}
		var themeRenderer *htmlcommon.ThemeRenderer
		if settings.Theme.OverrideTPLPath != "" {
			themeRenderer, err = htmlcommon.NewOSFSThemeRenderer(settings.Theme.OverrideTPLPath)
			if err != nil {
				return fmt.Errorf("could not load custom templates theme: %w", err)
			}
		}

		memFS := fstest.MapFS{}
		memFileManager := &memFSFileManager{fs: memFS}

		// Create the UI renderer.
		var repoUICreator storage.UICreator
		switch {
		case settings.Theme.Simple != nil:
			repoUICreator, err = themesimple.NewGenerator(htmlsimple.GeneratorConfig{
				ThemeRenderer: themeRenderer,
				FileManager:   memFileManager,
				OutPath:       "./",
				Logger:        logger,
			})
			if err != nil {
				return fmt.Errorf("could not create html generator: %w", err)
			}
		default:
			return fmt.Errorf("unknown theme")
		}

		repoPromCreator, err := prometheus.NewFSRepository(prometheus.RepositoryConfig{
			FileManager:     memFileManager,
			MetricsFilePath: filepath.Join("./", conventions.PrometheusMetricsPathName),
		})
		if err != nil {
			return fmt.Errorf("could not create prometheus metrics creator: %w", err)
		}

		repoFeedCreator, err := feed.NewFSRepository(feed.RepositoryConfig{
			FileManager:         memFileManager,
			AtomHistoryFilePath: filepath.Join("./", conventions.IRHistoryAtomFeedPathName),
		})
		if err != nil {
			return fmt.Errorf("could not create feed creator: %w", err)
		}

		genService, err := appgenerate.NewService(appgenerate.ServiceConfig{
			SettingsGetter:     roRepo,
			SystemGetter:       roRepo,
			IRGetter:           roRepo,
			UICreator:          repoUICreator,
			PromMetricsCreator: repoPromCreator,
			FeedCreator:        repoFeedCreator,
			Logger:             logger,
		})
		if err != nil {
			return fmt.Errorf("could not create generation service: %w", err)
		}

		_, portS, _ := strings.Cut(c.listenAddress, ":")
		if _, err := strconv.Atoi(portS); err != nil {
			return fmt.Errorf("could not get listen port: %w", err)
		}
		address := "http://127.0.0.1:" + portS

		_, err = genService.Generate(ctx, appgenerate.GenerateReq{OverrideSiteURL: address})
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		staticHandler := http.FileServerFS(memFS)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If index we need to change index.html because of this: https://github.com/golang/go/blob/3d33437c450aa74014ea1d41cd986b6ee6266984/src/net/http/fs.go#L680
			if r.URL.Path == "" || r.URL.Path == "/" {
				r.URL.Path = "index"
			}

			staticHandler.ServeHTTP(w, r)
		})

		server := http.Server{
			Addr:    c.listenAddress,
			Handler: handler,
		}

		g.Add(
			func() error {
				logger.WithValues(log.Kv{"address": c.listenAddress}).Infof("HTTP server listening...")
				return server.ListenAndServe()
			},
			func(err error) {
				logger.Infof("Draining server connections...")
				if err := server.Shutdown(ctx); err != nil {
					logger.Errorf("Error while draining server connections: %s", err)
				}
			},
		)
	}

	return g.Run()
}

type memFSFileManager struct {
	fs fstest.MapFS
}

func (m *memFSFileManager) WriteFile(ctx context.Context, path string, data []byte) error {
	// Sanitize path.
	path = strings.TrimPrefix(path, ".")
	path = strings.TrimPrefix(path, "/")

	// Store without .html extension as this will match the URL.
	path = strings.TrimSuffix(path, ".html")

	m.fs[path] = &fstest.MapFile{Data: data}
	return nil
}
