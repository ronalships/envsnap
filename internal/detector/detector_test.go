package detector

import (
	"context"
	"testing"
	"time"
)

func TestRunAllWithTimeout(t *testing.T) {
	// Verify that RunAll respects context and timeouts
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := RunAll(ctx)

	// Should have results from all registered detectors
	if len(results) == 0 {
		t.Error("expected at least one result from RunAll")
	}

	// Each result should have a section name
	for _, r := range results {
		if r.Section == "" {
			t.Error("result has empty section name")
		}
	}
}

func TestDefaultTimeout(t *testing.T) {
	if DefaultTimeout != 2*time.Second {
		t.Errorf("expected default timeout of 2s, got %v", DefaultTimeout)
	}
}
