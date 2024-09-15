package iofs

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	apiv1 "github.com/slok/stactus/pkg/api/v1"
	"gopkg.in/yaml.v3"

	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/memory"
)

type ReadRepositoryConfig struct {
	IncidentsFS     fs.FS
	StactusFileData string
	Logger          log.Logger
}

func (c *ReadRepositoryConfig) defaults() error {
	if c.StactusFileData == "" {
		return fmt.Errorf("stactus main file data is required")
	}

	if c.IncidentsFS == nil {
		return fmt.Errorf("incidents fs is required")
	}

	return nil
}

type ReadRepository struct {
	memory.Repository
}

func NewReadRepository(ctx context.Context, config ReadRepositoryConfig) (*ReadRepository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	r := &ReadRepository{}

	incidents, err := r.loadIncidents(ctx, config.IncidentsFS)
	if err != nil {
		return nil, fmt.Errorf("could not load incidents: %w", err)
	}

	systems, err := r.loadSystems(ctx, config.StactusFileData)
	if err != nil {
		return nil, fmt.Errorf("could not load systems: %w", err)
	}

	r.Repository = memory.NewRepository(systems, incidents)

	return r, nil
}

func (r ReadRepository) loadSystems(ctx context.Context, data string) ([]model.System, error) {
	spec := apiv1.StactusV1{}
	err := yaml.Unmarshal([]byte(data), &spec)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall YAML systems file correctly: %w", err)
	}

	if spec.Version != apiv1.StactusVersionV1 {
		return nil, fmt.Errorf("unsupported stactus API version")
	}

	systems := []model.System{}
	for _, s := range spec.Systems {
		s := model.System{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
		}

		err := s.Validate()
		if err != nil {
			return nil, fmt.Errorf("invalid system: %w", err)
		}
		systems = append(systems, s)
	}

	if len(systems) == 0 {
		return nil, fmt.Errorf("at least 1 system is required")
	}

	return systems, nil
}

func (r ReadRepository) loadIncidents(ctx context.Context, incidentFS fs.FS) ([]model.IncidentReport, error) {
	incidents := []model.IncidentReport{}

	err := fs.WalkDir(incidentFS, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Directories and non YAML files don't need to be handled.
		extension := strings.ToLower(filepath.Ext(path))
		if info.IsDir() || (extension != ".yml" && extension != ".yaml") {
			return nil
		}

		rawData, err := fs.ReadFile(incidentFS, path)
		if err != nil {
			return fmt.Errorf("could not read manifest %s: %w", path, err)
		}

		is, err := r.loadIncident(ctx, rawData)
		if err != nil {
			return fmt.Errorf("could not load incidents in %q: %w", path, err)
		}

		incidents = append(incidents, is...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk directory: %w", err)
	}

	return incidents, nil
}

func (r ReadRepository) loadIncident(ctx context.Context, data []byte) ([]model.IncidentReport, error) {
	// In case we have multiple YAML in a single file.
	models := []model.IncidentReport{}
	for _, rawData := range splitYAML(data) {
		if len(rawData) == 0 {
			return nil, fmt.Errorf("incident is empty, data is required")
		}

		spec := apiv1.IncidentV1{}
		err := yaml.Unmarshal(data, &spec)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshall YAML incident file correctly: %w", err)
		}

		m, err := r.mapIncidentV1(spec)
		if err != nil {
			return nil, fmt.Errorf("could not map spec to model: %w", err)
		}

		models = append(models, *m)
	}

	return models, nil
}

func (r ReadRepository) mapIncidentV1(s apiv1.IncidentV1) (*model.IncidentReport, error) {
	if s.Version != apiv1.IncidentVersionV1 {
		return nil, fmt.Errorf("unsupported incident API version")
	}

	impact, err := mapImpact(s.Impact)
	if err != nil {
		return nil, err
	}

	tl, err := mapTimeline(s.Timeline)
	if err != nil {
		return nil, err
	}

	m := &model.IncidentReport{
		ID:        s.ID,
		Name:      s.Name,
		SystemIDs: s.Systems,
		Impact:    impact,
		Timeline:  tl,
	}

	err = m.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid incident report: %w", err)
	}

	return m, nil
}

func mapImpact(s string) (model.IncidentImpact, error) {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "":
		return model.IncidentImpactNone, nil
	case "minor":
		return model.IncidentImpactMinor, nil
	case "major":
		return model.IncidentImpactMajor, nil
	case "critical":
		return model.IncidentImpactCritical, nil
	}

	return "", fmt.Errorf("unknown impact: %q", s)
}

func mapTimeline(tl []apiv1.IncidentV1TimelineEvent) ([]model.IncidentReportEvent, error) {
	mtl := []model.IncidentReportEvent{}
	for i, e := range tl {
		// Map TS, this a tricky as we support multiple formats.
		var prevTS time.Time
		if i-1 >= 0 {
			prevTS = mtl[i-1].TS
		}
		rawTS := strings.TrimSpace(e.TS)
		ts, err := mapEventTS(prevTS, rawTS)
		if err != nil {
			return nil, fmt.Errorf("could not map event timestamp: %q", rawTS)
		}

		mtl = append(mtl, model.IncidentReportEvent{
			TS:          ts,
			Kind:        mapEventKind(e),
			Description: strings.TrimSpace(e.Description),
		})
	}

	return mtl, nil
}

func mapEventKind(e apiv1.IncidentV1TimelineEvent) model.IncidentUpdateKind {
	switch {
	case e.Investigating:
		return model.IncidentUpdateKindInvestigating
	case e.Resolved:
		return model.IncidentUpdateKindResolved
	default:
		return model.IncidentUpdateKindUpdate
	}
}

var eventTSAbsolute = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.DateTime,
	"2006-01-02 15:04", // Like time.DateTime but without seconds.
}

var eventTSRelative = []string{
	"15:04",
	time.TimeOnly,
	time.Kitchen,
}

func mapEventTS(prevTS time.Time, s string) (time.Time, error) {
	// If TS starts with "+" means that we need to add it to the previous TS.
	if strings.HasPrefix(s, "+") {
		if prevTS.IsZero() {
			return time.Time{}, fmt.Errorf("can't use event Timestamp '+' format in the first event of the timeline")
		}

		dur, err := time.ParseDuration(s)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time duration: %w", err)
		}

		return prevTS.Add(dur), nil
	}

	// relative time (not day only hour)
	for _, f := range eventTSRelative {
		t, err := time.Parse(f, s)
		if err != nil {
			continue
		}

		if prevTS.IsZero() {
			return time.Time{}, fmt.Errorf("can't use hour based timestamp format in the first event of the timeline")
		}

		// Mix previous day + new hour/minute...
		return time.Date(prevTS.Year(), prevTS.Month(), prevTS.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), prevTS.Location()), nil
	}

	// Finally try mapping the different formats we support, this is the easiest one.
	for _, f := range eventTSAbsolute {
		s = strings.ReplaceAll(s, "/", "-") // We replace `/` with `- `to support both interchangeably.
		t, err := time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse timestamp, unknown format")

}
