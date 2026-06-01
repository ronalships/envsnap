package diff

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
)

// Options for the diff command.
type Options struct {
	File1 string
	File2 string
}

// Run executes the diff command.
// Returns true if files are identical, false otherwise.
func Run(opts Options) (bool, error) {
	content1, err := os.ReadFile(opts.File1)
	if err != nil {
		return false, fmt.Errorf("cannot read %s: %w", opts.File1, err)
	}

	content2, err := os.ReadFile(opts.File2)
	if err != nil {
		return false, fmt.Errorf("cannot read %s: %w", opts.File2, err)
	}

	text1 := string(content1)
	text2 := string(content2)

	// Check if identical
	if text1 == text2 {
		color.Green("✓ Snapshots are identical")
		return true, nil
	}

	// Generate unified diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(text1),
		B:        difflib.SplitLines(text2),
		FromFile: opts.File1,
		ToFile:   opts.File2,
		Context:  3,
	}

	result, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return false, fmt.Errorf("cannot generate diff: %w", err)
	}

	// Print colored diff
	printColoredDiff(result)

	return false, nil
}

func printColoredDiff(diff string) {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	bold := color.New(color.Bold)

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			bold.Println(line)
		case strings.HasPrefix(line, "@@"):
			cyan.Println(line)
		case strings.HasPrefix(line, "-"):
			red.Print("✗ ")
			red.Println(line[1:])
		case strings.HasPrefix(line, "+"):
			green.Print("✓ ")
			green.Println(line[1:])
		default:
			fmt.Println(line)
		}
	}
}
