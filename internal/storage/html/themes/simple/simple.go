package simple

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/util"
	utilerror "github.com/slok/stactus/internal/util/error"
)

var (
	//go:embed all:static
	staticFs embed.FS
	//go:embed all:templates
	templatesFs embed.FS
	// Theme renderer.
	renderer = utilerror.Must(util.NewThemeRenderer(staticFs, templatesFs))
)

type Generator struct {
	fileManager        util.FileManager
	renderer           util.ThemeRenderer
	outPath            string
	siteURL            string
	themeCustomization ThemeCustomization
}

type GeneratorConfig struct {
	FileManager        util.FileManager
	OutPath            string
	SiteURL            string
	Logger             log.Logger
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

	if c.Logger == nil {
		c.Logger = log.Noop
	}

	err := c.ThemeCustomization.defaults()
	if err != nil {
		return fmt.Errorf("invalid theme customization: %w", err)
	}

	return nil

}

func NewGenerator(config GeneratorConfig) (*Generator, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Generator{
		fileManager:        config.FileManager,
		renderer:           *renderer,
		outPath:            config.OutPath,
		siteURL:            config.SiteURL,
		themeCustomization: config.ThemeCustomization,
	}, nil
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
	}

	type tplData struct {
		CSSURL     string
		BrandTitle string
		BrandURL   string
		HasUpdate  bool
		UpdateText string
		HistoryURL string
		Systems    []System
	}

	data := tplData{
		CSSURL:     g.urlCSS(false),
		BrandTitle: g.themeCustomization.BrandTitle,
		BrandURL:   g.themeCustomization.BrandURL,
		HistoryURL: g.urlHistory(0, false),
	}

	if ui.LatestUpdate != nil {
		data.HasUpdate = true
		data.UpdateText = ui.LatestUpdate.Description
	}

	for _, s := range ui.SystemDetails {
		ok := true
		if s.LatestIR != nil && s.LatestIR.End.IsZero() {
			ok = false
		}
		data.Systems = append(data.Systems, System{
			Name:        s.System.Name,
			Description: s.System.Description,
			OK:          ok,
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
		StartTS      string
		EndTS        string
		Impact       string
	}

	type tplData struct {
		CSSURL      string
		BrandTitle  string
		BrandURL    string
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
				latestUpdate = ir.Timeline[0].Description
			}
			endTS := ""
			if !ir.End.IsZero() {
				endTS = historyTS(ir.End)
			}

			incidents = append(incidents, incidentTplData{
				Title:        ir.Name,
				URL:          g.urlIRDetail(ir.ID, false),
				LatestUpdate: latestUpdate,
				StartTS:      historyTS(ir.Start),
				EndTS:        endTS,
				Impact:       string(ir.Impact),
			})
		}

		data := tplData{
			CSSURL:      g.urlCSS(false),
			BrandTitle:  g.themeCustomization.BrandTitle,
			BrandURL:    g.themeCustomization.BrandURL,
			NextURL:     nextURL,
			PreviousURL: previousURL,
			Incidents:   incidents,
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
		TS     string
		Detail string
	}

	type tplData struct {
		CSSURL     string
		BrandTitle string
		Title      string
		ID         string
		Impact     string
		StartTS    string
		EndTS      string
		Duration   string
		Timeline   []timelineTplData
		IndexURL   string
	}

	// Render a IR per page.
	for _, ir := range ui.History {
		endTS := ""
		duration := ""
		if !ir.End.IsZero() {
			endTS = historyTS(ir.End)
			duration = ir.End.Sub(ir.Start).String()
		}

		timeline := []timelineTplData{}
		for _, d := range ir.Timeline {
			timeline = append(timeline, timelineTplData{
				Kind:   strTitle(string(d.Kind)),
				TS:     historyTS(d.TS),
				Detail: d.Description,
			})
		}

		data := tplData{
			CSSURL:     g.urlCSS(false),
			BrandTitle: g.themeCustomization.BrandTitle,
			Title:      ir.Name,
			ID:         ir.ID,
			Impact:     strTitle(string(ir.Impact)),
			StartTS:    historyTS(ir.Start),
			EndTS:      endTS,
			Duration:   duration,
			Timeline:   timeline,
			IndexURL:   g.urlIndex(false),
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

func (g Generator) urlCSS(fileName bool) string {
	u := "static/main.css"
	if fileName {
		return u + u
	}

	return g.siteURL + u
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

func historyTS(t time.Time) string {
	return t.Format(`Jan _2, 15:04`)
}

var enCaser = cases.Title(language.English)

func strTitle(s string) string {
	return enCaser.String(s)
}
