package detector

import (
	"context"
	"regexp"
	"strings"
	"time"
)

// DefaultTimeout is the maximum time a detector should take.
const DefaultTimeout = 2 * time.Second

// Global options for privacy control (all default to false = redact by default)
var (
	RevealAccountIDs bool // Show AWS/GCP account IDs in ARNs
	RevealEmail      bool // Show full email addresses
)

// Result holds the output of a detector.
type Result struct {
	// Section is the markdown section name (e.g., "System", "CLI Tools").
	Section string

	// Items contains the detected key-value pairs or lines.
	Items []Item

	// Error is set if detection failed (partial results may still exist).
	Error error
}

// Item represents a single detected item.
type Item struct {
	Key   string
	Value string
}

// Detector is the interface all detectors must implement.
type Detector interface {
	// Name returns a human-readable name for the detector.
	Name() string

	// Detect runs the detection and returns results.
	// The context should be used for timeout/cancellation.
	Detect(ctx context.Context) (Result, error)
}

// Registry holds all registered detectors in order.
var Registry []Detector

// Register adds a detector to the registry.
func Register(d Detector) {
	Registry = append(Registry, d)
}

// Privacy helper patterns
var (
	// AWS account ID: 12 digits
	awsAccountIDPattern = regexp.MustCompile(`\b[0-9]{12}\b`)
	// EKS ARN: arn:aws:eks:<region>:<account-id>:cluster/<name>
	eksARNPattern = regexp.MustCompile(`arn:aws:eks:([a-z0-9-]+):([0-9]{12}):cluster/(.+)`)
	// GKE context: gke_<project>_<zone>_<cluster>
	gkeContextPattern = regexp.MustCompile(`gke_([^_]+)_([^_]+)_(.+)`)
)

// RedactAccountID redacts AWS account IDs (12-digit numbers) in a string.
func RedactAccountID(s string) string {
	if RevealAccountIDs {
		return s
	}
	return awsAccountIDPattern.ReplaceAllString(s, "************")
}

// RedactEKSARN redacts the account ID portion of an EKS ARN.
func RedactEKSARN(s string) string {
	if RevealAccountIDs {
		return s
	}
	return eksARNPattern.ReplaceAllString(s, "arn:aws:eks:$1:************:cluster/$3")
}

// RedactGKEContext redacts the project ID portion of a GKE context.
func RedactGKEContext(s string) string {
	if RevealAccountIDs {
		return s
	}
	return gkeContextPattern.ReplaceAllString(s, "gke_****_${2}_${3}")
}

// RedactCloudIdentifiers applies all cloud identifier redactions.
func RedactCloudIdentifiers(s string) string {
	if RevealAccountIDs {
		return s
	}
	// Apply EKS ARN redaction first (more specific)
	s = RedactEKSARN(s)
	// Apply GKE context redaction
	s = RedactGKEContext(s)
	// Finally, catch any remaining 12-digit account IDs
	s = RedactAccountID(s)
	return s
}

// RedactEmailLocalPart redacts the local part of an email, keeping the domain.
// "user@example.com" becomes "***@example.com"
func RedactEmailLocalPart(email string) string {
	if RevealEmail {
		return email
	}
	atIndex := strings.LastIndex(email, "@")
	if atIndex <= 0 {
		return email // Not a valid email format
	}
	return "***" + email[atIndex:]
}

// RunAll executes all registered detectors and returns their results.
// Each detector runs with its own timeout context.
func RunAll(ctx context.Context) []Result {
	results := make([]Result, 0, len(Registry))

	for _, d := range Registry {
		// Create a timeout context for this detector
		dctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
		result, err := d.Detect(dctx)
		cancel()

		if err != nil {
			result.Error = err
		}
		results = append(results, result)
	}

	return results
}
