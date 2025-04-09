package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Add this at the top of the file, after the imports
var execCommand = exec.Command
var getAWSConfigValue = getAWSConfigValueFunc

// AWS STS Credentials structure
type Credentials struct {
	AccessKeyId     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      time.Time `json:"Expiration"`
}

// generateTempProfile is the main function that generates temporary AWS credentials
func generateTempProfile(sourceProfile, roleArn, mfaToken, newProfile, region string, duration int) error {
	var mfaSerial string
	var err error

	// Use default profile if no source profile is specified
	profileArg := "--profile"
	profileValue := sourceProfile

	if sourceProfile == "" {
		fmt.Println("No source profile specified, using default AWS profile")
		profileArg = "" // Don't use --profile flag when using default profile
		profileValue = ""
	} else {
		fmt.Printf("Using source profile: %s\n", sourceProfile)
	}

	// Only get MFA device if token is provided
	if mfaToken != "" {
		// Get the MFA device ARN for the source profile
		fmt.Printf("Getting MFA device ARN...\n")
		mfaSerial, err = getMFADeviceARN(profileArg, profileValue)
		if err != nil {
			return fmt.Errorf("error getting MFA device: %w", err)
		}

		if mfaSerial == "None" || mfaSerial == "" {
			return fmt.Errorf("no MFA device found, but MFA token was provided")
		}

		fmt.Printf("Found MFA device: %s\n", mfaSerial)
	} else {
		fmt.Println("No MFA token provided, assuming role without MFA")
	}

	// Display duration information
	if duration == 3600 {
		fmt.Println("Using default session duration of 1 hour (3600 seconds)")
	} else {
		fmt.Printf("Using specified session duration of %d hours (%d seconds)\n", duration/3600, duration)
	}

	// Assume the role with or without MFA
	fmt.Printf("Assuming role %s...\n", roleArn)
	credentials, err := assumeRole(profileArg, profileValue, roleArn, mfaSerial, mfaToken, duration)
	if err != nil {
		return fmt.Errorf("error assuming role: %w\n\nThis could be due to:\n"+
			"1. The MFA token has expired or is incorrect\n"+
			"2. Your device's time might be out of sync with AWS servers\n"+
			"3. You might not have permission to assume this role\n"+
			"4. The requested duration exceeds the role's maximum session duration\n"+
			"5. MFA might be required for this role\n\n"+
			"Try again with a shorter duration (e.g., 1 hour = 3600 seconds) or provide an MFA token if required", err)
	}

	// Set up the new profile with the credentials
	fmt.Printf("Setting up profile %s...\n", newProfile)
	if err := configureAWSProfile(newProfile, credentials, sourceProfile, region); err != nil {
		return fmt.Errorf("error configuring AWS profile: %w", err)
	}

	// Calculate session duration
	currentTime := time.Now()
	durationHours := int(credentials.Expiration.Sub(currentTime).Hours())
	durationMinutes := int(credentials.Expiration.Sub(currentTime).Minutes()) % 60

	fmt.Printf("Temporary credentials for profile '%s' have been successfully configured\n", newProfile)
	fmt.Printf("Credentials will expire at: %s (valid for approximately %dh %dm)\n\n",
		credentials.Expiration.Local().Format("2006-01-02 15:04:05 MST"), durationHours, durationMinutes)
	fmt.Printf("You can now use these credentials with: aws --profile %s <command>\n", newProfile)

	return nil
}

// getMFADeviceARN gets the MFA device ARN for the given profile
func getMFADeviceARN(profileArg, profileValue string) (string, error) {
	var args []string

	// Only add profile arguments if a profile is specified
	if profileArg != "" && profileValue != "" {
		args = append(args, profileArg, profileValue)
	}

	args = append(args, "iam", "list-mfa-devices", "--query", "MFADevices[0].SerialNumber", "--output", "text")

	cmd := execCommand("aws", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get MFA device: %w\nOutput: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// assumeRole assumes the specified role with or without MFA and returns the credentials
func assumeRole(profileArg, profileValue, roleArn, mfaSerial, mfaToken string, duration int) (*Credentials, error) {
	sessionName := fmt.Sprintf("TempSession-%d", time.Now().Unix())

	// Build command arguments
	var args []string

	// Only add profile arguments if a profile is specified
	if profileArg != "" && profileValue != "" {
		args = append(args, profileArg, profileValue)
	}

	args = append(args, "sts", "assume-role",
		"--role-arn", roleArn,
		"--role-session-name", sessionName,
		"--duration-seconds", fmt.Sprintf("%d", duration),
		"--query", "Credentials",
		"--output", "json")

	// Add MFA parameters only if MFA token is provided
	if mfaToken != "" && mfaSerial != "" {
		args = append(args, "--serial-number", mfaSerial, "--token-code", mfaToken)
	}

	cmd := execCommand("aws", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w\nOutput: %s", err, string(output))
	}

	var credentials Credentials
	if err := json.Unmarshal(output, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	if credentials.AccessKeyId == "" || credentials.SecretAccessKey == "" || credentials.SessionToken == "" {
		return nil, fmt.Errorf("failed to get valid credentials from AWS response")
	}

	return &credentials, nil
}

// configureAWSProfile sets up a new AWS profile with the given credentials
func configureAWSProfile(profile string, credentials *Credentials, sourceProfile, region string) error {
	// Set the AWS access key
	if err := runAWSConfigureCommand(profile, "aws_access_key_id", credentials.AccessKeyId); err != nil {
		return fmt.Errorf("failed to set access key: %w", err)
	}

	// Set the AWS secret key
	if err := runAWSConfigureCommand(profile, "aws_secret_access_key", credentials.SecretAccessKey); err != nil {
		return fmt.Errorf("failed to set secret key: %w", err)
	}

	// Set the AWS session token
	if err := runAWSConfigureCommand(profile, "aws_session_token", credentials.SessionToken); err != nil {
		return fmt.Errorf("failed to set session token: %w", err)
	}

	// Set the region for the new profile
	if region != "" {
		// Use the provided region
		fmt.Printf("Setting region to %s...\n", region)
		if err := runAWSConfigureCommand(profile, "region", region); err != nil {
			return fmt.Errorf("failed to set region: %w", err)
		}
	} else {
		// If no region was provided, use the region from the source profile
		sourceRegion, err := getAWSConfigValue(sourceProfile, "region")
		if err != nil {
			fmt.Println("Warning: Error getting region from source profile")
		} else if sourceRegion != "" {
			fmt.Printf("Setting region to %s (from source profile)...\n", sourceRegion)
			if err := runAWSConfigureCommand(profile, "region", sourceRegion); err != nil {
				return fmt.Errorf("failed to set region: %w", err)
			}
		} else {
			fmt.Println("Warning: No region specified and source profile has no region set.")
		}
	}

	return nil
}

// runAWSConfigureCommand runs the aws configure command to set a specific value
func runAWSConfigureCommand(profile, key, value string) error {
	cmd := execCommand("aws", "configure", "set", key, value, "--profile", profile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}
	return nil
}

// getAWSConfigValue gets a configuration value from an AWS profile
func getAWSConfigValueFunc(profile, key string) (string, error) {
	cmd := execCommand("aws", "configure", "get", key, "--profile", profile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get %s for profile %s: %w", key, profile, err)
	}
	return strings.TrimSpace(string(output)), nil
}
