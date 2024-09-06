package storage

import (
	"context"

	"github.com/slok/stactus/internal/model"
)

type SystemGetter interface {
	ListAllSystems(ctx context.Context) ([]model.System, error)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name SystemGetter

type IncidentReportGetter interface {
	ListAllIncidentReports(ctx context.Context) ([]model.IncidentReport, error)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name IncidentReportGetter
