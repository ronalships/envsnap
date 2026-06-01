package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ronalships/envsnap/internal/capture"
	"github.com/ronalships/envsnap/internal/diff"
	"github.com/ronalships/envsnap/internal/share"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "capture":
		runCapture(os.Args[2:])
	case "diff":
		runDiff(os.Args[2:])
	case "share":
		runShare(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("envsnap %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`envsnap - Stop saying "works on my machine."

Usage:
  envsnap <command> [options]

Commands:
  capture    Snapshot the local environment to markdown
  diff       Compare two snapshot files
  share      Upload a snapshot to GitHub Gist

Options:
  --version  Print version information
  --help     Show this help message

Examples:
  envsnap capture                  # Print snapshot to stdout
  envsnap capture -o snapshot.md   # Write snapshot to file
  envsnap diff old.md new.md       # Compare two snapshots
  envsnap share snapshot.md        # Upload snapshot to Gist

Run 'envsnap <command> --help' for more information on a command.`)
}

func runCapture(args []string) {
	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	output := fs.String("o", "", "Output file (default: stdout)")
	showValues := fs.Bool("values", false, "Show environment variable values (caution: may expose secrets)")

	fs.Usage = func() {
		fmt.Println(`Usage: envsnap capture [options]

Capture a snapshot of the local development environment.

Options:`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	opts := capture.Options{
		Output:     *output,
		ShowValues: *showValues,
	}

	if err := capture.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Println(`Usage: envsnap diff <file1> <file2>

Compare two snapshot files and highlight differences.

Exit codes:
  0  Snapshots are identical
  1  Snapshots differ`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	opts := diff.Options{
		File1: fs.Arg(0),
		File2: fs.Arg(1),
	}

	identical, err := diff.Run(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !identical {
		os.Exit(1)
	}
}

func runShare(args []string) {
	fs := flag.NewFlagSet("share", flag.ExitOnError)
	public := fs.Bool("public", false, "Create a public gist (default: secret)")
	description := fs.String("d", "envsnap snapshot", "Gist description")

	fs.Usage = func() {
		fmt.Println(`Usage: envsnap share [options] [file]

Upload a snapshot to GitHub Gist. Requires 'gh' CLI to be authenticated.

If no file is provided, runs capture first and shares the result.

Options:`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	opts := share.Options{
		File:        fs.Arg(0),
		Public:      *public,
		Description: *description,
	}

	if err := share.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
