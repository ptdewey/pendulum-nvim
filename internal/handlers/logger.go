package handlers

import (
	"fmt"
	"log"

	"github.com/tliron/glsp"
)

// TODO:

type activityData struct {
	active   bool
	branch   string
	cwd      string
	file     string
	filetype string
	project  string
	time     string
}

// TODO: output should just be error
func LogActivity(ctx *glsp.Context, args []any) (bool, error) {
	// TODO: log with ctx? (allow different log levels)
	log.Printf("executing command 'pendulum.logActivity' with args %v (%T)", args, args)

	m, ok := args[0].(map[string]any)
	if !ok {
		// TODO: log error?
		return false, fmt.Errorf("invalid args")
	}

	ad := activityData{}

	if active, exists := m["active"].(bool); exists {
		ad.active = active
	}
	if branch, exists := m["branch"].(string); exists {
		ad.branch = branch
	}
	if cwd, err := m["cwd"].(string); err {
		ad.cwd = cwd
	}
	if file, exists := m["file"].(string); exists {
		ad.file = file
	}
	if filetype, exists := m["filetype"].(string); exists {
		ad.filetype = filetype
	}
	if project, exists := m["project"].(string); exists {
		ad.project = project
	}
	if time, exists := m["time"].(string); exists {
		ad.time = time
	}

	// TODO: write to csv
	// - create if not exist

	return false, nil
}
