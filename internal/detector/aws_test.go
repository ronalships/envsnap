package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid regions
		{"us-east-1", true},
		{"us-west-2", true},
		{"eu-west-1", true},
		{"ap-northeast-1", true},
		{"ap-southeast-2", true},
		{"sa-east-1", true},
		{"ca-central-1", true},
		{"me-south-1", true},
		{"af-south-1", true},

		// Invalid regions
		{"", false},
		{"us-east", false},         // Missing number
		{"US-EAST-1", false},       // Uppercase
		{"us_east_1", false},       // Underscores
		{"useast1", false},         // No dashes
		{"invalid", false},         // Just a word
		{"123-456-789", false},     // Numbers only
		{"a]ws ecr get-login", false}, // Shell command injection
		{"us-east-1; rm -rf /", false}, // Command injection attempt
		{strings.Repeat("a", 100), false}, // Too long
		{"aws ecr get-login-password --region ap-northeast-3 | docker login --username AWS --password-stdin 149614785292.dkr.ecr.ap-northeast-3.amazonaws.com", false}, // The actual bug case
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidAWSRegion(tt.input)
			if got != tt.expected {
				t.Errorf("isValidAWSRegion(%q) = %v, expected %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidProfileName(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid names
		{"default", true},
		{"production", true},
		{"dev-account", true},
		{"staging_env", true},
		{"company.prod", true},
		{"Profile123", true},

		// Invalid names
		{"", false},
		{"profile with spaces", false},
		{"profile\nwith\nnewlines", false},
		{"$(whoami)", false},
		{"`id`", false},
		{"profile;rm -rf /", false},
		{strings.Repeat("a", 100), false}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidProfileName(tt.input)
			if got != tt.expected {
				t.Errorf("isValidProfileName(%q) = %v, expected %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseAWSConfigSafe(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "envsnap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name            string
		configContent   string
		expectedCount   int
		expectedRegions map[string]string
		description     string
	}{
		{
			name: "normal_config",
			configContent: `[default]
region = us-east-1
output = json

[profile production]
region = us-west-2
output = text
`,
			expectedCount: 2,
			expectedRegions: map[string]string{
				"default":    "us-east-1",
				"production": "us-west-2",
			},
			description: "Normal AWS config with two profiles",
		},
		{
			name: "malicious_multiline_value",
			configContent: `[default]
region = aws ecr get-login-password --region ap-northeast-3 | docker login --username AWS --password-stdin 149614785292.dkr.ecr.ap-northeast-3.amazonaws.com
`,
			expectedCount:   1,
			expectedRegions: map[string]string{}, // Region should be rejected
			description:     "Malicious config with shell command as region value",
		},
		{
			name: "embedded_account_id",
			configContent: `[profile prod]
region = us-east-1
account_id = 149614785292
role_arn = arn:aws:iam::149614785292:role/admin
`,
			expectedCount: 1,
			expectedRegions: map[string]string{
				"prod": "us-east-1",
			},
			description: "Config with account ID (should be ignored, only region extracted)",
		},
		{
			name: "very_long_values",
			configContent: `[default]
region = ` + strings.Repeat("a", 1000) + `
`,
			expectedCount:   1,
			expectedRegions: map[string]string{}, // Region too long, should be rejected
			description:     "Config with very long region value",
		},
		{
			name: "sso_session_ignored",
			configContent: `[default]
region = us-east-1

[sso-session my-sso]
sso_start_url = https://my-sso-portal.awsapps.com/start
sso_region = us-east-1

[profile sso-user]
sso_session = my-sso
region = us-west-2
`,
			expectedCount: 2, // default and sso-user (sso-session is skipped)
			expectedRegions: map[string]string{
				"default":  "us-east-1",
				"sso-user": "us-west-2",
			},
			description: "Config with SSO session (should be ignored)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test config
			configPath := filepath.Join(tmpDir, tt.name+"_config")
			if err := os.WriteFile(configPath, []byte(tt.configContent), 0600); err != nil {
				t.Fatal(err)
			}

			profiles := parseAWSConfigSafe(configPath)

			if len(profiles) != tt.expectedCount {
				t.Errorf("%s: expected %d profiles, got %d", tt.description, tt.expectedCount, len(profiles))
				for name := range profiles {
					t.Logf("  - profile: %q", name)
				}
			}

			for profileName, expectedRegion := range tt.expectedRegions {
				profile, ok := profiles[profileName]
				if !ok {
					t.Errorf("%s: expected profile %q not found", tt.description, profileName)
					continue
				}
				if profile.Region != expectedRegion {
					t.Errorf("%s: profile %q region = %q, expected %q",
						tt.description, profileName, profile.Region, expectedRegion)
				}
			}
		})
	}
}
