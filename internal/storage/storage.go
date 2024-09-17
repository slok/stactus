package storage

import (
	"context"

	"github.com/slok/stactus/internal/model"
)

type StatusPageSettingsGetter interface {
	GetStatusPageSettings(ctx context.Context) (*model.StatusPageSettings, error)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name StatusPageSettingsGetter

type SystemGetter interface {
	ListAllSystems(ctx context.Context) ([]model.System, error)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name SystemGetter

type IncidentReportGetter interface {
	ListAllIncidentReports(ctx context.Context) ([]model.IncidentReport, error)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name IncidentReportGetter

type UICreator interface {
	CreateUI(ctx context.Context, ui model.UI) error
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name UICreator
