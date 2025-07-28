package config

var cfg *config

type config struct {
	LogFile string
	Debug   bool
}

type option func(c *config)

func WithLogFile(path string) option {
	return func(c *config) {
		c.LogFile = path
	}
}

func WithDebug(debug bool) option {
	return func(c *config) {
		c.Debug = debug
	}
}

func Setup(opts ...option) {
	cfg = new(config)
	for _, o := range opts {
		o(cfg)
	}
}

func Config() *config {
	return cfg
}
