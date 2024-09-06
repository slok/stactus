package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	gohttpmetrics "github.com/slok/go-http-metrics/metrics"
	gohttpmetricsprom "github.com/slok/go-http-metrics/metrics/prometheus"
)

const prefix = "stactus"

// Types used to avoid collisions with the same interface naming.
type httpRecorder gohttpmetrics.Recorder

// Recorder satisfies multiple interfaces related with metrics measuring
// it will implement Prometheus based metrics backend.
type Recorder struct {
	httpRecorder

	info *prometheus.GaugeVec
}

// NewRecorder returns a new recorder implementation for prometheus.
func NewRecorder(reg prometheus.Registerer) Recorder {
	r := Recorder{
		httpRecorder: gohttpmetricsprom.NewRecorder(gohttpmetricsprom.Config{Registry: reg}),

		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: prefix,
			Subsystem: "app",
			Name:      "info",
			Help:      "The information of the application.",
		}, []string{"version"}),
	}

	reg.MustRegister(
		r.info,
	)

	return r
}

func (r Recorder) RecordInfo(version string) {
	r.info.WithLabelValues(version).Set(1)
}
