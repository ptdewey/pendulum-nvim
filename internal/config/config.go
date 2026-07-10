package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
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

	if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0755); err != nil {
		return fmt.Errorf("create pendulum activity log directory: %w", err)
	}

	if _, err := os.Stat(cfg.LogFile); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat pendulum activity log: %w", err)
		}

		if err := createActivityLog(cfg.LogFile); err != nil {
			return err
		}
	}

	return nil
}

func createActivityLog(path string) (err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("create pendulum activity log: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close pendulum activity log: %w", closeErr)
		}
	}()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"active", "branch", "cwd", "file", "filetype", "project", "time"}); err != nil {
		return fmt.Errorf("write pendulum header: %w", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("write pendulum header: %w", err)
	}

	return nil
}

func Config() *config {
	return cfg
}
