package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	sourceProfile string
	roleArn       string
	mfaToken      string
	newProfile    string
	region        string
	duration      int
)

var rootCmd = &cobra.Command{
	Use:   "awsomecreds",
	Short: "Assume roles and generate temporary AWS credential profiles",
	Long:  `AWSomeCreds is a CLI tool that generates temporary AWS credentials using AWS STS and sets them using the AWS CLI. It allows you to assume roles with or without MFA authentication and create temporary profiles for tools that support AWS CLI profiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		cmd.Help()
	},
}

var generateProfileCmd = &cobra.Command{
	Use:   "generate-profile",
	Short: "Generate an AWS temporary profile",
	Long: `Generate an AWS temporary profile by assuming a role.
Supports both MFA and non-MFA authentication methods.
The temporary profile can then be used with the AWS CLI.

Examples:
  # Using default profile with MFA
  awsomecreds generate-profile -r arn:aws:iam::123456789012:role/my-role -m 123456 -n my-temp-profile

  # Using a specific source profile without MFA
  awsomecreds generate-profile -s my-source-profile -r arn:aws:iam::123456789012:role/my-role -n my-temp-profile

  # Specifying region and duration
  awsomecreds generate-profile -r arn:aws:iam::123456789012:role/my-role -n my-temp-profile --region us-west-2 -d 7200`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateTempProfile(sourceProfile, roleArn, mfaToken, newProfile, region, duration)
	},
}

func init() {
	rootCmd.AddCommand(generateProfileCmd)

	// Define flags for the generate-profile command
	generateProfileCmd.Flags().StringVarP(&sourceProfile, "source-profile", "s", "", "The AWS profile to use as the source for authentication (optional, uses default profile if not specified)")
	generateProfileCmd.Flags().StringVarP(&roleArn, "role-arn", "r", "", "The ARN of the role to assume (required)")
	generateProfileCmd.Flags().StringVarP(&mfaToken, "mfa-token", "m", "", "The MFA token code (optional, required only if the role requires MFA)")
	generateProfileCmd.Flags().StringVarP(&newProfile, "new-profile", "n", "", "The name for the new profile to create (required)")
	generateProfileCmd.Flags().StringVarP(&region, "region", "", "", "AWS region to use for the new profile (optional, uses source profile's region if not specified)")
	generateProfileCmd.Flags().IntVarP(&duration, "duration", "d", 3600, "Session duration in seconds (900-43200, default is 3600/1 hour)")

	// Mark required flags
	generateProfileCmd.MarkFlagRequired("role-arn")
	generateProfileCmd.MarkFlagRequired("new-profile")
}
