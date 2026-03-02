package scrutineer

import "github.com/eepzii/f1trackside/internal/scrutineer/schema"

type Scrutineer struct {
	Root *schema.Field
}

type YamlConfig struct {
	Name     string   `yaml:"check"`
	TypeName string   `yaml:"onType"`
	Paths    []string `yaml:"withPaths"`
}
