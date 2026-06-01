package detector

import (
	"context"
	"os/exec"
	"strings"
)

func init() {
	Register(&ToolsDetector{})
}

// Tool defines a CLI tool to detect with its version command.
type Tool struct {
	Name       string
	Command    string
	VersionArg string
	Parser     func(string) string // Optional custom parser
}

// DefaultTools is the list of common dev tools to detect.
var DefaultTools = []Tool{
	{Name: "git", Command: "git", VersionArg: "--version"},
	{Name: "gh", Command: "gh", VersionArg: "--version"},
	{Name: "docker", Command: "docker", VersionArg: "--version"},
	{Name: "kubectl", Command: "kubectl", VersionArg: "version", Parser: parseKubectlVersion},
	{Name: "helm", Command: "helm", VersionArg: "version", Parser: parseHelmVersion},
	{Name: "terraform", Command: "terraform", VersionArg: "--version"},
	{Name: "aws", Command: "aws", VersionArg: "--version"},
	{Name: "gcloud", Command: "gcloud", VersionArg: "--version", Parser: parseGcloudVersion},
	{Name: "az", Command: "az", VersionArg: "--version", Parser: parseAzVersion},
	{Name: "node", Command: "node", VersionArg: "--version"},
	{Name: "npm", Command: "npm", VersionArg: "--version"},
	{Name: "yarn", Command: "yarn", VersionArg: "--version"},
	{Name: "pnpm", Command: "pnpm", VersionArg: "--version"},
	{Name: "bun", Command: "bun", VersionArg: "--version"},
	{Name: "deno", Command: "deno", VersionArg: "--version", Parser: parseDenoVersion},
	{Name: "python", Command: "python3", VersionArg: "--version"},
	{Name: "pip", Command: "pip3", VersionArg: "--version"},
	{Name: "go", Command: "go", VersionArg: "version"},
	{Name: "rustc", Command: "rustc", VersionArg: "--version"},
	{Name: "cargo", Command: "cargo", VersionArg: "--version"},
	{Name: "java", Command: "java", VersionArg: "-version"},
	{Name: "ruby", Command: "ruby", VersionArg: "--version"},
	{Name: "php", Command: "php", VersionArg: "--version", Parser: parsePhpVersion},
	{Name: "make", Command: "make", VersionArg: "--version", Parser: parseMakeVersion},
	{Name: "cmake", Command: "cmake", VersionArg: "--version", Parser: parseCmakeVersion},
	{Name: "curl", Command: "curl", VersionArg: "--version", Parser: parseCurlVersion},
	{Name: "wget", Command: "wget", VersionArg: "--version", Parser: parseWgetVersion},
	{Name: "jq", Command: "jq", VersionArg: "--version"},
	{Name: "yq", Command: "yq", VersionArg: "--version"},
}

// ToolsDetector detects installed CLI tools and their versions.
type ToolsDetector struct{}

func (d *ToolsDetector) Name() string {
	return "CLI Tools"
}

func (d *ToolsDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "CLI Tools"}

	for _, tool := range DefaultTools {
		version := detectToolVersion(ctx, tool)
		if version != "" {
			result.Items = append(result.Items, Item{Key: tool.Name, Value: version})
		}
	}

	return result, nil
}

func detectToolVersion(ctx context.Context, tool Tool) string {
	// Check if command exists
	_, err := exec.LookPath(tool.Command)
	if err != nil {
		return ""
	}

	cmd := exec.CommandContext(ctx, tool.Command, tool.VersionArg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	output := strings.TrimSpace(string(out))
	if tool.Parser != nil {
		return tool.Parser(output)
	}

	// Default: return first line
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

// Custom parsers for tools with non-standard version output

func parseKubectlVersion(output string) string {
	// kubectl version --client returns JSON or text depending on version
	if strings.Contains(output, "Client Version:") {
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, "Client Version:") {
				return strings.TrimPrefix(line, "Client Version: ")
			}
		}
	}
	return strings.Split(output, "\n")[0]
}

func parseHelmVersion(output string) string {
	// helm version returns version.BuildInfo{Version:"v3.x.x", ...}
	if strings.Contains(output, "Version:") {
		start := strings.Index(output, `Version:"`)
		if start != -1 {
			start += len(`Version:"`)
			end := strings.Index(output[start:], `"`)
			if end != -1 {
				return output[start : start+end]
			}
		}
	}
	return output
}

func parseGcloudVersion(output string) string {
	// First line is "Google Cloud SDK x.x.x"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parseAzVersion(output string) string {
	// First line contains the version
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "azure-cli") {
			return strings.TrimSpace(line)
		}
	}
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parseDenoVersion(output string) string {
	// First line is "deno x.x.x"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parsePhpVersion(output string) string {
	// First line is "PHP x.x.x ..."
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parseMakeVersion(output string) string {
	// First line is "GNU Make x.x"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parseCmakeVersion(output string) string {
	// First line is "cmake version x.x.x"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parseCurlVersion(output string) string {
	// First line is "curl x.x.x ..."
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}

func parseWgetVersion(output string) string {
	// First line is "GNU Wget x.x.x ..."
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return output
}
