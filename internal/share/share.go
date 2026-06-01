package share

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/ronalships/envsnap/internal/capture"
)

// Options for the share command.
type Options struct {
	File        string // File to share (empty to capture first)
	Public      bool   // Create public gist
	Description string // Gist description
}

// Run executes the share command.
func Run(opts Options) error {
	var content []byte
	var err error
	var filename string

	if opts.File == "" {
		// Capture first
		var buf bytes.Buffer
		if err := capture.CaptureToWriter(&buf, false); err != nil {
			return fmt.Errorf("capture failed: %w", err)
		}
		content = buf.Bytes()
		filename = "envsnap.md"
	} else {
		content, err = os.ReadFile(opts.File)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", opts.File, err)
		}
		filename = opts.File
		// Use just the base name for gist
		if idx := strings.LastIndex(filename, "/"); idx != -1 {
			filename = filename[idx+1:]
		}
	}

	// Try using gh CLI first (preferred method)
	url, err := shareViaGH(content, filename, opts.Description, opts.Public)
	if err != nil {
		// Fall back to direct API if gh is not available
		url, err = shareViaAPI(content, filename, opts.Description, opts.Public)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Snapshot shared: %s\n", url)
	return nil
}

// shareViaGH uses the gh CLI to create a gist.
func shareViaGH(content []byte, filename, description string, public bool) (string, error) {
	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return "", fmt.Errorf("gh CLI not found")
	}

	// Check if gh is authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh not authenticated: run 'gh auth login' first")
	}

	// Create temp file for gist content
	tmpfile, err := os.CreateTemp("", "envsnap-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		return "", err
	}
	tmpfile.Close()

	// Build gh gist create command
	args := []string{"gist", "create", tmpfile.Name(), "--filename", filename}
	if description != "" {
		args = append(args, "--desc", description)
	}
	if public {
		args = append(args, "--public")
	}

	cmd = exec.Command("gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh gist create failed: %s", string(out))
	}

	// gh outputs the gist URL
	url := strings.TrimSpace(string(out))
	return url, nil
}

// shareViaAPI uses the GitHub API directly.
// Requires GITHUB_TOKEN environment variable.
func shareViaAPI(content []byte, filename, description string, public bool) (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", fmt.Errorf("no authentication available: install 'gh' CLI and run 'gh auth login', or set GITHUB_TOKEN")
	}

	gist := struct {
		Description string                       `json:"description"`
		Public      bool                         `json:"public"`
		Files       map[string]map[string]string `json:"files"`
	}{
		Description: description,
		Public:      public,
		Files: map[string]map[string]string{
			filename: {"content": string(content)},
		},
	}

	body, err := json.Marshal(gist)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.HTMLURL, nil
}
