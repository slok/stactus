package model

import "time"

type SystemDetails struct {
	System   System
	LatestIR *IncidentReport
	IRs      []*IncidentReport
}

type UIStats struct {
	TotalSystems     int
	TotalIRs         int
	TotalMinorIRs    int
	TotalMajorIRs    int
	TotalCriticalIRs int
	TotalOpenIRs     int
	MTTR             time.Duration
}

// UI represents the all the details a UI requires to be generated.
type UI struct {
	Stats         UIStats
	Settings      StatusPageSettings
	SystemDetails []SystemDetails
	History       []*IncidentReport
	OpenedIRs     []*IncidentReport
}
