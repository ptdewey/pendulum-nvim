package handlers

import (
	"log"

	"github.com/tliron/glsp"
)

// TODO: rename to "summary" report
func GenerateMetricsReport(ctx *glsp.Context, args []any) (string, error) {
	log.Printf("executing command 'pendulum.generateMetricsReport' with args %v (%T)", args, args)
	// TODO:
	return "", nil
}

// TODO: find a better name for this one
func GenerateHourlyReport(ctx *glsp.Context, args []any) (string, error) {
	log.Printf("executing command 'pendulum.generateHourlyReport' with args %v (%T)", args, args)

	// ctx.Notify()

	// TODO:
	return "", nil
}
