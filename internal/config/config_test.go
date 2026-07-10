package config

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSetupCreatesParentDirectoryAndCSVHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "activity.csv")

	if err := Setup(WithActivityFile(path)); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open activity log: %v", err)
	}
	defer f.Close()

	header, err := csv.NewReader(f).Read()
	if err != nil {
		t.Fatalf("read activity log header: %v", err)
	}
	want := []string{"active", "branch", "cwd", "file", "filetype", "project", "time"}
	if !reflect.DeepEqual(header, want) {
		t.Errorf("header = %q, want %q", header, want)
	}
}
