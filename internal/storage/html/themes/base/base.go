package base

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/util"
)

var (
	//go:embed all:static
	staticFs embed.FS
	//go:embed all:templates
	templatesFs embed.FS
)

type Generator struct {
	fileManager        util.FileManager
	renderer           util.ThemeRenderer
	outPath            string
	siteURL            string
	themeCustomization ThemeCustomization
	tplCommonData      tplCommonData
}

type GeneratorConfig struct {
	FileManager util.FileManager
	OutPath     string
	SiteURL     string
	Logger      log.Logger
	// Renderer is made public so base theme is easily customizable by only changing the HTML templates.
	Renderer           *util.ThemeRenderer
	ThemeCustomization ThemeCustomization
}

type ThemeCustomization struct {
	BrandTitle       string
	BrandURL         string
	HistoryIRPerPage int
}

func (c *ThemeCustomization) defaults() error {
	if c.BrandTitle == "" {
		return fmt.Errorf("Brand title is required")
	}
	if c.BrandURL == "" {
		return fmt.Errorf("Brand URL is required")
	}

	if c.HistoryIRPerPage == 0 {
		c.HistoryIRPerPage = 10
	}

	return nil
}

func (c *GeneratorConfig) defaults() error {
	if c.FileManager == nil {
		c.FileManager = util.StdFileManager
	}

	if c.OutPath == "" {
		return fmt.Errorf("out path is required")
	}

	// Ensure correct out path.
	c.OutPath = strings.TrimSpace(c.OutPath)
	c.OutPath = strings.TrimSuffix(c.OutPath, "/")
	c.OutPath = c.OutPath + "/"

	// Ensure correct out path.
	c.SiteURL = strings.TrimSpace(c.SiteURL)
	c.SiteURL = strings.TrimSuffix(c.SiteURL, "/")
	c.SiteURL = c.SiteURL + "/"

	if c.Renderer == nil {
		rend, err := util.NewThemeRenderer(staticFs, templatesFs)
		if err != nil {
			return fmt.Errorf("could not create theme renderer: %w", err)
		}
		c.Renderer = rend
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}

	err := c.ThemeCustomization.defaults()
	if err != nil {
		return fmt.Errorf("invalid theme customization: %w", err)
	}

	return nil

}

type tplCommonData struct {
	URLPrefix  string
	BrandTitle string
	BrandURL   string
	HistoryURL string
}

// NewGenerator returns a base theme generator, this is the simplest theme, it can be used as a base (hence the name)
// to create new themes on top instead of creating a new one from scratch.
func NewGenerator(config GeneratorConfig) (*Generator, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	g := &Generator{
		fileManager:        config.FileManager,
		renderer:           *config.Renderer,
		outPath:            config.OutPath,
		siteURL:            config.SiteURL,
		themeCustomization: config.ThemeCustomization,
	}

	g.tplCommonData = tplCommonData{
		URLPrefix:  strings.TrimSuffix(g.siteURL, "/"),
		BrandTitle: g.themeCustomization.BrandTitle,
		BrandURL:   g.themeCustomization.BrandURL,
		HistoryURL: g.urlHistory(0, false),
	}

	return g, nil
}

func (g Generator) CreateUI(ctx context.Context, ui model.UI) error {
	err := g.genStatic(ctx)
	if err != nil {
		return fmt.Errorf("could not generate static files: %w", err)
	}

	err = g.genDashboard(ctx, ui)
	if err != nil {
		return fmt.Errorf("could not generate dashboard: %w", err)
	}

	err = g.genHistory(ctx, ui)
	if err != nil {
		return fmt.Errorf("could not generate history: %w", err)
	}

	err = g.genIRs(ctx, ui)
	if err != nil {
		return fmt.Errorf("could not generate IRs details: %w", err)
	}

	return nil
}

// genStatic will generate the static files  (static CSS, JS...).
func (g Generator) genStatic(ctx context.Context) error {
	files, err := g.renderer.Statics(ctx)
	if err != nil {
		return fmt.Errorf("could not get static files: %w", err)
	}

	for path, f := range files {
		err := g.fileManager.WriteFile(ctx, g.outPath+path, []byte(f))
		if err != nil {
			return fmt.Errorf("could not write %q static file: %w", path, err)
		}
	}

	return nil
}

// genDashboard will generate the dashboard related files.
func (g Generator) genDashboard(ctx context.Context, ui model.UI) error {
	type System struct {
		Name        string
		Description string
		OK          bool
		Impact      string
	}

	type ongoingIRsTplData struct {
		Name         string
		URL          string
		LatestUpdate string
		TS           time.Time
		Impact       string
	}

	type tplData struct {
		tplCommonData
		AllOK      bool
		OngoingIRs []ongoingIRsTplData
		Systems    []System
	}

	data := tplData{
		tplCommonData: g.tplCommonData,
		AllOK:         len(ui.OpenedIRs) == 0,
	}

	for _, ir := range ui.OpenedIRs {
		data.OngoingIRs = append(data.OngoingIRs, ongoingIRsTplData{
			Name:         ir.Name,
			URL:          g.urlIRDetail(ir.ID, false),
			LatestUpdate: renderMarkdown(ir.Timeline[0].Description),
			TS:           ir.Timeline[0].TS,
			Impact:       string(ir.Impact),
		})
	}

	for _, s := range ui.SystemDetails {
		ok := true
		impact := model.IncidentImpactNone
		if s.LatestIR != nil && s.LatestIR.End.IsZero() {
			ok = false
			impact = s.LatestIR.Impact
		}
		data.Systems = append(data.Systems, System{
			Name:        s.System.Name,
			Description: s.System.Description,
			OK:          ok,
			Impact:      string(impact),
		})
	}

	// Render index dashboard.
	index, err := g.renderer.Render(ctx, "page_index", data)
	if err != nil {
		return fmt.Errorf("could not render index: %w", err)
	}

	err = g.fileManager.WriteFile(ctx, g.outPath+g.urlIndex(true), []byte(index))
	if err != nil {
		return fmt.Errorf("could not write index: %w", err)
	}

	return nil
}

