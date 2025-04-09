package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Mock for exec.Command to avoid actual AWS CLI calls
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess isn't a real test - it's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Get the command being "executed"
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(1)
	}

	// Mock responses based on the command
	if args[0] == "aws" {
		// Check for list-mfa-devices command (more flexible matching)
		if contains(args, "list-mfa-devices") {
			fmt.Fprintf(os.Stdout, "arn:aws:iam::123456789012:mfa/user\n")
			os.Exit(0)
		}

		// Check for assume-role command (more flexible matching)
		if contains(args, "assume-role") {
			fmt.Fprintf(os.Stdout, `{
				"AccessKeyId": "ASIAMOCK123456789012",
				"SecretAccessKey": "mockSecretKey123456789012345678901234",
				"SessionToken": "mockSessionToken123456789012345678901234567890123456789012345678901234567890",
				"Expiration": "2023-12-31T23:59:59Z"
			}`)
			os.Exit(0)
		}

		// Check for configure command (more flexible matching)
		if contains(args, "configure") {
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "Unrecognized command: %v\n", args)
	os.Exit(1)
}

// Helper function to check if a slice contains a string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// Test getMFADeviceARN function
func TestGetMFADeviceARN(t *testing.T) {
	// Save original exec.Command and restore it after the test
	origExecCommand := execCommand
	execCommand = mockExecCommand
	defer func() { execCommand = origExecCommand }()

	// Test with profile
	mfaSerial, err := getMFADeviceARN("--profile", "test-profile")
	if err != nil {
		t.Errorf("getMFADeviceARN with profile failed: %v", err)
	}
	if mfaSerial != "arn:aws:iam::123456789012:mfa/user" {
		t.Errorf("Expected MFA serial 'arn:aws:iam::123456789012:mfa/user', got '%s'", mfaSerial)
	}

	// Test without profile
	mfaSerial, err = getMFADeviceARN("", "")
	if err != nil {
		t.Errorf("getMFADeviceARN without profile failed: %v", err)
	}
	if mfaSerial != "arn:aws:iam::123456789012:mfa/user" {
		t.Errorf("Expected MFA serial 'arn:aws:iam::123456789012:mfa/user', got '%s'", mfaSerial)
	}
}

// Test assumeRole function
func TestAssumeRole(t *testing.T) {
	// Save original exec.Command and restore it after the test
	origExecCommand := execCommand
	execCommand = mockExecCommand
	defer func() { execCommand = origExecCommand }()

	// Test with MFA
	creds, err := assumeRole("--profile", "test-profile", "arn:aws:iam::123456789012:role/TestRole",
		"arn:aws:iam::123456789012:mfa/user", "123456", 3600)
	if err != nil {
		t.Errorf("assumeRole with MFA failed: %v", err)
	}
	if creds.AccessKeyId != "ASIAMOCK123456789012" {
		t.Errorf("Expected AccessKeyId 'ASIAMOCK123456789012', got '%s'", creds.AccessKeyId)
	}

	// Test without MFA
	creds, err = assumeRole("--profile", "test-profile", "arn:aws:iam::123456789012:role/TestRole",
		"", "", 3600)
	if err != nil {
		t.Errorf("assumeRole without MFA failed: %v", err)
	}
	if creds.AccessKeyId != "ASIAMOCK123456789012" {
		t.Errorf("Expected AccessKeyId 'ASIAMOCK123456789012', got '%s'", creds.AccessKeyId)
	}
}

// Test configureAWSProfile function
func TestConfigureAWSProfile(t *testing.T) {
	// Save original exec.Command and restore it after the test
	origExecCommand := execCommand
	execCommand = mockExecCommand
	defer func() { execCommand = origExecCommand }()

	// Create test credentials
	testCreds := &Credentials{
		AccessKeyId:     "ASIAMOCK123456789012",
		SecretAccessKey: "mockSecretKey123456789012345678901234",
		SessionToken:    "mockSessionToken123456789012345678901234567890123456789012345678901234567890",
		Expiration:      time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	// Test with region
	err := configureAWSProfile("test-profile", testCreds, "source-profile", "us-west-2")
	if err != nil {
		t.Errorf("configureAWSProfile with region failed: %v", err)
	}

	// Test without region
	err = configureAWSProfile("test-profile", testCreds, "source-profile", "")
	if err != nil {
		t.Errorf("configureAWSProfile without region failed: %v", err)
	}
}
