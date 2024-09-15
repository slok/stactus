package api

const (
	StactusVersionV1 = "stactus/v1"
)

type StactusV1 struct {
	Version   string            `yaml:"version"`
	BrandName string            `yaml:"brandName"`
	BrandURL  string            `yaml:"brandURL"`
	Systems   []StactusV1System `yaml:"systems"`
}

type StactusV1System struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}
