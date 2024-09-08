package model

type SystemDetails struct {
	System   System
	LatestIR *IncidentReport
	IRs      []*IncidentReport
}

// UI represents the all the details a UI requires to be generated.
type UI struct {
	SystemDetails []SystemDetails
	History       []*IncidentReport
	LatestUpdate  *IncidentReportDetail
}
