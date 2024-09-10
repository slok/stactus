package dev

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/slok/stactus/internal/model"
	storagememory "github.com/slok/stactus/internal/storage/memory"
)

var impacts = []model.IncidentImpact{
	model.IncidentImpactMinor,
	model.IncidentImpactMajor,
	model.IncidentImpactCritical,
}

var updateKinds = []model.IncidentUpdateKind{
	model.IncidentUpdateKindInvestigating,
	model.IncidentUpdateKindResolved,
	model.IncidentUpdateKindUpdate,
}

func NewDevelopmentRepository() storagememory.Repository {
	t0 := time.Now().UTC().Add(-365 * 25 * time.Hour)

	systems := []model.System{}
	// Generate systems.
	for i := 0; i < 10; i++ {
		systems = append(systems, model.System{
			ID:          fmt.Sprintf("test-%d", i),
			Name:        fmt.Sprintf("test %d", i),
			Description: fmt.Sprintf("System %d is the testing system", i),
		})
	}

	// Generate IRs.
	irs := []model.IncidentReport{}
	for i := 0; i < 100; i++ {
		start := t0.Add(time.Duration(i) * time.Hour)
		end := start.Add(1 * time.Hour)

		details := []model.IncidentReportDetail{}
		for i := 0; i < rand.Intn(15); i++ {
			details = append(details, model.IncidentReportDetail{
				Description: fmt.Sprintf("something that is a detail %d", i),
				Kind:        updateKinds[rand.Intn(len(updateKinds))],
				TS:          start.Add(time.Duration(i) * time.Minute),
			})
		}

		irs = append(irs, model.IncidentReport{
			ID:       fmt.Sprintf("ir-%d", i),
			Name:     fmt.Sprintf("Incident report %d", i),
			SystemID: systems[rand.Intn(len(systems))].ID,
			Start:    start,
			End:      end,
			Impact:   impacts[rand.Intn(len(impacts))],
			Details:  details,
		})
	}

	return storagememory.NewRepository(systems, irs)
}