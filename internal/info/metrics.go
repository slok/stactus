package info

import "sync"

// MetricsRecorder is the service used to record info metrics.
type MetricsRecorder interface {
	RecordInfo(version string)
}

var onceMeasureInfo sync.Once

// Measure Info knows how to measure app information.
func MeasureInfo(m MetricsRecorder) {
	onceMeasureInfo.Do(func() {
		m.RecordInfo(Version)
	})
}
