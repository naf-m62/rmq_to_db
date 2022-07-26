package buffer

type (
	Config struct {
		Name  string       `yaml:"name"`
		Redis *ConfigRedis `yaml:"redis"`
		Size  int64        `yaml:"size"`
	}

	ConfigRedis struct {
		Addr        string `yaml:"addr"`
		IdleTimeout int    `yaml:"idleTimeout"`
		Key         string `yaml:"key"`
	}
)
