package config

import "time"

type Config struct {
	Threads   int                `yaml:"threads,omitempty"`
	BatchSize int                `yaml:"batchSize,omitempty"`
	Interval  time.Duration      `yaml:"interval,omitempty"`
	Schemas   map[string][]Table `yaml:"schemas,omitempty"`
}

type Table struct {
	Table    string        `yaml:"table"`
	Batch    int           `yaml:"batch,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty"`
}

type Cfg interface {
	Load() (*Config, error)
}
