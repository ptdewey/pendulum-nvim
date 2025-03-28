package prettify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// reformat time.Duration values into a more readable format
func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := float32(d) / (24 * float32(time.Hour))
		return fmt.Sprintf("%.2fd", days)
	} else if d >= time.Hour {
		hours := float32(d) / float32(time.Hour)
		return fmt.Sprintf("%.2fh", hours)
	} else if d >= time.Minute {
		minutes := float32(d) / float32(time.Minute)
		return fmt.Sprintf("%.2fm", minutes)
	} else {
		seconds := float32(d) / float32(time.Second)
		return fmt.Sprintf("%.2fs", seconds)
	}
}

// Truncate long path strings and replace /home/user with ~/
func truncatePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "<No Name>"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	if strings.HasPrefix(path, home) {
		rpath, err := filepath.Rel(home, path)
		if err == nil {
			path = "~" + string(filepath.Separator) + rpath
		}
	}

	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) > 7 {
		path = strings.Join(parts[len(parts)-7:], string(filepath.Separator))
	}

	return path
}
