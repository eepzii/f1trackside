package scrutineer

type Report struct {
	Root *TypeStats
}

type ScanConfig struct {
	Name     string   `yaml:"check"`
	TypeName string   `yaml:"onType"`
	Paths    []string `yaml:"withPaths"`
}

type TypeStats struct {
	Name               string
	IsMissing          bool
	UnsafeDefaultValue int
	Children           map[string]*TypeStats
}
