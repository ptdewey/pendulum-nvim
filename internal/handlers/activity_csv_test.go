package handlers

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ptdewey/pendulum-server/internal/config"
)

func TestWriteActivityToCSVEscapesFieldsAndSerializesAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "activity.csv")
	if err := config.Setup(config.WithActivityFile(path)); err != nil {
		t.Fatalf("setup config: %v", err)
	}

	const writes = 64
	var wg sync.WaitGroup
	errCh := make(chan error, writes)
	for i := range writes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- writeActivityToCSV(&activityData{
				Active:   true,
				Branch:   fmt.Sprintf("feature,%d", i),
				Cwd:      "/tmp/project",
				File:     fmt.Sprintf("report %d \"draft\".go", i),
				Filetype: "go",
				Project:  "pendulum",
				Time:     "2026-07-10 12:00:00",
			})
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("writeActivityToCSV() error = %v", err)
		}
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open activity log: %v", err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read activity log: %v", err)
	}

	if got, want := len(records), writes+1; got != want {
		t.Fatalf("record count = %d, want %d", got, want)
	}
	seen := make(map[string]bool, writes)
	for _, record := range records[1:] {
		if got, want := len(record), 7; got != want {
			t.Fatalf("field count = %d, want %d: %q", got, want, record)
		}
		seen[record[1]] = true
	}
	for i := range writes {
		branch := fmt.Sprintf("feature,%d", i)
		if !seen[branch] {
			t.Errorf("missing CSV record for branch %q", branch)
		}
	}
}
