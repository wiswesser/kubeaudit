package label

type Config struct {
	Label string `yaml:"label"`
}

func (config *Config) GetLabel() string {
	if config == nil {
		return ""
	}
	return config.Label
}