// genHistory will generate the history files.
func (g Generator) genHistory(ctx context.Context, ui model.UI) error {
	type incidentTplData struct {
		Title        string
		URL          string
		LatestUpdate string
		StartTS      time.Time
		EndTS        time.Time
		Impact       string
	}

	type tplData struct {
		tplCommonData
		NextURL     string
		PreviousURL string
		Incidents   []incidentTplData
	}

	// Split incidents in pages.
	pageIncidents := [][]*model.IncidentReport{}
	for i := 0; i < len(ui.History); i += g.themeCustomization.HistoryIRPerPage {
		end := i + g.themeCustomization.HistoryIRPerPage
		if end > len(ui.History) {
			end = len(ui.History)
		}
		pageIncidents = append(pageIncidents, ui.History[i:end])
	}

	// Render a history per page.
	for i, page := range pageIncidents {
		nextURL := g.urlHistory(i-1, false)
		previousURL := g.urlHistory(i+1, false)

		// Special page cases (first, last).
		switch {
		case i == 0:
			nextURL = ""
		case len(pageIncidents)-1 == i:
			previousURL = ""
		}

		incidents := []incidentTplData{}
		for _, ir := range page {
			latestUpdate := ""
			if len(ir.Timeline) > 0 {
				latestUpdate = renderMarkdown(ir.Timeline[0].Description)
			}

			incidents = append(incidents, incidentTplData{
				Title:        ir.Name,
				URL:          g.urlIRDetail(ir.ID, false),
				LatestUpdate: latestUpdate,
				StartTS:      ir.Start,
				EndTS:        ir.End,
				Impact:       string(ir.Impact),
			})
		}

		data := tplData{
			tplCommonData: g.tplCommonData,
			NextURL:       nextURL,
			PreviousURL:   previousURL,
			Incidents:     incidents,
		}

		// Render history first page.
		index, err := g.renderer.Render(ctx, "page_history", data)
		if err != nil {
			return fmt.Errorf("could not render index: %w", err)
		}

		err = g.fileManager.WriteFile(ctx, g.outPath+g.urlHistory(i, true), []byte(index))
		if err != nil {
			return fmt.Errorf("could not write index: %w", err)
		}
	}

	return nil
}

// genIRs will generate the incident report files.
func (g Generator) genIRs(ctx context.Context, ui model.UI) error {
	type timelineTplData struct {
		Kind   string
		TS     time.Time
		Detail string
	}

	type tplData struct {
		tplCommonData
		Title    string
		ID       string
		Impact   string
		StartTS  time.Time
		EndTS    time.Time
		Duration time.Duration
		Timeline []timelineTplData
	}

	// Render a IR per page.
	for _, ir := range ui.History {
		var duration time.Duration
		if !ir.End.IsZero() {
			duration = ir.End.Sub(ir.Start)
		}

		timeline := []timelineTplData{}
		for _, d := range ir.Timeline {
			timeline = append(timeline, timelineTplData{
				Kind:   string(d.Kind),
				TS:     d.TS,
				Detail: renderMarkdown(d.Description),
			})
		}

		data := tplData{
			tplCommonData: g.tplCommonData,
			Title:         ir.Name,
			ID:            ir.ID,
			Impact:        string(ir.Impact),
			StartTS:       ir.Start,
			EndTS:         ir.End,
			Duration:      duration,
			Timeline:      timeline,
		}

		// Render history first page.
		index, err := g.renderer.Render(ctx, "page_ir", data)
		if err != nil {
			return fmt.Errorf("could not render index: %w", err)
		}

		err = g.fileManager.WriteFile(ctx, g.outPath+g.urlIRDetail(ir.ID, true), []byte(index))
		if err != nil {
			return fmt.Errorf("could not write index: %w", err)
		}
	}

	return nil
}

func (g Generator) urlHistory(page int, fileName bool) string {
	u := fmt.Sprintf("history/%d", page)
	if fileName {
		return u + ".html"
	}

	return g.siteURL + u
}

func (g Generator) urlIndex(fileName bool) string {
	const url = "index"
	if fileName {
		return url + ".html"
	}

	return g.siteURL + url
}

func (g Generator) urlIRDetail(irID string, fileName bool) string {
	u := fmt.Sprintf("ir/%s", irID)
	if fileName {
		return u + ".html"
	}

	return g.siteURL + u
}

func renderMarkdown(md string) string {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	renderer := html.NewRenderer(html.RendererOptions{Flags: htmlFlags})

	return string(markdown.Render(doc, renderer))
}
