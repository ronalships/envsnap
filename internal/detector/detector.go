package detector

import (
	"context"
	"time"
)

// DefaultTimeout is the maximum time a detector should take.
const DefaultTimeout = 2 * time.Second

// Result holds the output of a detector.
type Result struct {
	// Section is the markdown section name (e.g., "System", "CLI Tools").
	Section string

	// Items contains the detected key-value pairs or lines.
	Items []Item

	// Error is set if detection failed (partial results may still exist).
	Error error
}

// Item represents a single detected item.
type Item struct {
	Key   string
	Value string
}

// Detector is the interface all detectors must implement.
type Detector interface {
	// Name returns a human-readable name for the detector.
	Name() string

	// Detect runs the detection and returns results.
	// The context should be used for timeout/cancellation.
	Detect(ctx context.Context) (Result, error)
}

// Registry holds all registered detectors in order.
var Registry []Detector

// Register adds a detector to the registry.
func Register(d Detector) {
	Registry = append(Registry, d)
}

// RunAll executes all registered detectors and returns their results.
// Each detector runs with its own timeout context.
func RunAll(ctx context.Context) []Result {
	results := make([]Result, 0, len(Registry))

	for _, d := range Registry {
		// Create a timeout context for this detector
		dctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
		result, err := d.Detect(dctx)
		cancel()

		if err != nil {
			result.Error = err
		}
		results = append(results, result)
	}

	return results
}
