package model

import (
	"fmt"
	"sort"
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

func (i *IncidentReport) Validate() error {
	if len(i.Timeline) == 0 {
		return fmt.Errorf("timeline is required")
	}

	if i.ID == "" {
		return fmt.Errorf("id is required")
	}

	if i.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Sort desc in the event TS.
	sort.SliceStable(i.Timeline, func(ii, jj int) bool { return i.Timeline[ii].TS.After(i.Timeline[jj].TS) })

	// Set the end of the incident to the resolved event (if any of the events has the resolved kind).
	for _, ev := range i.Timeline {
		if ev.Kind == IncidentUpdateKindResolved {
			i.End = ev.TS
			break
		}
	}

	// Set the start to the first event.
	i.Start = i.Timeline[len(i.Timeline)-1].TS

	return nil
}
