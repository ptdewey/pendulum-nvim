package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ptdewey/pendulum-server/internal/config"
	"github.com/tliron/glsp"
)

type ActivityManager struct {
	mu             sync.RWMutex
	lastActiveTime time.Time
	activeFlag     bool
	timeout        time.Duration
	interval       time.Duration
	cancel         context.CancelFunc
	ctx            *glsp.Context
	currentData    *activityData
}

var (
	activityManager *ActivityManager
	managerMu       sync.Mutex
	activityLogMu   sync.Mutex
)

func GetActivityManager() *ActivityManager {
	managerMu.Lock()
	defer managerMu.Unlock()
	return activityManager
}

func InitializeActivityManager(ctx *glsp.Context, timeout, interval time.Duration) *ActivityManager {
	managerMu.Lock()
	defer managerMu.Unlock()

	if activityManager != nil {
		activityManager.Stop()
	}

	activityManager = &ActivityManager{
		lastActiveTime: time.Now(),
		activeFlag:     true,
		timeout:        timeout,
		interval:       interval,
		ctx:            ctx,
	}

	activityManager.start()
	return activityManager
}

func (am *ActivityManager) start() {
	ctx, cancel := context.WithCancel(context.Background())
	am.cancel = cancel

	go func() {
		ticker := time.NewTicker(am.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				am.checkActiveStatus()
			}
		}
	}()

	log.Printf("Activity manager started with timeout: %v, interval: %v", am.timeout, am.interval)
}

func (am *ActivityManager) Stop() {
	if am.cancel != nil {
		am.cancel()
	}
}

func (am *ActivityManager) UpdateActivity() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.lastActiveTime = time.Now()
}

func (am *ActivityManager) SetCurrentData(data *activityData) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.currentData = data
}

func (am *ActivityManager) checkActiveStatus() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	isActive := now.Sub(am.lastActiveTime) < am.timeout

	// Log inactive period when transitioning from active to inactive
	if !isActive && am.activeFlag {
		am.activeFlag = false
		if am.currentData != nil {
			// Log the last active time as an active entry
			data := *am.currentData
			data.Active = true
			data.Time = am.lastActiveTime.UTC().Format(time.RFC3339)
			am.logActivityData(&data)
		}
	} else if isActive && !am.activeFlag {
		am.activeFlag = true
	}

	// Always log current state if we have data
	if am.currentData != nil {
		data := *am.currentData
		data.Active = isActive
		data.Time = now.UTC().Format(time.RFC3339)
		am.logActivityData(&data)
	}
}

func (am *ActivityManager) logActivityData(data *activityData) {
	if data.File == "" {
		return
	}

	// Data received from LogActivity is already enriched. Keep this fallback
	// for callers that construct activityData directly.
	if data.Project == "" {
		data.Project = getGitProject(data.Cwd)
	}
	if data.Branch == "" {
		data.Branch = getGitBranch(data.Cwd)
	}

	if err := writeActivityToCSV(data); err != nil {
		log.Printf("Failed to write activity data: %v", err)
	}
}

func writeActivityToCSV(data *activityData) (err error) {
	activityLogMu.Lock()
	defer activityLogMu.Unlock()

	f, err := os.OpenFile(config.Config().LogFile, os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	w := csv.NewWriter(f)
	if err := w.Write([]string{
		fmt.Sprintf("%t", data.Active),
		data.Branch,
		data.Cwd,
		data.File,
		data.Filetype,
		data.Project,
		data.Time,
	}); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
