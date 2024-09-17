package api

const (
	IncidentVersionV1 = "incident/v1"
)

type IncidentV1 struct {
	Version  string                    `yaml:"version"`
	ID       string                    `yaml:"id"`
	Name     string                    `yaml:"name"`
	Impact   string                    `yaml:"impact"`
	Systems  []string                  `yaml:"systems"`
	Timeline []IncidentV1TimelineEvent `yaml:"timeline"`
}

type IncidentV1TimelineEvent struct {
	TS            string `yaml:"ts"`
	Description   string `yaml:"description"`
	Investigating bool   `yaml:"investigating,omitempty"`
	Resolved      bool   `yaml:"resolved,omitempty"`
}
