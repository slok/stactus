package system

import (
	"fmt"

	"github.com/slok/stactus/internal/internalerrors"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/storage"
)

type ServiceConfig struct {
	SystemGetter storage.SystemGetter

	Logger log.Logger
}

func (c *ServiceConfig) defaults() error {
	if c.SystemGetter == nil {
		return fmt.Errorf("system getter is required")
	}
	if c.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	c.Logger = c.Logger.WithValues(log.Kv{"srv": "app.system.Service"})

	return nil
}

type Service struct {
	sysGetter storage.SystemGetter
	logger    log.Logger
}

func NewService(config ServiceConfig) (*Service, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", internalerrors.ErrNotValid, err)
	}

	return &Service{
		sysGetter: config.SystemGetter,
		logger:    config.Logger,
	}, nil
}
