package detector

import (
	"context"
	"os"
	"path/filepath"
	"sort"
)

func init() {
	Register(&DotfilesDetector{})
}

// DotfilesDetector inventories which dotfiles exist in $HOME.
// Never reads file contents, only checks existence.
type DotfilesDetector struct{}

func (d *DotfilesDetector) Name() string {
	return "Dotfiles"
}

// commonDotfiles is the list of dotfiles to check for.
var commonDotfiles = []string{
	".bashrc",
	".bash_profile",
	".bash_login",
	".profile",
	".zshrc",
	".zshenv",
	".zprofile",
	".zlogin",
	".config/fish/config.fish",
	".vimrc",
	".config/nvim/init.vim",
	".config/nvim/init.lua",
	".emacs",
	".emacs.d/init.el",
	".gitconfig",
	".gitignore_global",
	".ssh/config",
	".tmux.conf",
	".screenrc",
	".inputrc",
	".curlrc",
	".wgetrc",
	".npmrc",
	".yarnrc",
	".editorconfig",
	".prettierrc",
	".eslintrc",
	".config/starship.toml",
	".tool-versions",
	".envrc",
	".direnvrc",
}

func (d *DotfilesDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "Dotfiles"}

	home := os.Getenv("HOME")
	if home == "" {
		return result, nil
	}

	var present []string
	var absent []string

	for _, dotfile := range commonDotfiles {
		fullPath := filepath.Join(home, dotfile)
		if _, err := os.Stat(fullPath); err == nil {
			present = append(present, dotfile)
		} else {
			absent = append(absent, dotfile)
		}
	}

	// Count SSH keys (not their names or contents)
	sshKeyCount := countSSHKeys(home)

	sort.Strings(present)

	if len(present) > 0 {
		result.Items = append(result.Items, Item{Key: "Present", Value: formatDotfileList(present)})
	}

	if sshKeyCount > 0 {
		result.Items = append(result.Items, Item{Key: "SSH Keys", Value: itoa(sshKeyCount) + " key(s) in ~/.ssh/"})
	}

	return result, nil
}

func countSSHKeys(home string) int {
	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Common private key patterns (we count private keys, not public)
		if name == "id_rsa" || name == "id_ed25519" || name == "id_ecdsa" || name == "id_dsa" {
			count++
		}
		// Also count any id_* files that don't end in .pub
		if len(name) > 3 && name[:3] == "id_" && !hasPublicKeySuffix(name) {
			// Avoid double counting
			if name != "id_rsa" && name != "id_ed25519" && name != "id_ecdsa" && name != "id_dsa" {
				count++
			}
		}
	}
	return count
}

func hasPublicKeySuffix(name string) bool {
	return len(name) > 4 && name[len(name)-4:] == ".pub"
}

func formatDotfileList(dotfiles []string) string {
	if len(dotfiles) <= 5 {
		result := ""
		for i, df := range dotfiles {
			if i > 0 {
				result += ", "
			}
			result += df
		}
		return result
	}

	// For longer lists, just show count and a few examples
	result := itoa(len(dotfiles)) + " files ("
	for i := 0; i < 3; i++ {
		if i > 0 {
			result += ", "
		}
		result += dotfiles[i]
	}
	result += ", ...)"
	return result
}
