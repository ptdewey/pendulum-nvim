package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tliron/glsp"
)

// TODO:

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

	row := ad.toCSV()

	// TODO: write to csv
	return false, nil
}

func (ad *activityData) toCSV() string {
	return fmt.Sprintf("%t,%s,%s,%s,%s,%s,%s",
		ad.Active,
		ad.Branch,
		ad.Cwd,
		ad.File,
		ad.Filetype,
		ad.Project,
		ad.Time,
	)
}
