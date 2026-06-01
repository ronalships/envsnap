package detector

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	Register(&KubernetesDetector{})
}

// KubernetesDetector captures Kubernetes context and cluster names.
// Never captures tokens, certs, or server URLs.
type KubernetesDetector struct{}

func (d *KubernetesDetector) Name() string {
	return "Kubernetes"
}

func (d *KubernetesDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "Kubernetes"}

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return result, nil
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Check if kubeconfig exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return result, nil
	}

	// Try kubectl first for current context
	currentContext := getCurrentKubeContext(ctx)
	if currentContext != "" {
		result.Items = append(result.Items, Item{Key: "Current Context", Value: currentContext})
	}

	// Parse kubeconfig for available contexts and clusters
	contexts, clusters := parseKubeconfig(kubeconfigPath)

	if len(contexts) > 0 {
		result.Items = append(result.Items, Item{Key: "Available Contexts", Value: strings.Join(contexts, ", ")})
	}

	if len(clusters) > 0 {
		result.Items = append(result.Items, Item{Key: "Configured Clusters", Value: strings.Join(clusters, ", ")})
	}

	return result, nil
}

func getCurrentKubeContext(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "kubectl", "config", "current-context")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// parseKubeconfig extracts context and cluster names from kubeconfig.
// Uses simple line parsing to avoid yaml dependency.
// Only extracts names, never sensitive data.
func parseKubeconfig(path string) (contexts []string, clusters []string) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var inContexts, inClusters bool

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track which section we're in
		if strings.HasPrefix(trimmed, "contexts:") {
			inContexts = true
			inClusters = false
			continue
		}
		if strings.HasPrefix(trimmed, "clusters:") {
			inClusters = true
			inContexts = false
			continue
		}
		if strings.HasPrefix(trimmed, "users:") || strings.HasPrefix(trimmed, "current-context:") {
			inContexts = false
			inClusters = false
			continue
		}

		// Extract names (look for "- name:" or "name:" patterns)
		if inContexts && strings.HasPrefix(trimmed, "- name:") {
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
			name = strings.Trim(name, `"'`)
			if name != "" {
				contexts = append(contexts, name)
			}
		}
		if inClusters && strings.HasPrefix(trimmed, "- name:") {
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
			name = strings.Trim(name, `"'`)
			if name != "" {
				clusters = append(clusters, name)
			}
		}
	}

	return contexts, clusters
}
