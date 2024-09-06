package memory

import (
	"context"
	"sort"

	"github.com/slok/stactus/internal/model"
)

type Repository struct {
	systemsByID         map[string]model.System
	incidentReportsByID map[string]model.IncidentReport
}

func NewRepository(systems []model.System, incidentReports []model.IncidentReport) Repository {
	systemsByID := map[string]model.System{}
	for _, s := range systems {
		systemsByID[s.ID] = s
	}

	irByID := map[string]model.IncidentReport{}
	for _, ir := range incidentReports {
		irByID[ir.ID] = ir
	}

	return Repository{
		systemsByID:         systemsByID,
		incidentReportsByID: irByID,
	}
}

func (r Repository) ListAllSystems(ctx context.Context) ([]model.System, error) {
	ss := []model.System{}
	for _, s := range r.systemsByID {
		ss = append(ss, s)
	}

	// Sort by name.
	sort.SliceStable(ss, func(i, j int) bool { return ss[i].Name < ss[j].Name })

	return ss, nil
}

func (r Repository) ListAllIncidentReports(ctx context.Context) ([]model.IncidentReport, error) {
	irs := []model.IncidentReport{}
	for _, ir := range r.incidentReportsByID {
		irs = append(irs, ir)
	}

	// Sort Latest created.
	sort.SliceStable(irs, func(i, j int) bool { return irs[i].Start.After(irs[j].Start) })

	return irs, nil
}
