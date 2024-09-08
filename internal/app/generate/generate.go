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
	UICreator    storage.UICreator

	Logger log.Logger
}

func (c *ServiceConfig) defaults() error {
	if c.SystemGetter == nil {
		return fmt.Errorf("system getter is required")
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

	// Generate.
	ui := model.UI{}
	for _, s := range systems {
		ui.SystemDetails = append(ui.SystemDetails, model.SystemDetails{
			System: s,
		})
	}

	err = s.uiCreator.CreateUI(ctx, ui)
	if err != nil {
		return GenerateResp{}, fmt.Errorf("could not generate UI: %w", err)
	}

	return GenerateResp{}, nil
}
