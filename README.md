# SSM Vault

SSM Vault is a lightweight tool for using [AWS Systems Manager (SSM) Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) from CLI.

You can use SSM Vault to store credentials in AWS and retrieve them during development or in production.  It allows you to show parameters in tree views, embed parameters in configuration files, and expose parameters as environment variables.

## Documentation

- [Usage](USAGE.md)
- [SSM Parameter Store from Docker](DOCKER.md)
- [SSM Parameter Store from Terraform](TERRAFORM.md)

## Installation

Use Homebrew on macOS.  See "[Installation](USAGE.md#installation)" for other environments.

```
$ brew tap k24d/ssm-vault
$ brew install ssm-vault
```

## Basic Usage

```bash
# Store a secret value as SecureString (encrypted)
$ ssm-vault write /app/dev/DB_PASSWORD
Enter secret: ********

# Store a text value as String (plain text)
$ ssm-vault write /app/dev/DB_USERNAME -s
Enter text: dbuser

# Show parameters in tree format
$ ssm-vault tree
.
└── /app/
    └── dev/
        ├── DB_PASSWORD🔐 (alias/aws/ssm)
        └── DB_USERNAME

# Copy a value to clipboard
$ ssm-vault c /app/dev/DB_PASSWORD
Copied to clipboard: /app/dev/DB_PASSWORD

# Get a value in shell scripts
$ export MYSQL_PWD=`ssm-vault read /app/dev/DB_PASSWORD`

# Render template and output to a file (mode 0600 by default)
$ ssm-vault render -o ~/.my.cnf <<EOT
[client]
user={{aws_ssm_parameter "/app/dev/DB_USERNAME"}}
password={{aws_ssm_parameter "/app/dev/DB_PASSWORD"}}
EOT

$ cat ~/.my.cnf
[client]
user=dbuser
password=MY-SUPER-SECRET

# Execute a command with environment variables
$ ssm-vault exec -p /app/dev -- env | grep DB_
DB_PASSWORD=MY-SUPER-SECRET
DB_USERNAME=dbuser
```

See "[Usage](USAGE.md)" for detailed usage.  For Docker users, "[SSM Parameter Store from Docker](DOCKER.md)" describes how to use SSM Vault from Docker containers.

## Infrastructure as Code

When you organize application parameters and secrets as part of "Infrastructure as Code", I recommend that you use [Terraform](https://www.terraform.io) or similar tools instead of using SSM Vault manually.  "[SSM Parameter Store from Terraform](TERRAFORM.md)" briefly describes how to use Terraform for parameter management.
