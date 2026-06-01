package detector

import (
	"context"
	"os"
	"sort"
	"strings"
)

func init() {
	Register(&EnvVarsDetector{})
}

// EnvVarsDetector captures environment variable names (values redacted by default).
type EnvVarsDetector struct {
	ShowValues bool
}

func (d *EnvVarsDetector) Name() string {
	return "Environment Variables"
}

func (d *EnvVarsDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "Environment Variables"}

	env := os.Environ()
	sort.Strings(env) // Stable ordering

	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// Skip some uninteresting variables
		if shouldSkipEnvVar(key) {
			continue
		}

		if d.ShowValues {
			if isSensitiveKey(key) {
				value = "[REDACTED]"
			}
		} else {
			// Show only the key
			if value != "" {
				value = "[set]"
			} else {
				value = "[empty]"
			}
		}

		result.Items = append(result.Items, Item{Key: key, Value: value})
	}

	return result, nil
}

// shouldSkipEnvVar returns true for env vars that are not useful in snapshots.
func shouldSkipEnvVar(key string) bool {
	// Skip internal/system variables that don't help with debugging
	skipPrefixes := []string{
		"_",          // Internal
		"LESS",       // Pager config
		"LS_COLORS",  // Terminal colors
		"LSCOLORS",   // Terminal colors
		"PS1",        // Prompt
		"PS2",        // Prompt
		"PROMPT_",    // Prompt
		"COMP_",      // Completion
		"BASH_",      // Bash internals
		"ZSH_",       // Zsh internals
		"HISTFILE",   // History
		"HISTSIZE",   // History
		"SAVEHIST",   // History
		"OLDPWD",     // Previous directory
		"SHLVL",      // Shell level
		"WINDOWID",   // X11
		"DISPLAY",    // X11 (usually not useful)
		"COLORTERM",  // Terminal
		"TERM_",      // Terminal internals
		"VTE_",       // Terminal internals
		"KONSOLE_",   // Terminal internals
		"ITERM_",     // iTerm internals
		"SECURITYSESSIONID",
		"LaunchInstanceID",
		"XPC_",
		"Apple_",
		"__CF",
		"__fish",
		"fish_",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Skip exact matches
	skipExact := map[string]bool{
		"PWD":           true,
		"TMPDIR":        true,
		"TEMP":          true,
		"TMP":           true,
		"LOGNAME":       true,
		"_":             true,
		"COLUMNS":       true,
		"LINES":         true,
		"TERM_PROGRAM":  true, // Captured in System section
		"TERM":          true, // Captured in System section
		"SHELL":         true, // Captured in System section
	}

	return skipExact[key]
}

// isSensitiveKey returns true if the key likely contains sensitive data.
func isSensitiveKey(key string) bool {
	key = strings.ToUpper(key)
	sensitivePatterns := []string{
		"KEY",
		"SECRET",
		"TOKEN",
		"PASSWORD",
		"PASSWD",
		"CREDENTIAL",
		"AUTH",
		"PRIVATE",
		"ACCESS",
		"API_KEY",
		"APIKEY",
		"BEARER",
		"JWT",
		"SESSION",
		"COOKIE",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(key, pattern) {
			return true
		}
	}

	return false
}

// SetShowValues is a helper to configure the detector.
// Used by the capture command when --values flag is set.
func SetEnvVarsShowValues(show bool) {
	for _, d := range Registry {
		if evd, ok := d.(*EnvVarsDetector); ok {
			evd.ShowValues = show
		}
	}
}
