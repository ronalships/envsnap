package capture

import (
	"context"
	"io"
	"os"

	"github.com/ronalships/envsnap/internal/detector"
	"github.com/ronalships/envsnap/internal/output"
)

// Options for the capture command.
type Options struct {
	Output     string // Output file path (empty for stdout)
	ShowValues bool   // Show env var values
}

// Run executes the capture command.
func Run(opts Options) error {
	// Configure detectors based on options
	detector.SetEnvVarsShowValues(opts.ShowValues)

	// Run all detectors
	ctx := context.Background()
	results := detector.RunAll(ctx)

	// Determine output destination
	var w io.Writer
	if opts.Output == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(opts.Output)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	// Render to markdown
	md := output.NewMarkdown(w)
	return md.Render(results)
}

// CaptureToWriter runs capture and writes to the provided writer.
// Used by the share command.
func CaptureToWriter(w io.Writer, showValues bool) error {
	detector.SetEnvVarsShowValues(showValues)
	ctx := context.Background()
	results := detector.RunAll(ctx)
	md := output.NewMarkdown(w)
	return md.Render(results)
}
