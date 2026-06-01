package detector

import (
	"context"
	"os/exec"
	"strings"
)

func init() {
	Register(&GitDetector{})
}

// GitDetector captures global Git configuration (user.name and user.email only).
// Email local part is redacted by default to protect privacy.
type GitDetector struct{}

func (d *GitDetector) Name() string {
	return "Git"
}

func (d *GitDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "Git Configuration"}

	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return result, nil
	}

	// Git version
	cmd := exec.CommandContext(ctx, "git", "--version")
	out, err := cmd.Output()
	if err == nil {
		result.Items = append(result.Items, Item{Key: "Version", Value: strings.TrimSpace(string(out))})
	}

	// Global user.name - shown as-is (intentional public identity)
	cmd = exec.CommandContext(ctx, "git", "config", "--global", "user.name")
	out, err = cmd.Output()
	if err == nil {
		name := strings.TrimSpace(string(out))
		if name != "" {
			result.Items = append(result.Items, Item{Key: "user.name", Value: name})
		}
	}

	// Global user.email - redact local part by default
	cmd = exec.CommandContext(ctx, "git", "config", "--global", "user.email")
	out, err = cmd.Output()
	if err == nil {
		email := strings.TrimSpace(string(out))
		if email != "" {
			// Redact local part: "user@example.com" -> "***@example.com"
			redactedEmail := RedactEmailLocalPart(email)
			result.Items = append(result.Items, Item{Key: "user.email", Value: redactedEmail})
		}
	}

	// Default branch (if configured)
	cmd = exec.CommandContext(ctx, "git", "config", "--global", "init.defaultBranch")
	out, err = cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			result.Items = append(result.Items, Item{Key: "init.defaultBranch", Value: branch})
		}
	}

	return result, nil
}
