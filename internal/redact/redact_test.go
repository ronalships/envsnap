package redact

import "testing"

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"AWS_ACCESS_KEY_ID", true},
		{"AWS_SECRET_ACCESS_KEY", true},
		{"GITHUB_TOKEN", true},
		{"API_KEY", true},
		{"APIKEY", true},
		{"DATABASE_PASSWORD", true},
		{"DB_PASSWD", true},
		{"AUTH_SECRET", true},
		{"PRIVATE_KEY", true},
		{"JWT_SECRET", true},
		{"SESSION_TOKEN", true},
		{"HOME", false},
		{"PATH", false},
		{"USER", false},
		{"GOPATH", false},
		{"NODE_ENV", false},
		{"DEBUG", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := IsSensitive(tt.key)
			if got != tt.expected {
				t.Errorf("IsSensitive(%q) = %v, expected %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		expected string
	}{
		{"AWS_SECRET_ACCESS_KEY", "mysecret", "[REDACTED]"},
		{"GITHUB_TOKEN", "ghp_xxx", "[REDACTED]"},
		{"HOME", "/home/user", "/home/user"},
		{"PATH", "/usr/bin", "/usr/bin"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := Value(tt.key, tt.value)
			if got != tt.expected {
				t.Errorf("Value(%q, %q) = %q, expected %q", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}
