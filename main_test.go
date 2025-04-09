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
			args:    []string{"generate", "-n", "test-profile"},
			wantErr: true,
		},
		{
			name:    "missing new-profile",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole"},
			wantErr: true,
		},
		{
			name:    "all required flags",
			args:    []string{"generate", "-r", "arn:aws:iam::123456789012:role/TestRole", "-n", "test-profile"},
			wantErr: false,
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

			var roleArn, newProfile string
			generateCmd.Flags().StringVarP(&roleArn, "role-arn", "r", "", "")
			generateCmd.Flags().StringVarP(&newProfile, "new-profile", "n", "", "")
			generateCmd.MarkFlagRequired("role-arn")
			generateCmd.MarkFlagRequired("new-profile")

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
