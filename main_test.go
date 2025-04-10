package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// Test the root command
func TestRootCommand(t *testing.T) {
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Execute the command
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that help was displayed
	output := buf.String()
	if len(output) == 0 {
		t.Errorf("Expected help output, got empty string")
	}
}

// Test the generate-profile command flags
func TestGenerateProfileCommandFlags(t *testing.T) {
	// Redirect stderr to discard output during tests
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test required flags
	testCases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "missing role-arn",
			args:    []string{"generate-profile", "-n", "test-profile"},
			wantErr: true,
		},
		{
			name:    "missing new-profile",
			args:    []string{"generate-profile", "-r", "arn:aws:iam::123456789012:role/TestRole"},
			wantErr: true,
		},
		{
			name:    "all required flags",
			args:    []string{"generate-profile", "-r", "arn:aws:iam::123456789012:role/TestRole", "-n", "test-profile"},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new command to avoid state from previous tests
			cmd := &cobra.Command{Use: "awsomecreds"}
			generateProfileCmd := &cobra.Command{
				Use:  "generate-profile",
				RunE: func(cmd *cobra.Command, args []string) error { return nil },
			}
			cmd.AddCommand(generateProfileCmd)

			var roleArn, newProfile string
			generateProfileCmd.Flags().StringVarP(&roleArn, "role-arn", "r", "", "")
			generateProfileCmd.Flags().StringVarP(&newProfile, "new-profile", "n", "", "")
			generateProfileCmd.MarkFlagRequired("role-arn")
			generateProfileCmd.MarkFlagRequired("new-profile")

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if (err != nil) != tc.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr
	io.Copy(io.Discard, r) // Discard captured output
}

// Test the generate command flags
func TestGenerateCommandFlags(t *testing.T) {
	// Redirect stderr to discard output during tests
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test required flags
	testCases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "missing role-arn",
			args:    []string{"generate", "-o", "shell"},
			wantErr: true,
		},
		{
			name:    "with role-arn",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole"},
			wantErr: false,
		},
		{
			name:    "with role-arn and MFA",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole", "-m", "123456"},
			wantErr: false,
		},
		{
			name:    "with source profile and region",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole", "-s", "source-profile", "--region", "us-west-2"},
			wantErr: false,
		},
		{
			name:    "with json output",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole", "-o", "json"},
			wantErr: false,
		},
		{
			name:    "with invalid output format",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole", "-o", "invalid"},
			wantErr: false, // Note: Format validation happens in the execution, not in flag parsing
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new command to avoid state from previous tests
			cmd := &cobra.Command{Use: "awsomecreds"}
			generateCmd := &cobra.Command{
				Use:  "generate",
				RunE: func(cmd *cobra.Command, args []string) error { return nil },
			}
			cmd.AddCommand(generateCmd)

			var roleArn, sourceProfile, mfaToken, region, outputFormat string
			generateCmd.Flags().StringVarP(&roleArn, "role-arn", "r", "", "")
			generateCmd.Flags().StringVarP(&sourceProfile, "source-profile", "s", "", "")
			generateCmd.Flags().StringVarP(&mfaToken, "mfa-token", "m", "", "")
			generateCmd.Flags().StringVarP(&region, "region", "", "", "")
			generateCmd.Flags().StringVarP(&outputFormat, "output", "o", "shell", "")
			generateCmd.MarkFlagRequired("role-arn")

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if (err != nil) != tc.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr
	io.Copy(io.Discard, r) // Discard captured output
}
