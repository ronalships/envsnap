package detector

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	Register(&AWSDetector{})
}

// AWSDetector captures AWS profile names and regions from ~/.aws/config.
// Never reads credentials.
type AWSDetector struct{}

func (d *AWSDetector) Name() string {
	return "AWS"
}

func (d *AWSDetector) Detect(ctx context.Context) (Result, error) {
	result := Result{Section: "AWS Configuration"}

	home := os.Getenv("HOME")
	if home == "" {
		return result, nil
	}

	configPath := filepath.Join(home, ".aws", "config")
	profiles, regions := parseAWSConfig(configPath)

	if len(profiles) == 0 {
		return result, nil
	}

	// Current profile
	currentProfile := os.Getenv("AWS_PROFILE")
	if currentProfile == "" {
		currentProfile = os.Getenv("AWS_DEFAULT_PROFILE")
	}
	if currentProfile == "" {
		currentProfile = "default"
	}

	result.Items = append(result.Items, Item{Key: "Active Profile", Value: currentProfile})
	result.Items = append(result.Items, Item{Key: "Configured Profiles", Value: strings.Join(profiles, ", ")})

	// Current region
	currentRegion := os.Getenv("AWS_REGION")
	if currentRegion == "" {
		currentRegion = os.Getenv("AWS_DEFAULT_REGION")
	}
	if currentRegion == "" {
		// Try to get from config for current profile
		if r, ok := regions[currentProfile]; ok {
			currentRegion = r
		}
	}
	if currentRegion != "" {
		result.Items = append(result.Items, Item{Key: "Region", Value: currentRegion})
	}

	return result, nil
}

// parseAWSConfig reads profile names and regions from ~/.aws/config.
// It never reads ~/.aws/credentials.
func parseAWSConfig(configPath string) (profiles []string, regions map[string]string) {
	regions = make(map[string]string)

	file, err := os.Open(configPath)
	if err != nil {
		return nil, regions
	}
	defer file.Close()

	var currentProfile string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Profile header: [profile name] or [default]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := line[1 : len(line)-1]
			if section == "default" {
				currentProfile = "default"
			} else if strings.HasPrefix(section, "profile ") {
				currentProfile = strings.TrimPrefix(section, "profile ")
			} else {
				currentProfile = section
			}
			profiles = append(profiles, currentProfile)
			continue
		}

		// Key-value pairs
		if currentProfile != "" && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key == "region" {
				regions[currentProfile] = value
			}
		}
	}

	return profiles, regions
}
