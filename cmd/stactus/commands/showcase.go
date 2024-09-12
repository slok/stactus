package commands

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/alecthomas/kingpin/v2"
	"github.com/oklog/run"

	appgenerate "github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/dev"
	"github.com/slok/stactus/internal/storage"
	htmlbase "github.com/slok/stactus/internal/storage/html/themes/base"
	htmlsimple "github.com/slok/stactus/internal/storage/html/themes/simple"
)

type ShowcaseCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	outPath string
	siteURL string
}

func NewShowcaseCommand(rootConfig *RootCommand, app *kingpin.Application) *ShowcaseCommand {
	cmd := app.Command("showcase", "Generates a bunch of examples using multiple Atlassian status page as examples.")
	c := &ShowcaseCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("out", "The directory where all the generated files will be written.").Default("./out").StringVar(&c.outPath)
	cmd.Flag("site-url", "The site base url.").Default("").StringVar(&c.siteURL)

	return c
}

func (c *ShowcaseCommand) Name() string { return c.cmd.FullCommand() }
func (c *ShowcaseCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	type client struct {
		Name string
		Path string
		URL  string
	}

	statusPageClients := []client{
		{Name: "Cloudflare", Path: "cloudflare", URL: "https://www.cloudflarestatus.com/"},
		{Name: "Datadog", Path: "datadog", URL: "https://status.datadoghq.com/"},
		{Name: "Digital Ocean", Path: "digitalocean", URL: "https://status.digitalocean.com/"},
		{Name: "Discord", Path: "discord", URL: "https://discordstatus.com/"},
		{Name: "Github", Path: "github", URL: "https://www.githubstatus.com/"},
		{Name: "Grafana", Path: "grafana", URL: "https://status.grafana.com/"},
		{Name: "Hashicorp", Path: "hashicorp", URL: "https://status.hashicorp.com/"},
		{Name: "MIT", Path: "mit", URL: "https://atlas-status.mit.edu/"},
		{Name: "RedisLabs", Path: "redislabs", URL: "https://status.redis.io/"},
		{Name: "Twilio", Path: "twilio", URL: "https://status.twilio.com/"},
	}

	themes := []string{
		themeBase,
		themeSimple,
	}

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

	// Add showcase simple index file.
	{
		g.Add(
			func() error {
				indexFile := path.Join(c.outPath, "index.html")

				showcaseLinks := ""
				for _, client := range statusPageClients {
					for _, theme := range themes {
						url := c.siteURL + "/" + theme + "/" + client.Path
						name := fmt.Sprintf("%s (%s)", client.Name, theme)
						showcaseLinks += fmt.Sprintf(`<div><a href="%s">%s</a><div>`, url, name)
					}
				}

				err := os.WriteFile(indexFile, []byte(`<html>
				<head><title>Stactus showcase</title></head>
				<body>
				<h1>Showcase</h1>
				<p>`+showcaseLinks+`</p>
				</body>
				</html>`), 0666)
				if err != nil {
					return err
				}

				return nil
			},
			func(err error) {
				cancel(err)
			},
		)
	}

	// Add clietn generators.
	for _, client := range statusPageClients {
		client := client

		g.Add(
			func() error {
				logger.Infof("Generating %s example", client.Name)

				// Setup repositories.
				devRepo, err := dev.NewStatusPageRepository(client.URL)
				if err != nil {
					return fmt.Errorf("could not create repository: %w", err)
				}

				// Render a client per theme.
				for _, theme := range themes {
					outPath := path.Join(c.outPath, theme, client.Path)
					siteURL := c.siteURL + "/" + theme + "/" + client.Path

					var uiCreator storage.UICreator
					switch theme {
					case themeBase:
						uiCreator, err = htmlbase.NewGenerator(htmlbase.GeneratorConfig{
							OutPath: outPath,
							SiteURL: siteURL,
							Logger:  logger,
							ThemeCustomization: htmlbase.ThemeCustomization{
								BrandTitle: client.Name,
								BrandURL:   client.URL,
							},
						})
						if err != nil {
							return fmt.Errorf("could not create HTML generator: %w", err)
						}

					case themeSimple:
						uiCreator, err = htmlsimple.NewGenerator(htmlsimple.GeneratorConfig{
							OutPath: outPath,
							SiteURL: siteURL,
							Logger:  logger,
							ThemeCustomization: htmlsimple.ThemeCustomization{
								BrandTitle: client.Name,
								BrandURL:   client.URL,
							},
						})
						if err != nil {
							return fmt.Errorf("could not create HTML generator: %w", err)
						}
					}

					// Generator service.
					genService, err := appgenerate.NewService(appgenerate.ServiceConfig{
						SystemGetter: devRepo,
						IRGetter:     devRepo,
						UICreator:    uiCreator,
						Logger:       logger,
					})
					if err != nil {
						return fmt.Errorf("could not create generation service: %w", err)
					}

					_, err = genService.Generate(ctx, appgenerate.GenerateReq{})
					if err != nil {
						return fmt.Errorf("generation failed: %w", err)
					}
				}

				return nil
			},
			func(err error) {},
		)
	}

	return g.Run()
}
