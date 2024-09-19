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

	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/common"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

var (
	//go:embed all:static
	staticFs embed.FS
	//go:embed all:templates
	templatesFs embed.FS
)

type Generator struct {
	fileManager        utilfs.FileManager
	renderer           common.ThemeRenderer
	outPath            string
	themeCustomization ThemeCustomization
}

type GeneratorConfig struct {
	FileManager utilfs.FileManager
	OutPath     string
	Logger      log.Logger
	// Renderer is made public so base theme is easily customizable by only changing the HTML templates.
	Renderer           *common.ThemeRenderer
	ThemeCustomization ThemeCustomization
}

type ThemeCustomization struct {
	HistoryIRPerPage int
}

func (c *ThemeCustomization) defaults() error {
	if c.HistoryIRPerPage == 0 {
		c.HistoryIRPerPage = 10
	}

	return nil
}

func (c *GeneratorConfig) defaults() error {
	if c.FileManager == nil {
		c.FileManager = utilfs.StdFileManager
	}

	if c.OutPath == "" {
		return fmt.Errorf("out path is required")
	}

	// Ensure correct out path.
	c.OutPath = strings.TrimSpace(c.OutPath)
	c.OutPath = strings.TrimSuffix(c.OutPath, "/")
	c.OutPath = c.OutPath + "/"

	if c.Renderer == nil {
		rend, err := common.NewThemeRenderer(staticFs, templatesFs)
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
		themeCustomization: config.ThemeCustomization,
	}

	return g, nil
}

func (g Generator) CreateUI(ctx context.Context, ui model.UI) error {
	// Ensure correct out path.
	siteURL := strings.TrimSpace(ui.Settings.URL)
	siteURL = strings.TrimSuffix(siteURL, "/")

	tplCommonData := tplCommonData{
		BrandTitle:            ui.Settings.Name,
		URLPrefix:             siteURL,
		PrometheusMetricsPath: conventions.PrometheusMetricsPathName,
	}
	tplCommonData.HistoryURL = tplCommonData.urlHistory(0, false)

	err := g.genStatic(ctx)
	if err != nil {
		return fmt.Errorf("could not generate static files: %w", err)
	}

	err = g.genDashboard(ctx, ui, tplCommonData)
	if err != nil {
		return fmt.Errorf("could not generate dashboard: %w", err)
	}

	err = g.genHistory(ctx, ui, tplCommonData)
	if err != nil {
		return fmt.Errorf("could not generate history: %w", err)
	}

	err = g.genIRs(ctx, ui, tplCommonData)
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
func (g Generator) genDashboard(ctx context.Context, ui model.UI, tplCommon tplCommonData) error {
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
		tplCommonData: tplCommon,
		AllOK:         len(ui.OpenedIRs) == 0,
	}

	for _, ir := range ui.OpenedIRs {
		data.OngoingIRs = append(data.OngoingIRs, ongoingIRsTplData{
			Name:         ir.Name,
			URL:          tplCommon.urlIRDetail(ir.ID, false),
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

	err = g.fileManager.WriteFile(ctx, g.outPath+tplCommon.urlIndex(true), []byte(index))
	if err != nil {
		return fmt.Errorf("could not write index: %w", err)
	}

	return nil
}

// genHistory will generate the history files.
func (g Generator) genHistory(ctx context.Context, ui model.UI, tplCommon tplCommonData) error {
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
		nextURL := tplCommon.urlHistory(i-1, false)
		previousURL := tplCommon.urlHistory(i+1, false)

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
				URL:          tplCommon.urlIRDetail(ir.ID, false),
				LatestUpdate: latestUpdate,
				StartTS:      ir.Start,
				EndTS:        ir.End,
				Impact:       string(ir.Impact),
			})
		}

		data := tplData{
			tplCommonData: tplCommon,
			NextURL:       nextURL,
			PreviousURL:   previousURL,
			Incidents:     incidents,
		}

		// Render history first page.
		index, err := g.renderer.Render(ctx, "page_history", data)
		if err != nil {
			return fmt.Errorf("could not render index: %w", err)
		}

		err = g.fileManager.WriteFile(ctx, g.outPath+tplCommon.urlHistory(i, true), []byte(index))
		if err != nil {
			return fmt.Errorf("could not write index: %w", err)
		}
	}

	return nil
}

// genIRs will generate the incident report files.
func (g Generator) genIRs(ctx context.Context, ui model.UI, tplCommon tplCommonData) error {
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
			tplCommonData: tplCommon,
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

		err = g.fileManager.WriteFile(ctx, g.outPath+tplCommon.urlIRDetail(ir.ID, true), []byte(index))
		if err != nil {
			return fmt.Errorf("could not write index: %w", err)
		}
	}

	return nil
}

type tplCommonData struct {
	URLPrefix             string
	BrandTitle            string
	HistoryURL            string
	PrometheusMetricsPath string
}

func (t tplCommonData) urlHistory(page int, fileName bool) string {
	u := fmt.Sprintf("history/%d", page)
	if fileName {
		return u + ".html"
	}

	return t.URLPrefix + "/" + u
}

func (t tplCommonData) urlIndex(fileName bool) string {
	const url = "index"
	if fileName {
		return url + ".html"
	}

	return t.URLPrefix + "/" + url
}

func (t tplCommonData) urlIRDetail(irID string, fileName bool) string {
	u := fmt.Sprintf("ir/%s", irID)
	if fileName {
		return u + ".html"
	}

	return t.URLPrefix + "/" + u
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
