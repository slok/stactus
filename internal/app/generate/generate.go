package generate

import (
	"context"
	"fmt"

	"github.com/slok/stactus/internal/internalerrors"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage"
)

type ServiceConfig struct {
	SystemGetter storage.SystemGetter
	IRGetter     storage.IncidentReportGetter
	UICreator    storage.UICreator

	Logger log.Logger
}

func (c *ServiceConfig) defaults() error {
	if c.SystemGetter == nil {
		return fmt.Errorf("system getter is required")
	}

	if c.IRGetter == nil {
		return fmt.Errorf("IR getter is required")
	}

	if c.UICreator == nil {
		return fmt.Errorf("UI creator is required")
	}

	if c.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	c.Logger = c.Logger.WithValues(log.Kv{"srv": "app.generate.Service"})

	return nil
}

type Service struct {
	sysGetter storage.SystemGetter
	irGetter  storage.IncidentReportGetter
	uiCreator storage.UICreator
	logger    log.Logger
}

func NewService(config ServiceConfig) (*Service, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", internalerrors.ErrNotValid, err)
	}

	return &Service{
		sysGetter: config.SystemGetter,
		irGetter:  config.IRGetter,
		uiCreator: config.UICreator,
		logger:    config.Logger,
	}, nil
}

type GenerateReq struct{}

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

	irsBySystem := map[string][]*model.IncidentReport{}
	for _, ir := range history {
		irsBySystem[ir.SystemID] = append(irsBySystem[ir.SystemID], ir)
	}

	// Add latest update if the incident is ongoing
	var latestUpdate *model.IncidentReportEvent
	if len(history) > 0 && history[0].End.IsZero() && len(history[0].Timeline) > 0 {
		latestUpdate = &history[0].Timeline[0]
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

	// Generate.
	ui := model.UI{
		SystemDetails: systemDetails,
		History:       history,
		LatestUpdate:  latestUpdate,
	}
	err = s.uiCreator.CreateUI(ctx, ui)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not generate UI: %w", err)
	}

	return GenerateResp{}, nil
}
