package model

import (
	"time"
)

type IncidentImpact string

const (
	IncidentImpactNone     IncidentImpact = "none"     // Black.
	IncidentImpactMinor    IncidentImpact = "minor"    // Yellow.
	IncidentImpactMajor    IncidentImpact = "major"    // Orange.
	IncidentImpactCritical IncidentImpact = "critical" // Red.
)

type IncidentReport struct {
	ID        string
	Name      string
	SystemIDs []string
	Start     time.Time
	End       time.Time
	Impact    IncidentImpact
	Timeline  []IncidentReportEvent
}

type IncidentUpdateKind string

const (
	IncidentUpdateKindUpdate        IncidentUpdateKind = "update"
	IncidentUpdateKindInvestigating IncidentUpdateKind = "investigating"
	IncidentUpdateKindResolved      IncidentUpdateKind = "resolved"
)

type IncidentReportEvent struct {
	Description string
	Kind        IncidentUpdateKind
	TS          time.Time
}
