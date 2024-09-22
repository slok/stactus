package commands

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"
	"golang.org/x/sync/errgroup"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/storage"
	"github.com/slok/stactus/internal/storage/feed"
	htmlbase "github.com/slok/stactus/internal/storage/html/themes/base"
	htmlsimple "github.com/slok/stactus/internal/storage/html/themes/simple"
	"github.com/slok/stactus/internal/storage/iofs"
	"github.com/slok/stactus/internal/storage/prometheus"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

type showcaseClient struct {
	Name string
	Path string
	URL  string
}

var (
	showcaseStatusPageClients = []showcaseClient{
		{Name: "1Password", Path: "onepassword", URL: "https://status.1password.com/"},
		{Name: "Atlassian Statuspage", Path: "atlassianstatuspage", URL: "https://metastatuspage.com/"},
		{Name: "Bambulab", Path: "bambulab", URL: "https://status.bambulab.com/"},
		{Name: "Cloudflare", Path: "cloudflare", URL: "https://www.cloudflarestatus.com/"},
		{Name: "Datadog", Path: "datadog", URL: "https://status.datadoghq.com/"},
		{Name: "Digital Ocean", Path: "digitalocean", URL: "https://status.digitalocean.com/"},
		{Name: "Discord", Path: "discord", URL: "https://discordstatus.com/"},
		{Name: "Figma", Path: "figma", URL: "https://status.figma.com/"},
		{Name: "FlyIO", Path: "flyio", URL: "https://status.flyio.net/"},
		{Name: "Github", Path: "github", URL: "https://www.githubstatus.com/"},
		{Name: "Grafana", Path: "grafana", URL: "https://status.grafana.com/"},
		{Name: "Hashicorp", Path: "hashicorp", URL: "https://status.hashicorp.com/"},
		{Name: "MIT", Path: "mit", URL: "https://atlas-status.mit.edu/"},
		{Name: "MongoDB", Path: "mongodb", URL: "https://status.mongodb.com/"},
		{Name: "New Relic", Path: "newrelic", URL: "https://status.newrelic.com/"},
		{Name: "Reddit", Path: "reddit", URL: "https://www.redditstatus.com/"},
		{Name: "RedisLabs", Path: "redislabs", URL: "https://status.redis.io/"},
		{Name: "Twilio", Path: "twilio", URL: "https://status.twilio.com/"},
		{Name: "Twitch", Path: "twitch", URL: "https://status.twitch.com"},
		{Name: "Uber", Path: "uber", URL: "https://flgtt5cfx545.statuspage.io/"},
		{Name: "Ubiquiti", Path: "ubiquiti", URL: "https://status.ui.com/"},
		{Name: "Zoom", Path: "zoom", URL: "https://status.zoom.us/"},
	}

	showcaseThemes = []string{
		themeBase,
		themeSimple,
	}
)

type ShowcaseGenerateCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	inPath  string
	outPath string
	siteURL string
}

func NewShowcaseGenerateCommand(rootConfig *RootCommand, app ShowcaseCommand) *ShowcaseGenerateCommand {
	cmd := app.Cmd.Command("generate", "Will generate the showcase from the migrated stactus files (showcase migrate).")
	c := &ShowcaseGenerateCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("in", "The directory where all the generated files will be written.").Required().Short('i').StringVar(&c.inPath)
	cmd.Flag("out", "The directory where all the generated files will be written.").Required().Short('o').StringVar(&c.outPath)
	cmd.Flag("site-url", "The site base url.").Required().Short('u').StringVar(&c.siteURL)

	return c
}

func (c *ShowcaseGenerateCommand) Name() string { return c.cmd.FullCommand() }
func (c *ShowcaseGenerateCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger
	var fileManager utilfs.FileManager = utilfs.StdFileManager

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

			// Generate simple index file.
			indexFile := path.Join(c.outPath, "index.html")

			showcaseLinks := ""
			for _, client := range showcaseStatusPageClients {
				for _, theme := range showcaseThemes {
					url := c.siteURL + "/" + theme + "/" + client.Path
					name := fmt.Sprintf("%s (%s)", client.Name, theme)
					showcaseLinks += fmt.Sprintf(`<div><a href="%s">%s</a><div>`, url, name)
				}
			}

			err := fileManager.WriteFile(ctx, indexFile, []byte(`<html>
				<head><title>Stactus showcase</title></head>
				<body>
				<h1>Showcase</h1>
				<p>`+showcaseLinks+`</p>
				</body>
				</html>`))
			if err != nil {
				return err
			}

			// Generate all themes and clients status pages.
			group, ctx := errgroup.WithContext(ctx)
			group.SetLimit(5)

			for _, client := range showcaseStatusPageClients {
				group.Go(func() error {
					client := client

					logger.Infof("Generating %s example", client.Name)

					// Setup repositories.
					stactusFilePath := filepath.Join(c.inPath, client.Path, defaultStactusFile)
					stactusFileData, err := os.ReadFile(stactusFilePath)
					if err != nil {
						return fmt.Errorf("could not load stactus file: %w", err)
					}

					rootFS := os.DirFS(path.Dir(stactusFilePath))
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

					// Render a client per theme.
					for _, theme := range showcaseThemes {
						outPath := path.Join(c.outPath, theme, client.Path)
						siteURL := c.siteURL + "/" + theme + "/" + client.Path

						var uiCreator storage.UICreator
						switch theme {
						case themeBase:
							uiCreator, err = htmlbase.NewGenerator(htmlbase.GeneratorConfig{
								OutPath: outPath,
								Logger:  logger,
							})
							if err != nil {
								return fmt.Errorf("could not create HTML generator: %w", err)
							}

						case themeSimple:
							uiCreator, err = htmlsimple.NewGenerator(htmlsimple.GeneratorConfig{
								OutPath: outPath,
								Logger:  logger,
							})
							if err != nil {
								return fmt.Errorf("could not create HTML generator: %w", err)
							}
						}

						promRepo, err := prometheus.NewFSRepository(prometheus.RepositoryConfig{
							MetricsFilePath: filepath.Join(outPath, conventions.PrometheusMetricsPathName),
						})
						if err != nil {
							return fmt.Errorf("could not create prometheus metrics creator: %w", err)
						}

						repoFeedCreator, err := feed.NewFSRepository(feed.RepositoryConfig{
							AtomHistoryFilePath: filepath.Join(filepath.Join(outPath, conventions.PrometheusMetricsPathName), conventions.IRHistoryAtomFeedPathName),
						})
						if err != nil {
							return fmt.Errorf("could not create feed creator: %w", err)
						}

						// Generator service.
						genService, err := appgenerate.NewService(appgenerate.ServiceConfig{
							SettingsGetter:     roRepo,
							SystemGetter:       roRepo,
							IRGetter:           roRepo,
							UICreator:          uiCreator,
							PromMetricsCreator: promRepo,
							FeedCreator:        repoFeedCreator,
							Logger:             logger,
						})
						if err != nil {
							return fmt.Errorf("could not create generation service: %w", err)
						}

						_, err = genService.Generate(ctx, appgenerate.GenerateReq{
							OverrideSiteURL: siteURL,
						})
						if err != nil {
							return fmt.Errorf("generation failed: %w", err)
						}
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
