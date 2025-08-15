package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TotalPriority        int     `yaml:"totalPriority"`
	EnableAntiStarvation bool    `yaml:"enableAntiStarvation"`
	MaximumWaitSeconds   int     `yaml:"maximumWaitSeconds"`
	BTreeDegree          int     `yaml:"btreeDegree"`
	TimeBoost            float64 `yaml:"timeBoost"`
}

// LoadConfig reads YAML from disk.
func LoadConfig(path string) (Config, error) {
	var cfg Config
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	// sensible defaults if missing
	if cfg.TotalPriority <= 0 {
		cfg.TotalPriority = 1
	}
	if cfg.BTreeDegree <= 0 {
		cfg.BTreeDegree = 16
	}
	return cfg, nil
}
