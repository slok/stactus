package storage

import (
	"context"

	"github.com/slok/stactus/internal/model"
)

type SystemGetter interface {
	ListAllSystems(ctx context.Context) ([]model.System, error)
}

type IncidentReportGetter interface {
	ListAllIncidentReports(ctx context.Context) ([]model.IncidentReport, error)
}
