package config

import (
	"fmt"
	"os"
	"time"
)

var cfg *config

type config struct {
	LogFile    string
	LspLogFile string
	Debug      bool
	TimeoutLen time.Duration
	TimerLen   time.Duration
}

type option func(c *config)

func WithActivityFile(path string) option {
	return func(c *config) {
		c.LogFile = path
	}
}

func WithLogFile(path string) option {
	return func(c *config) {
		c.LspLogFile = path
	}
}

func WithDebug(debug bool) option {
	return func(c *config) {
		c.Debug = debug
	}
}

func WithTimeoutLen(timeout time.Duration) option {
	return func(c *config) {
		c.TimeoutLen = timeout
	}
}

func WithTimerLen(timer time.Duration) option {
	return func(c *config) {
		c.TimerLen = timer
	}
}

func testFunc() {
	return
}

func Setup(opts ...option) error {
	cfg = new(config)

	for _, o := range opts {
		o(cfg)
	}

	if cfg.LogFile == "" {
		return fmt.Errorf("pendulum activity log file option not set")
	}

	if _, err := os.Stat(cfg.LogFile); err != nil && os.IsNotExist(err) {
		f, err := os.Create(cfg.LogFile)
		if err != nil {
			return fmt.Errorf("pendulum activity log does not exist and could not be created")
		}
		if _, err := f.Write([]byte("active,branch,cwd,file,filetype,project,time\n")); err != nil {
			return fmt.Errorf("failed to write pendulum header")
		}
	}

	return nil
}

func Config() *config {
	return cfg
}
