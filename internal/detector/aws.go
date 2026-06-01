package detector

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	Register(&AWSDetector{})
}

// AWSDetector captures AWS profile names and regions from ~/.aws/config.
// Never reads credentials. Only extracts whitelisted keys.
type AWSDetector struct{}

func (d *AWSDetector) Name() string {
	return "AWS"
}

// validRegionPattern matches AWS region format: us-east-1, eu-west-2, ap-northeast-3, etc.
// Max 30 chars, pattern: 2 lowercase letters, dash, lowercase letters, dash, digit(s)
var validRegionPattern = regexp.MustCompile(`^[a-z]{2}-[a-z]+-[0-9]+$`)

// isValidAWSRegion validates that a string looks like an AWS region.
func isValidAWSRegion(s string) bool {
	if len(s) > 30 {
		return false
	}
	return validRegionPattern.MatchString(s)
}

func (d *AWSDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "AWS Configuration"}

	home := os.Getenv("HOME")
	if home == "" {
		return result, nil
	}

	configPath := filepath.Join(home, ".aws", "config")
	profiles := parseAWSConfigSafe(configPath)

	if len(profiles) == 0 {
		return result, nil
	}

	// Current profile from environment
	currentProfile := os.Getenv("AWS_PROFILE")
	if currentProfile == "" {
		currentProfile = os.Getenv("AWS_DEFAULT_PROFILE")
	}
	if currentProfile == "" {
		currentProfile = "default"
	}

	result.Items = append(result.Items, Item{Key: "Active Profile", Value: currentProfile})

	// List profile names only
	profileNames := make([]string, 0, len(profiles))
	for name := range profiles {
		profileNames = append(profileNames, name)
	}
	result.Items = append(result.Items, Item{Key: "Configured Profiles", Value: strings.Join(profileNames, ", ")})

	// Current region - check environment first, then config
	currentRegion := os.Getenv("AWS_REGION")
	if currentRegion == "" {
		currentRegion = os.Getenv("AWS_DEFAULT_REGION")
	}
	if currentRegion == "" {
		if profile, ok := profiles[currentProfile]; ok {
			currentRegion = profile.Region
		}
	}

	// Validate region before outputting
	if currentRegion != "" && isValidAWSRegion(currentRegion) {
		result.Items = append(result.Items, Item{Key: "Region", Value: currentRegion})
	}

	return result, nil
}

// awsProfile holds only the whitelisted fields we extract.
type awsProfile struct {
	Name   string
	Region string
}

// parseAWSConfigSafe reads profile names and regions from ~/.aws/config.
// Uses proper INI parsing and only extracts whitelisted keys.
// It never reads ~/.aws/credentials.
func parseAWSConfigSafe(configPath string) map[string]*awsProfile {
	profiles := make(map[string]*awsProfile)

	file, err := os.Open(configPath)
	if err != nil {
		return profiles
	}
	defer file.Close()

	var currentProfile *awsProfile
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Remove inline comments (but be careful with # in values)
		// INI standard: comments start with # or ; at beginning of line or after whitespace
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and full-line comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}

		// Profile header: [profile name] or [default]
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := trimmed[1 : len(trimmed)-1]
			section = strings.TrimSpace(section)

			var profileName string
			if section == "default" {
				profileName = "default"
			} else if strings.HasPrefix(section, "profile ") {
				profileName = strings.TrimSpace(strings.TrimPrefix(section, "profile "))
			} else {
				// Could be [sso-session name] or other sections - skip
				currentProfile = nil
				continue
			}

			// Validate profile name - should be simple identifier
			if !isValidProfileName(profileName) {
				currentProfile = nil
				continue
			}

			currentProfile = &awsProfile{Name: profileName}
			profiles[profileName] = currentProfile
			continue
		}

		// Key-value pairs - only if we're in a valid profile section
		if currentProfile == nil {
			continue
		}

		// Parse key=value, handling potential whitespace
		eqIndex := strings.Index(trimmed, "=")
		if eqIndex <= 0 {
			continue
		}

		key := strings.TrimSpace(trimmed[:eqIndex])
		value := strings.TrimSpace(trimmed[eqIndex+1:])

		// WHITELIST: Only extract 'region' key
		// Ignore all other keys to prevent leaking arbitrary content
		if key == "region" {
			// Validate the region value
			if isValidAWSRegion(value) {
				currentProfile.Region = value
			}
			// If invalid, silently ignore (don't expose the invalid value)
		}
		// All other keys are intentionally ignored
	}

	return profiles
}

// isValidProfileName checks if a profile name is a reasonable identifier.
// Prevents injection of weird values.
func isValidProfileName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, c := range name {
		// Allow alphanumeric, dash, underscore, dot
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
