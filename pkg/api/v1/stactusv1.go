package api

const (
	StactusVersionV1 = "stactus/v1"
)

type StactusV1 struct {
	Version string            `yaml:"version"`
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Theme   *StactusV1Theme   `yaml:"theme,omitempty"`
	Systems []StactusV1System `yaml:"systems"`
}

type StactusV1System struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type StactusV1Theme struct {
	Simple *StactusV1ThemeSimple `yaml:"simple,omitempty"`
}

type StactusV1ThemeSimple struct {
	ThemePath string `yaml:"themePath,omitempty"`
}
