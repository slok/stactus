package prometheus

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/model"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

type RepositoryConfig struct {
	FileManager     utilfs.FileManager
	MetricsFilePath string
}

func (c *RepositoryConfig) defaults() error {
	if c.FileManager == nil {
		c.FileManager = utilfs.StdFileManager
	}

	c.MetricsFilePath = filepath.Clean(c.MetricsFilePath)
	if c.MetricsFilePath == "" {
		return fmt.Errorf("metrics file path is required")
	}

	// Let's force the user a defacto standard.
	if !strings.HasSuffix(c.MetricsFilePath, conventions.PrometheusMetricsPathName) {
		return fmt.Errorf("metrics fule must end with 'metrics' path")
	}

	return nil
}

type Repository struct {
	fileManager utilfs.FileManager
	filePath    string
}

func NewFSRepository(config RepositoryConfig) (*Repository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Repository{
		fileManager: config.FileManager,
		filePath:    config.MetricsFilePath,
	}, nil
}

func (r Repository) CreatePromMetrics(ctx context.Context, ui model.UI) error {
	reg := prometheus.NewRegistry()

	metrics, err := r.genMetrics(ui, reg)
	if err != nil {
		return fmt.Errorf("could not create")
	}

	err = r.fileManager.WriteFile(ctx, r.filePath, []byte(metrics))
	if err != nil {
		return fmt.Errorf("could not generate metrics: %w", err)
	}

	return nil
}

func (r Repository) genMetrics(ui model.UI, reg *prometheus.Registry) (string, error) {
	prefix := "stactus"
	constLabels := prometheus.Labels(map[string]string{
		"status_page": ui.Settings.Name,
	})

	// Create metrics.
	mttr := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   prefix,
		Name:        "incident_mttr_seconds",
		Help:        "The MTTR based on all the incident history.",
		ConstLabels: constLabels,
	}, []string{})
	mttr.WithLabelValues().Set(ui.Stats.MTTR.Seconds())

	allSystemsOperational := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   prefix,
		Name:        "all_systems_status",
		Help:        "Tells if all systems are operational or not.",
		ConstLabels: constLabels,
	}, []string{"status_ok"})
	allOK := len(ui.OpenedIRs) == 0
	allSystemsOperational.WithLabelValues(strconv.FormatBool(allOK)).Set(1)

	systemsStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   prefix,
		Name:        "system_status",
		Help:        "Tells Systems are operational or not.",
		ConstLabels: constLabels,
	}, []string{"id", "name", "status_ok", "impact"})
	for _, s := range ui.SystemDetails {
		systemOK := true
		impact := model.IncidentImpactNone
		for _, ir := range s.IRs {
			if !ir.End.IsZero() {
				continue
			}
			systemOK = false
			switch {
			case impact == model.IncidentImpactNone:
				impact = ir.Impact
			case ir.Impact == model.IncidentImpactCritical:
				impact = ir.Impact
			case ir.Impact == model.IncidentImpactMajor && impact == model.IncidentImpactMinor:
				impact = ir.Impact
			}
		}
		systemsStatus.WithLabelValues(s.System.ID, s.System.Name, strconv.FormatBool(systemOK), string(impact)).Set(1)
	}

	openIRs := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   prefix,
		Name:        "open_incident",
		Help:        "The details of open (not resolved) incidents.",
		ConstLabels: constLabels,
	}, []string{"id", "impact"})
	for _, ir := range ui.OpenedIRs {
		openIRs.WithLabelValues(ir.ID, string(ir.Impact)).Set(1)
	}

	// Register metrics.
	reg.MustRegister(
		openIRs,
		allSystemsOperational,
		mttr,
		systemsStatus,
	)

	mfs, err := reg.Gather()
	if err != nil {
		return "", fmt.Errorf("could not gater metrics: %w", err)
	}

	var b bytes.Buffer
	enc := expfmt.NewEncoder(&b, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range mfs {
		err := enc.Encode(mf)
		if err != nil {
			return "", fmt.Errorf("could not enconde metrics: %w", err)
		}
	}

	return b.String(), nil
}
