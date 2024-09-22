package generate

import (
	"context"
	"fmt"
	"time"

	"github.com/slok/stactus/internal/internalerrors"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage"
)

type ServiceConfig struct {
	SettingsGetter     storage.StatusPageSettingsGetter
	SystemGetter       storage.SystemGetter
	IRGetter           storage.IncidentReportGetter
	UICreator          storage.UICreator
	PromMetricsCreator storage.PromMetricsCreator
	FeedCreator        storage.FeedCreator

	Logger log.Logger
}

func (c *ServiceConfig) defaults() error {
	if c.SettingsGetter == nil {
		return fmt.Errorf("settings getter is required")
	}

	if c.SystemGetter == nil {
		return fmt.Errorf("system getter is required")
	}

	if c.IRGetter == nil {
		return fmt.Errorf("ir getter is required")
	}

	if c.UICreator == nil {
		return fmt.Errorf("ui creator is required")
	}

	if c.PromMetricsCreator == nil {
		return fmt.Errorf("prom metrics creator is required")
	}

	if c.FeedCreator == nil {
		return fmt.Errorf("feed creator is required")
	}

	if c.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	c.Logger = c.Logger.WithValues(log.Kv{"srv": "app.generate.Service"})

	return nil
}

type Service struct {
	settingsGetter storage.StatusPageSettingsGetter
	sysGetter      storage.SystemGetter
	irGetter       storage.IncidentReportGetter
	uiCreator      storage.UICreator
	promCreator    storage.PromMetricsCreator
	feedCreator    storage.FeedCreator
	logger         log.Logger
}

func NewService(config ServiceConfig) (*Service, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", internalerrors.ErrNotValid, err)
	}

	return &Service{
		settingsGetter: config.SettingsGetter,
		sysGetter:      config.SystemGetter,
		irGetter:       config.IRGetter,
		uiCreator:      config.UICreator,
		promCreator:    config.PromMetricsCreator,
		feedCreator:    config.FeedCreator,
		logger:         config.Logger,
	}, nil
}

type GenerateReq struct {
	OverrideSiteURL string
}

func (r *GenerateReq) validate() error {
	return nil
}

type GenerateResp struct {
	Message string
}

func (s Service) Generate(ctx context.Context, req GenerateReq) (GenerateResp, error) {
	// Validate inputs.
	err := req.validate()
	if err != nil {
		return GenerateResp{Message: err.Error()}, internalerrors.ErrNotValid
	}

	settings, err := s.settingsGetter.GetStatusPageSettings(ctx)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not get site settings: %w", err)
	}
	if req.OverrideSiteURL != "" {
		settings.URL = req.OverrideSiteURL
	}

	// Get all systems.
	systems, err := s.sysGetter.ListAllSystems(ctx)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not list systems: %w", err)
	}

	// Get all IRs.
	irs, err := s.irGetter.ListAllIncidentReports(ctx)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not list IRs: %w", err)
	}

	// Prepare data.
	history := []*model.IncidentReport{}
	for _, ir := range irs {
		history = append(history, &ir)
	}

	openedIRs := []*model.IncidentReport{}
	irsBySystem := map[string][]*model.IncidentReport{}
	for _, ir := range history {
		for _, id := range ir.SystemIDs {
			irsBySystem[id] = append(irsBySystem[id], ir)
		}

		if ir.End.IsZero() {
			openedIRs = append(openedIRs, ir)
		}
	}

	systemDetails := []model.SystemDetails{}
	for _, s := range systems {
		var latestIR *model.IncidentReport
		if len(irsBySystem[s.ID]) > 0 {
			latestIR = irsBySystem[s.ID][0]
		}
		systemDetails = append(systemDetails, model.SystemDetails{
			System:   s,
			LatestIR: latestIR,
			IRs:      irsBySystem[s.ID],
		})
	}

	// Calcualate stats.
	stats := model.UIStats{
		TotalSystems: len(systemDetails),
		TotalIRs:     len(history),
		TotalOpenIRs: len(openedIRs),
	}
	var mttrTotalTime time.Duration
	mttrTotalIRs := 0
	for _, ir := range history {
		switch ir.Impact {
		case model.IncidentImpactMinor:
			stats.TotalMinorIRs++
		case model.IncidentImpactMajor:
			stats.TotalMajorIRs++
		case model.IncidentImpactCritical:
			stats.TotalCriticalIRs++
		}

		if ir.Duration != 0 {
			mttrTotalIRs++
			mttrTotalTime += ir.Duration
		}
	}

	if mttrTotalTime != 0 {
		mttr := mttrTotalTime / time.Duration(mttrTotalIRs)
		stats.MTTR = mttr
	}

	// Generate UI.
	ui := model.UI{
		Settings:      *settings,
		Stats:         stats,
		SystemDetails: systemDetails,
		History:       history,
		OpenedIRs:     openedIRs,
	}
	err = s.uiCreator.CreateUI(ctx, ui)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not generate UI: %w", err)
	}

	// Generate metrics.
	err = s.promCreator.CreatePromMetrics(ctx, ui)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not generate prom metrics: %w", err)
	}

	// Generate feeds.
	err = s.feedCreator.CreateHistoryFeed(ctx, ui)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not generate feeds: %w", err)
	}

	return GenerateResp{}, nil
}
