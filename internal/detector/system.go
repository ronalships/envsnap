package detector

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func init() {
	Register(&SystemDetector{})
}

// SystemDetector captures OS, kernel, architecture, shell, and terminal info.
type SystemDetector struct{}

func (d *SystemDetector) Name() string {
	return "System"
}

func (d *SystemDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "System"}

	// OS and architecture
	result.Items = append(result.Items, Item{Key: "OS", Value: runtime.GOOS})
	result.Items = append(result.Items, Item{Key: "Architecture", Value: runtime.GOARCH})

	// Kernel version
	if kernel := getKernelVersion(ctx); kernel != "" {
		result.Items = append(result.Items, Item{Key: "Kernel", Value: kernel})
	}

	// Shell
	if shell := os.Getenv("SHELL"); shell != "" {
		result.Items = append(result.Items, Item{Key: "Shell", Value: shell})
	}

	// Terminal
	if term := os.Getenv("TERM"); term != "" {
		result.Items = append(result.Items, Item{Key: "Terminal", Value: term})
	}

	// Terminal program (if available)
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		version := os.Getenv("TERM_PROGRAM_VERSION")
		if version != "" {
			termProgram += " " + version
		}
		result.Items = append(result.Items, Item{Key: "Terminal Program", Value: termProgram})
	}

	return result, nil
}

func getKernelVersion(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "uname", "-r")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
