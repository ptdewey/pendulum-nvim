package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ptdewey/pendulum-server/internal/config"
	"github.com/tliron/glsp"
)

type activityData struct {
	Active   bool   `json:"active"`
	Branch   string `json:"branch"`
	Cwd      string `json:"cwd"`
	File     string `json:"file"`
	Filetype string `json:"filetype"`
	Project  string `json:"project"`
	Time     string `json:"time"`
}

// TODO: output should just be error
func LogActivity(ctx *glsp.Context, args []any) (bool, error) {
	log.Printf("executing command 'pendulum.logActivity' with args %v (%T)", args, args)

	if len(args) == 0 {
		return false, fmt.Errorf("no arguments provided")
	}

	// Convert to JSON and back to populate struct fields automatically
	jsonBytes, err := json.Marshal(args[0])
	if err != nil {
		return false, fmt.Errorf("invalid args: %w", err)
	}

	var ad activityData
	if err := json.Unmarshal(jsonBytes, &ad); err != nil {
		return false, fmt.Errorf("failed to parse activity data: %w", err)
	}

	am := GetActivityManager()
	if am != nil {
		// Update activity manager with current data
		am.SetCurrentData(&ad)
		am.UpdateActivity()

		// Immediately log this activity data
		ad.Project = getGitProject(ad.Cwd)
		ad.Branch = getGitBranch(ad.Cwd)
		if err := writeActivityToCSV(&ad); err != nil {
			log.Printf("Failed to write activity data: %v", err)
		}
	} else {
		// Fallback to direct logging if manager not available
		ad.Project = getGitProject(ad.Cwd)
		ad.Branch = getGitBranch(ad.Cwd)

		row := ad.toCSV()

		f, err := os.OpenFile(config.Config().LogFile, os.O_WRONLY|os.O_APPEND, 0664)
		if err != nil {
			return false, err
		}
		defer f.Close()

		if _, err := f.Write([]byte(row)); err != nil {
			return false, err
		}
	}

	return true, nil
}

func getGitBranch(cwd string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = cwd

	output, err := cmd.Output()
	if err != nil {
		return "unknown_branch"
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" || strings.HasPrefix(branch, "fatal:") {
		return "unknown_branch"
	}

	return branch
}

func getGitProject(cwd string) string {
	cmd := exec.Command("git", "config", "--local", "remote.origin.url")
	cmd.Dir = cwd

	output, err := cmd.Output()
	if err != nil {
		return "unknown_project"
	}

	url := strings.TrimSpace(string(output))

	re := regexp.MustCompile(`.*/([^.]+)\.git$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}

	return "unknown_project"
}

func (ad *activityData) toCSV() string {
	return fmt.Sprintf("%t,%s,%s,%s,%s,%s,%s\n",
		ad.Active,
		ad.Branch,
		ad.Cwd,
		ad.File,
		ad.Filetype,
		ad.Project,
		ad.Time,
	)
}

type sessionConfig struct {
	TimeoutLen int `json:"timeout_len"`
	TimerLen   int `json:"timer_len"`
}

func ActivityPing(ctx *glsp.Context, args []any) (bool, error) {
	log.Printf("executing command 'pendulum.activityPing'")

	am := GetActivityManager()
	if am == nil {
		return false, fmt.Errorf("activity manager not initialized")
	}

	am.UpdateActivity()
	return true, nil
}

func StartSession(ctx *glsp.Context, args []any) (bool, error) {
	log.Printf("executing command 'pendulum.startSession' with args %v", args)

	// Default values
	timeoutLen := 5 * time.Second
	timerLen := 1 * time.Second

	// Parse config if provided
	if len(args) > 0 {
		jsonBytes, err := json.Marshal(args[0])
		if err == nil {
			var cfg sessionConfig
			if json.Unmarshal(jsonBytes, &cfg) == nil {
				if cfg.TimeoutLen > 0 {
					timeoutLen = time.Duration(cfg.TimeoutLen) * time.Second
				}
				if cfg.TimerLen > 0 {
					timerLen = time.Duration(cfg.TimerLen) * time.Second
				}
			}
		}
	}

	InitializeActivityManager(ctx, timeoutLen, timerLen)
	log.Printf("Activity session started with timeout: %v, timer: %v", timeoutLen, timerLen)
	return true, nil
}

func EndSession(ctx *glsp.Context, args []any) (bool, error) {
	log.Printf("executing command 'pendulum.endSession'")

	am := GetActivityManager()
	if am != nil {
		am.Stop()
	}

	log.Println("Activity session ended")
	return true, nil
}
