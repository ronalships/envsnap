package detector

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	Register(&VersionManagerDetector{})
}

// VersionManagerDetector detects version managers like nvm, pyenv, rbenv, asdf.
type VersionManagerDetector struct{}

func (d *VersionManagerDetector) Name() string {
	return "Version Managers"
}

func (d *VersionManagerDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "Version Managers"}

	home := os.Getenv("HOME")
	if home == "" {
		return result, nil
	}

	// nvm
	if nvmDir := os.Getenv("NVM_DIR"); nvmDir != "" {
		if _, err := os.Stat(nvmDir); err == nil {
			nodeVersion := detectNvmVersion(ctx)
			result.Items = append(result.Items, Item{Key: "nvm", Value: "installed" + nodeVersion})
		}
	} else if _, err := os.Stat(filepath.Join(home, ".nvm")); err == nil {
		nodeVersion := detectNvmVersion(ctx)
		result.Items = append(result.Items, Item{Key: "nvm", Value: "installed" + nodeVersion})
	}

	// pyenv
	if pyenvRoot := os.Getenv("PYENV_ROOT"); pyenvRoot != "" {
		if _, err := os.Stat(pyenvRoot); err == nil {
			pyVersion := detectPyenvVersion(ctx)
			result.Items = append(result.Items, Item{Key: "pyenv", Value: "installed" + pyVersion})
		}
	} else if _, err := os.Stat(filepath.Join(home, ".pyenv")); err == nil {
		pyVersion := detectPyenvVersion(ctx)
		result.Items = append(result.Items, Item{Key: "pyenv", Value: "installed" + pyVersion})
	}

	// rbenv
	if rbenvRoot := os.Getenv("RBENV_ROOT"); rbenvRoot != "" {
		if _, err := os.Stat(rbenvRoot); err == nil {
			rubyVersion := detectRbenvVersion(ctx)
			result.Items = append(result.Items, Item{Key: "rbenv", Value: "installed" + rubyVersion})
		}
	} else if _, err := os.Stat(filepath.Join(home, ".rbenv")); err == nil {
		rubyVersion := detectRbenvVersion(ctx)
		result.Items = append(result.Items, Item{Key: "rbenv", Value: "installed" + rubyVersion})
	}

	// asdf
	if asdfDir := os.Getenv("ASDF_DIR"); asdfDir != "" {
		if _, err := os.Stat(asdfDir); err == nil {
			asdfVersion := detectAsdfVersion(ctx)
			result.Items = append(result.Items, Item{Key: "asdf", Value: "installed" + asdfVersion})
		}
	} else if _, err := os.Stat(filepath.Join(home, ".asdf")); err == nil {
		asdfVersion := detectAsdfVersion(ctx)
		result.Items = append(result.Items, Item{Key: "asdf", Value: "installed" + asdfVersion})
	}

	// goenv
	if _, err := os.Stat(filepath.Join(home, ".goenv")); err == nil {
		result.Items = append(result.Items, Item{Key: "goenv", Value: "installed"})
	}

	// rustup
	if _, err := os.Stat(filepath.Join(home, ".rustup")); err == nil {
		rustupVersion := detectRustupVersion(ctx)
		result.Items = append(result.Items, Item{Key: "rustup", Value: "installed" + rustupVersion})
	}

	// sdkman
	if _, err := os.Stat(filepath.Join(home, ".sdkman")); err == nil {
		result.Items = append(result.Items, Item{Key: "sdkman", Value: "installed"})
	}

	return result, nil
}

func detectNvmVersion(ctx context.Context) string {
	// Try to get current node version via nvm
	cmd := exec.CommandContext(ctx, "bash", "-c", "source $NVM_DIR/nvm.sh && nvm current 2>/dev/null")
	out, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(out))
		if version != "" && version != "none" && version != "system" {
			return " (active: " + version + ")"
		}
	}
	return ""
}

func detectPyenvVersion(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "pyenv", "version-name")
	out, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(out))
		if version != "" && version != "system" {
			return " (active: " + version + ")"
		}
	}
	return ""
}

func detectRbenvVersion(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "rbenv", "version-name")
	out, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(out))
		if version != "" && version != "system" {
			return " (active: " + version + ")"
		}
	}
	return ""
}

func detectAsdfVersion(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "asdf", "current")
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 && len(lines) <= 5 {
			// Return a summary of active versions
			return " (" + strings.Join(lines, "; ") + ")"
		} else if len(lines) > 5 {
			return " (" + string(rune(len(lines))) + " tools active)"
		}
	}
	return ""
}

func detectRustupVersion(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "rustup", "show", "active-toolchain")
	out, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(out))
		if version != "" {
			// Just get the toolchain name
			parts := strings.Fields(version)
			if len(parts) > 0 {
				return " (active: " + parts[0] + ")"
			}
		}
	}
	return ""
}
