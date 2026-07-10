package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

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

func LogActivity(ctx *glsp.Context, args []any) (bool, error) {
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
	// Resolve repository data once when the activity changes. The periodic
	// ticker reuses this populated currentData instead of running git twice for
	// every activity sample, which otherwise delays queued report commands.
	ad.Project = getGitProject(ad.Cwd)
	ad.Branch = getGitBranch(ad.Cwd)

	am := GetActivityManager()
	if am != nil {
		// Update activity manager with current data
		am.SetCurrentData(&ad)
		am.UpdateActivity()

		// Immediately log this activity data
		if err := writeActivityToCSV(&ad); err != nil {
			log.Printf("Failed to write activity data: %v", err)
		}
	} else {
		// Fallback to direct logging if manager not available
		if err := writeActivityToCSV(&ad); err != nil {
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

func ActivityPing(ctx *glsp.Context, args []any) (bool, error) {
	am := GetActivityManager()
	if am == nil {
		return false, fmt.Errorf("activity manager not initialized")
	}

	am.UpdateActivity()
	return true, nil
}

func StartSession(ctx *glsp.Context, args []any) (bool, error) {
	log.Printf("executing command 'pendulum.startSession' with args %v", args)

	// Start with CLI config defaults
	cfg := config.Config()
	timeoutLen := cfg.TimeoutLen
	timerLen := cfg.TimerLen

	// Stop existing manager if any
	if am := GetActivityManager(); am != nil {
		am.Stop()
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
