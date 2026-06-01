package detector

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

func init() {
	Register(&DockerDetector{})
}

// DockerDetector captures Docker version, daemon status, and container count.
type DockerDetector struct{}

func (d *DockerDetector) Name() string {
	return "Docker"
}

func (d *DockerDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "Docker"}

	// Check if docker is installed
	if _, err := exec.LookPath("docker"); err != nil {
		return result, nil
	}

	// Docker version
	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Client.Version}}")
	out, err := cmd.Output()
	if err == nil {
		result.Items = append(result.Items, Item{Key: "Client Version", Value: strings.TrimSpace(string(out))})
	}

	// Check daemon status and server version
	cmd = exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	out, err = cmd.Output()
	if err != nil {
		result.Items = append(result.Items, Item{Key: "Daemon Status", Value: "not running"})
	} else {
		result.Items = append(result.Items, Item{Key: "Server Version", Value: strings.TrimSpace(string(out))})
		result.Items = append(result.Items, Item{Key: "Daemon Status", Value: "running"})

		// Count running containers (only if daemon is running)
		containerCount := countRunningContainers(ctx)
		if containerCount >= 0 {
			result.Items = append(result.Items, Item{Key: "Running Containers", Value: itoa(containerCount)})
		}
	}

	// Docker context
	cmd = exec.CommandContext(ctx, "docker", "context", "show")
	out, err = cmd.Output()
	if err == nil {
		contextName := strings.TrimSpace(string(out))
		if contextName != "" && contextName != "default" {
			result.Items = append(result.Items, Item{Key: "Context", Value: contextName})
		}
	}

	return result, nil
}

func countRunningContainers(ctx context.Context) int {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return -1
	}

	// Each line is a JSON object for a container
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Validate it's JSON
		var container map[string]interface{}
		if json.Unmarshal([]byte(line), &container) == nil {
			count++
		}
	}
	return count
}

// Simple int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
