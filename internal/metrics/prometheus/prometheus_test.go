package prometheus_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	metrics "github.com/slok/stactus/internal/metrics/prometheus"
)

func TestRecorder(t *testing.T) {
	tests := map[string]struct {
		measure    func(rec metrics.Recorder)
		expMetrics string
	}{
		"Info metric should be measured.": {
			measure: func(rec metrics.Recorder) {
				rec.RecordInfo("v99.42")
			},
			expMetrics: `
# HELP stactus_app_info The information of the application.
# TYPE stactus_app_info gauge
stactus_app_info{version="v99.42"} 1
			`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			reg := prometheus.NewRegistry()
			rec := metrics.NewRecorder(reg)
			test.measure(rec)

			err := testutil.GatherAndCompare(reg, strings.NewReader(test.expMetrics))
			require.NoError(err)
		})
	}
}
