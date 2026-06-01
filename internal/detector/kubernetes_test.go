package detector

import "testing"

func TestRedactEKSARN(t *testing.T) {
	// Ensure redaction is enabled for tests
	originalReveal := RevealAccountIDs
	RevealAccountIDs = false
	defer func() { RevealAccountIDs = originalReveal }()

	tests := []struct {
		input    string
		expected string
	}{
		// EKS ARNs should have account ID redacted
		{
			"arn:aws:eks:us-east-1:505753670778:cluster/stage-eks-us-east-1-cluster",
			"arn:aws:eks:us-east-1:************:cluster/stage-eks-us-east-1-cluster",
		},
		{
			"arn:aws:eks:ap-northeast-3:149614785292:cluster/prod-cluster",
			"arn:aws:eks:ap-northeast-3:************:cluster/prod-cluster",
		},
		{
			"arn:aws:eks:eu-west-1:123456789012:cluster/my-cluster",
			"arn:aws:eks:eu-west-1:************:cluster/my-cluster",
		},
		// Non-ARN strings should be unchanged
		{
			"minikube",
			"minikube",
		},
		{
			"docker-desktop",
			"docker-desktop",
		},
		{
			"kind-kind",
			"kind-kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactEKSARN(tt.input)
			if got != tt.expected {
				t.Errorf("RedactEKSARN(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactGKEContext(t *testing.T) {
	// Ensure redaction is enabled for tests
	originalReveal := RevealAccountIDs
	RevealAccountIDs = false
	defer func() { RevealAccountIDs = originalReveal }()

	tests := []struct {
		input    string
		expected string
	}{
		// GKE contexts should have project ID redacted
		{
			"gke_my-company-prod-123456_us-central1-a_prod-cluster",
			"gke_****_us-central1-a_prod-cluster",
		},
		{
			"gke_project-id_europe-west1-b_staging",
			"gke_****_europe-west1-b_staging",
		},
		// Non-GKE strings should be unchanged
		{
			"minikube",
			"minikube",
		},
		{
			"arn:aws:eks:us-east-1:123456789012:cluster/test",
			"arn:aws:eks:us-east-1:123456789012:cluster/test", // EKS, not GKE
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactGKEContext(tt.input)
			if got != tt.expected {
				t.Errorf("RedactGKEContext(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactCloudIdentifiers(t *testing.T) {
	// Ensure redaction is enabled for tests
	originalReveal := RevealAccountIDs
	RevealAccountIDs = false
	defer func() { RevealAccountIDs = originalReveal }()

	tests := []struct {
		input    string
		expected string
	}{
		// EKS ARN
		{
			"arn:aws:eks:us-east-1:505753670778:cluster/my-cluster",
			"arn:aws:eks:us-east-1:************:cluster/my-cluster",
		},
		// GKE context
		{
			"gke_my-project_us-central1_cluster",
			"gke_****_us-central1_cluster",
		},
		// Plain context names should be unchanged
		{
			"minikube",
			"minikube",
		},
		{
			"docker-desktop",
			"docker-desktop",
		},
		// AKS context (no special redaction, but should pass through)
		{
			"aks-prod-westus2",
			"aks-prod-westus2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactCloudIdentifiers(tt.input)
			if got != tt.expected {
				t.Errorf("RedactCloudIdentifiers(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactCloudIdentifiers_RevealEnabled(t *testing.T) {
	// Enable reveal mode
	originalReveal := RevealAccountIDs
	RevealAccountIDs = true
	defer func() { RevealAccountIDs = originalReveal }()

	tests := []struct {
		input    string
		expected string
	}{
		// With reveal enabled, nothing should be redacted
		{
			"arn:aws:eks:us-east-1:505753670778:cluster/my-cluster",
			"arn:aws:eks:us-east-1:505753670778:cluster/my-cluster",
		},
		{
			"gke_my-project_us-central1_cluster",
			"gke_my-project_us-central1_cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactCloudIdentifiers(tt.input)
			if got != tt.expected {
				t.Errorf("RedactCloudIdentifiers(%q) with reveal=true = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidKubeName(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid names
		{"minikube", true},
		{"docker-desktop", true},
		{"kind-kind", true},
		{"arn:aws:eks:us-east-1:123456789012:cluster/test", true},
		{"gke_project_zone_cluster", true},

		// Invalid names
		{"", false},
		{"name with spaces", false},
		{"name\nwith\nnewline", false},
		{"$(whoami)", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidKubeName(tt.input)
			if got != tt.expected {
				t.Errorf("isValidKubeName(%q) = %v, expected %v", tt.input, got, tt.expected)
			}
		})
	}
}
