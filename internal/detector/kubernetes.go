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
// Redacts account IDs in EKS ARNs and GKE context names by default.
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
		// Redact account IDs in EKS ARNs and GKE contexts
		redactedContext := RedactCloudIdentifiers(currentContext)
		result.Items = append(result.Items, Item{Key: "Current Context", Value: redactedContext})
	}

	// Parse kubeconfig for available contexts and clusters
	contexts, clusters := parseKubeconfig(kubeconfigPath)

	if len(contexts) > 0 {
		// Redact each context name
		redactedContexts := make([]string, len(contexts))
		for i, c := range contexts {
			redactedContexts[i] = RedactCloudIdentifiers(c)
		}
		result.Items = append(result.Items, Item{Key: "Available Contexts", Value: strings.Join(redactedContexts, ", ")})
	}

	if len(clusters) > 0 {
		// Redact each cluster name
		redactedClusters := make([]string, len(clusters))
		for i, c := range clusters {
			redactedClusters[i] = RedactCloudIdentifiers(c)
		}
		result.Items = append(result.Items, Item{Key: "Configured Clusters", Value: strings.Join(redactedClusters, ", ")})
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
			if name != "" && isValidKubeName(name) {
				contexts = append(contexts, name)
			}
		}
		if inClusters && strings.HasPrefix(trimmed, "- name:") {
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
			name = strings.Trim(name, `"'`)
			if name != "" && isValidKubeName(name) {
				clusters = append(clusters, name)
			}
		}
	}

	return contexts, clusters
}

// isValidKubeName checks if a name is a reasonable Kubernetes identifier.
// Prevents parsing errors from leaking arbitrary content.
func isValidKubeName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	// Allow typical k8s name chars plus : and / for ARNs
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' ||
			c == ':' || c == '/') {
			return false
		}
	}
	return true
}
