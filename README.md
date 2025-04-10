# AWSomeCreds

[![AWSomeCreds CI](https://github.com/coreyculler/awsomecreds/actions/workflows/go-test.yml/badge.svg)](https://github.com/coreyculler/awsomecreds/actions/workflows/go-test.yml)

AWSomeCreds is a CLI tool that generates temporary AWS credentials using AWS STS and sets them using aws configure. It allows you to assume roles with or without MFA authentication and create temporary profiles for AWS CLI usage.

## Features

- Assume AWS IAM roles with or without MFA authentication
- Create temporary AWS CLI profiles with the assumed credentials
- Export temporary credentials as environment variables in your shell
- Output credentials in JSON format
- Configurable session duration
- Support for custom AWS regions
- Works with default or named AWS profiles as the source

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/coreyculler/awsomecreds.git
cd awsomecreds

# Build and install
make install
```

### Using Go

```bash
go install github.com/coreyculler/awsomecreds@latest
```

## Usage

```bash
# Generate a new AWS profile with temporary credentials
awsomecreds generate-profile [flags]

# Generate temporary AWS credentials and export to environment variables
awsomecreds generate [flags]
```

### Commands

#### generate-profile

Generate a temporary AWS credential profile that can be used with AWS CLI and SDKs.

##### Flags

- `--source-profile`, `-s`: The AWS profile to use as the source for authentication (optional, uses default profile if not specified)
- `--role-arn`, `-r`: The ARN of the role to assume (required)
- `--mfa-token`, `-m`: The MFA token code (optional, required only if the role requires MFA)
- `--new-profile`, `-n`: The name for the new profile to create (required)
- `--region`: AWS region to use for the new profile (optional, uses source profile's region if not specified)
- `--duration`, `-d`: Session duration in seconds (900-43200, default is 3600/1 hour)

##### Examples

###### Using default profile with MFA
```bash
awsomecreds generate-profile -r arn:aws:iam::123456789012:role/my-role -m 123456 -n my-temp-profile
```

###### Using a specific source profile without MFA
```bash
awsomecreds generate-profile -s my-source-profile -r arn:aws:iam::123456789012:role/my-role -n my-temp-profile
```

###### Specifying region and duration
```bash
awsomecreds generate-profile -r arn:aws:iam::123456789012:role/my-role -n my-temp-profile --region us-west-2 -d 7200
```

After running the command, you can use the temporary profile with AWS CLI:

```bash
aws --profile my-temp-profile s3 ls
```

#### generate

Generate temporary AWS credentials and output them to stdout for setting as environment variables in the parent shell.

##### Flags

- `--source-profile`, `-s`: The AWS profile to use as the source for authentication (optional, uses default profile if not specified)
- `--role-arn`, `-r`: The ARN of the role to assume (required)
- `--mfa-token`, `-m`: The MFA token code (optional, required only if the role requires MFA)
- `--region`: AWS region to use for the new profile (optional, uses source profile's region if not specified)
- `--duration`, `-d`: Session duration in seconds (900-43200, default is 3600/1 hour)
- `--output`, `-o`: Output format: 'shell' for shell environment variables or 'json' for JSON format (default is 'shell')

##### Examples

###### Using default profile with MFA (shell variables)
```bash
eval $(awsomecreds generate -r arn:aws:iam::123456789012:role/my-role -m 123456)
```

###### Using a specific source profile without MFA
```bash
eval $(awsomecreds generate -s my-source-profile -r arn:aws:iam::123456789012:role/my-role)
```

###### Specifying region and duration
```bash
eval $(awsomecreds generate -r arn:aws:iam::123456789012:role/my-role --region us-west-2 -d 7200)
```

###### Output credentials in JSON format
```bash
awsomecreds generate -r arn:aws:iam::123456789012:role/my-role -o json
```

After running the command with `eval $(...)`, the AWS environment variables will be set in your current shell session:

```bash
# Variables set in your shell
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_SESSION_TOKEN=...
AWS_REGION=... (if region was specified)
AWS_DEFAULT_REGION=... (if region was specified)
AWS_CREDENTIAL_EXPIRATION=...
```

## Prerequisites

- AWS CLI installed and configured
- AWS IAM user with permissions to assume the target role
- MFA device configured (if required by the role)

## Development

### Requirements

- Go 1.21 or higher
- AWS CLI

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run integration tests (requires AWS credentials)
make integration-test
```

See the Makefile for additional commands.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request