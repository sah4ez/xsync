package config

type ConfigSQL struct {
}

func (cs *ConfigSQL) Load() (*Config, error) {
	return &Config{}, nil
}
