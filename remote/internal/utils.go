package internal

import (
	"fmt"
	"os"
    "path/filepath"
	"strings"
	"time"
)

// calculate the duration between two string timestamps (curr - prev)
func calcDuration(curr string, prev string) (time.Duration, error) {
    layout := "2006-01-02 15:04:05"

    curr_t, err := time.Parse(layout, curr)
    if err != nil {
        return time.Duration(0), err
    }

    prev_t, err := time.Parse(layout, prev)
    if err != nil {
        return time.Duration(0), err
    }

    return curr_t.Sub(prev_t), nil
}

// reformat time.Duration values into a more readable format
func formatDuration(d time.Duration) string {
    if d >= 24 * time.Hour {
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
        path = strings.Join(parts[len(parts) - 7:], string(filepath.Separator))
    }

    return path
}
