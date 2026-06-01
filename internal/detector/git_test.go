package detector

import "testing"

func TestRedactEmailLocalPart(t *testing.T) {
	// Ensure redaction is enabled for tests
	originalReveal := RevealEmail
	RevealEmail = false
	defer func() { RevealEmail = originalReveal }()

	tests := []struct {
		input    string
		expected string
	}{
		// Normal emails should have local part redacted
		{"ronal@orchestro.ai", "***@orchestro.ai"},
		{"user@example.com", "***@example.com"},
		{"john.doe@company.co.uk", "***@company.co.uk"},
		{"test+alias@gmail.com", "***@gmail.com"},

		// Edge cases
		{"a@b.com", "***@b.com"},            // Single char local part
		{"@example.com", "@example.com"},   // No local part (invalid, pass through)
		{"noatsign", "noatsign"},           // No @ sign
		{"", ""},                           // Empty string

		// Multiple @ signs (use last one)
		{"weird@email@domain.com", "***@domain.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactEmailLocalPart(tt.input)
			if got != tt.expected {
				t.Errorf("RedactEmailLocalPart(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactEmailLocalPart_RevealEnabled(t *testing.T) {
	// Enable reveal mode
	originalReveal := RevealEmail
	RevealEmail = true
	defer func() { RevealEmail = originalReveal }()

	tests := []struct {
		input    string
		expected string
	}{
		// With reveal enabled, nothing should be redacted
		{"ronal@orchestro.ai", "ronal@orchestro.ai"},
		{"user@example.com", "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactEmailLocalPart(tt.input)
			if got != tt.expected {
				t.Errorf("RedactEmailLocalPart(%q) with reveal=true = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}
