package redact

import "strings"

// SensitivePatterns contains substrings that indicate sensitive env var keys.
var SensitivePatterns = []string{
	"KEY",
	"SECRET",
	"TOKEN",
	"PASSWORD",
	"PASSWD",
	"CREDENTIAL",
	"AUTH",
	"PRIVATE",
	"ACCESS",
	"API_KEY",
	"APIKEY",
	"BEARER",
	"JWT",
	"SESSION",
	"COOKIE",
	"CERT",
	"SIGNING",
}

// IsSensitive checks if a key name indicates sensitive data.
func IsSensitive(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range SensitivePatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

// Value redacts a value if the key is sensitive.
func Value(key, value string) string {
	if IsSensitive(key) {
		return "[REDACTED]"
	}
	return value
}
