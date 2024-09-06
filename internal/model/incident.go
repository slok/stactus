package model

import (
	"time"
)

type IncidentImpact string

const (
	IncidentImpactMinor    IncidentImpact = "minor"    // Yellow.
	IncidentImpactMajor    IncidentImpact = "major"    // Orange.
	IncidentImpactCritical IncidentImpact = "critical" // Red.
)

type IncidentReport struct {
	ID       string
	Name     string
	SystemID string
	Start    time.Time
	End      time.Time
	Impact   IncidentImpact
	Details  []IncidentReportDetail
}

type IncidentUpdateKind string

const (
	IncidentUpdateKindInvestigating IncidentUpdateKind = "investigating"
	IncidentUpdateKindUpdate        IncidentUpdateKind = "update"
	IncidentUpdateKindResolved      IncidentUpdateKind = "resolved"
)

type IncidentReportDetail struct {
	Description string
	Kind        IncidentUpdateKind
	TS          time.Time
}
