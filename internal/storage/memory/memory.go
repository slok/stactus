package memory

import (
	"context"
	"slices"

	"github.com/slok/stactus/internal/model"
)

type Repository struct {
	settings        model.StatusPageSettings
	systems         []model.System
	incidentReports []model.IncidentReport
}

func NewRepository(systems []model.System, settings model.StatusPageSettings, incidentReports []model.IncidentReport) Repository {
	return Repository{
		settings:        settings,
		systems:         slices.Clone(systems),
		incidentReports: slices.Clone(incidentReports),
	}
}

func (r Repository) GetStatusPageSettings(ctx context.Context) (*model.StatusPageSettings, error) {
	return &r.settings, nil
}

func (r Repository) ListAllSystems(ctx context.Context) ([]model.System, error) {
	return slices.Clone(r.systems), nil
}

func (r Repository) ListAllIncidentReports(ctx context.Context) ([]model.IncidentReport, error) {
	return slices.Clone(r.incidentReports), nil
}
