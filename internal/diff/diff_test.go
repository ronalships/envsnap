package diff

import (
	"path/filepath"
	"testing"
)

func TestParseSnapshot(t *testing.T) {
	snap, err := parseSnapshot(filepath.Join("testdata", "snapshot1.md"))
	if err != nil {
		t.Fatalf("failed to parse snapshot: %v", err)
	}

	// Check sections exist
	expectedSections := []string{"System", "CLI Tools", "AWS Configuration", "Git Configuration"}
	for _, name := range expectedSections {
		if snap.Sections[name] == nil {
			t.Errorf("expected section %q not found", name)
		}
	}

	// Check System section fields
	system := snap.Sections["System"]
	if system == nil {
		t.Fatal("System section not found")
	}

	if system.Fields["OS"] != "darwin" {
		t.Errorf("OS = %q, expected %q", system.Fields["OS"], "darwin")
	}
	if system.Fields["Architecture"] != "arm64" {
		t.Errorf("Architecture = %q, expected %q", system.Fields["Architecture"], "arm64")
	}

	// Check CLI Tools section
	tools := snap.Sections["CLI Tools"]
	if tools == nil {
		t.Fatal("CLI Tools section not found")
	}

	if tools.Fields["git"] != "2.39.3" {
		t.Errorf("git = %q, expected %q", tools.Fields["git"], "2.39.3")
	}
	if tools.Fields["yarn"] != "1.22.22" {
		t.Errorf("yarn = %q, expected %q", tools.Fields["yarn"], "1.22.22")
	}
}

func TestRunIdentical(t *testing.T) {
	opts := Options{
		File1: filepath.Join("testdata", "snapshot1.md"),
		File2: filepath.Join("testdata", "snapshot1_copy.md"),
	}

	identical, err := Run(opts)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !identical {
		t.Error("expected identical snapshots to return true")
	}
}

func TestRunDifferent(t *testing.T) {
	opts := Options{
		File1: filepath.Join("testdata", "snapshot1.md"),
		File2: filepath.Join("testdata", "snapshot2.md"),
	}

	identical, err := Run(opts)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if identical {
		t.Error("expected different snapshots to return false")
	}
}

func TestFieldDiffDetection(t *testing.T) {
	snap1, _ := parseSnapshot(filepath.Join("testdata", "snapshot1.md"))
	snap2, _ := parseSnapshot(filepath.Join("testdata", "snapshot2.md"))

	// Check that differences are detected in System section
	sys1 := snap1.Sections["System"]
	sys2 := snap2.Sections["System"]

	if sys1.Fields["OS"] == sys2.Fields["OS"] {
		t.Error("expected OS to be different")
	}
	if sys1.Fields["Architecture"] == sys2.Fields["Architecture"] {
		t.Error("expected Architecture to be different")
	}

	// Check added field (kubectl in snapshot2, not in snapshot1)
	tools1 := snap1.Sections["CLI Tools"]
	tools2 := snap2.Sections["CLI Tools"]

	if _, has := tools1.Fields["kubectl"]; has {
		t.Error("kubectl should not exist in snapshot1")
	}
	if _, has := tools2.Fields["kubectl"]; !has {
		t.Error("kubectl should exist in snapshot2")
	}

	// Check removed field (yarn in snapshot1, not in snapshot2)
	if _, has := tools1.Fields["yarn"]; !has {
		t.Error("yarn should exist in snapshot1")
	}
	if _, has := tools2.Fields["yarn"]; has {
		t.Error("yarn should not exist in snapshot2")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, got, tt.expected)
		}
	}
}
