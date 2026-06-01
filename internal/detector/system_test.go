package detector

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestSystemDetector_Name(t *testing.T) {
	d := &SystemDetector{}
	if d.Name() != "System" {
		t.Errorf("expected name 'System', got '%s'", d.Name())
	}
}

func TestSystemDetector_Detect(t *testing.T) {
	d := &SystemDetector{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := d.Detect(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Section != "System" {
		t.Errorf("expected section 'System', got '%s'", result.Section)
	}

	// Should have at least OS and Architecture
	if len(result.Items) < 2 {
		t.Errorf("expected at least 2 items, got %d", len(result.Items))
	}

	// Verify OS is correct
	var foundOS, foundArch bool
	for _, item := range result.Items {
		if item.Key == "OS" {
			if item.Value != runtime.GOOS {
				t.Errorf("expected OS '%s', got '%s'", runtime.GOOS, item.Value)
			}
			foundOS = true
		}
		if item.Key == "Architecture" {
			if item.Value != runtime.GOARCH {
				t.Errorf("expected Architecture '%s', got '%s'", runtime.GOARCH, item.Value)
			}
			foundArch = true
		}
	}

	if !foundOS {
		t.Error("OS item not found in result")
	}
	if !foundArch {
		t.Error("Architecture item not found in result")
	}
}
