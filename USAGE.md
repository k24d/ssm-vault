# Usage

- [Installation](#installation)
- [Configuration](#configuration)
  - [Using aws-vault for development](#using-aws-vault-for-development)
- [Listing parameters](#listing-parameters)
- [Reading parameters](#reading-parameters)
  - [Clipboard](#clipboard)
  - [Multiline values](#multiline-values)
- [Writing parameters](#writing-parameters)
  - [Overwrite values](#overwrite-values)
- [Renaming parameters](#renaming-parameters)
- [Deleting parameters](#deleting-parameters)
- [Rendering template](#rendering-template)
- [Running command with environment variables](#running-command-with-environment-variables)
  - [Overwrite environment variables](#overwrite-environment-variables)

## Installation

### Homebrew (macOS)

```bash
$ brew tap k24d/ssm-vault
$ brew install ssm-vault
```

### Pre-compiled binaries (darwin, freebsd, linux, and windows)

- https://github.com/k24d/ssm-vault/releases

### Build from source

```bash
$ VERSION=`git describe --tags --candidates=1 --dirty`
$ go build -ldflags="-X main.Version=$VERSION -s -w" -trimpath -o ssm-vault
```

## Configuration

SSM Vault uses "AWS SDK for Go" with its default credential provider chain.  Please follow the configuration guide of AWS SDK for Go:

- [Configuring the AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-the-region)

Unlike AWS CLI, SSM Vault does *NOT* recognize `~/.aws/config`.  On local machines, you need to set environment variables (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`) or create a shared credential file `~/.aws/credentials`:

```
[dev]
aws_access_key_id=...
aws_secret_access_key=...
```

Then, you can run `ssm-vault` as follows:

```bash
$ export AWS_REGION=us-east-1
$ export AWS_PROFLIE=dev
$ ssm-vault ...

# or

$ AWS_REGION=us-east-1 AWS_PROFLIE=dev ssm-vault ...
```

Don't forget to set `AWS_REGION`.  Otherwise, you'll see an error like this:

```bash
$ AWS_PROFLIE=dev ssm-vault ...
2020/01/14 14:48:38 ERROR: unable to resolve endpoint for service "ssm", region "", err: UnknownEndpointError: could not resolve endpoint
	partition: "aws", service: "ssm", region: "", known: [ap-northeast-1 ap-southeast-1 ap-southeast-2 me-south-1 sa-east-1 us-west-1 eu-central-1 eu-north-1 us-east-1 ap-east-1 ap-northeast-2 ap-south-1 eu-west-2 eu-west-3 us-east-2 ca-central-1 eu-west-1 us-west-2]
ssm-vault: error: MissingRegion: could not find region configuration
exit status 1
```

### Using aws-vault for development

If you use IAM roles for engineers (recommended), you can use [aws-vault](https://github.com/99designs/aws-vault) to get things done easily.  Just create `~/.aws/config` in the same way as AWS CLI:

```
[profile dev]
region = us-east-1
role_arn = arn:aws:iam::********:role/dev
source_profile = default
```

Then, run `ssm-vault` through `aws-vault`:

```bash
# AWS access keys are provided by aws-vault
$ aws-vault exec dev -- ssm-vault ...

or

# Run a metadata server as a separate process
$ aws-vault exec dev --server
...

# Get access keys from the metadata server
$ export AWS_REGION=us-east-1
$ ssm-vault ...
```

## Listing parameters

Simple parameter list by `ssm-vault list`(alias `ls`):

```bash
$ ssm-vault ls
/app/dev/DB_PASSWORD
/app/dev/DB_USERNAME

$ ssm-vault ls --format json
{
  KeyId: "alias/aws/ssm",
  LastModifiedDate: 2020-01-14 03:59:08 +0000 UTC,
  LastModifiedUser: "arn:aws:sts::********:assumed-role/dev/1234...",
  Name: "/app/dev/DB_PASSWORD",
  Policies: [],
  Tier: "Standard",
  Type: "SecureString",
  Version: 1
}
...
```

Tree list by `ssm-vault tree`:

```bash
$ ssm-vault tree
.
└── /app/
    └── dev/
        ├── DB_PASSWORD🔐 (alias/aws/ssm)
        └── DB_USERNAME
```

## Reading parameters

Simple read by `ssm-vault read`(alias `get`):

```bash
$ ssm-vault read /app/dev/DB_PASSWORD
MY-SUPER-SECRET

$ echo "Password: [`ssm-vault read /app/dev/DB_PASSWORD`]"
Password: [MY-SUPER-SECRET]
```

### Clipboard

Copy to clipboard by `ssm-vault clipboard`(alias `c`):

```bash
$ ssm-vault c /app/dev/DB_PASSWORD
Copied to clipboard: /app/dev/DB_PASSWORD
```

### Multiline values

Read/write multiline secret files (e.g., X.509 private keys):

```bash
$ cat pkey.pem
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBABKCAQEAqFKPHQkB3LYqTV4+G2yW06HpXnvtCi+imTgmboKNOdisZmn8
GvJaSFulyf3YIMMuRwAn/KTFYAEcJ3Tsm3zoENRhnEmcC8JLGPL8nHGQjNxPpT5Q
...

# Store secret data from a file
$ ssm-vault write /app/dev/private_key < pkey.pem

# Save to a file with mode 0600
$ ssm-vault read /app/dev/private_key -o out.pem -m 0600

$ cat out.pem
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBABKCAQEAqFKPHQkB3LYqTV4+G2yW06HpXnvtCi+imTgmboKNOdisZmn8
GvJaSFulyf3YIMMuRwAn/KTFYAEcJ3Tsm3zoENRhnEmcC8JLGPL8nHGQjNxPpT5Q
...
```

## Writing parameters

Simple write by `ssm-vault write`(alias `put`):

```bash
# Store a secret value as SecureString (encrypted)
$ ssm-vault write /app/dev/DB_PASSWORD
Enter secret: ********

# Get a secret value from stdin
# DON'T DO THIS: the command line is stored in your shell history!
$ echo -n "MY-SUPER-SECRET" | ssm-vault write /app/dev/DB_PASSWORD
```

Writing plain text values:

```bash
# Store a text value as String (plain text)
$ ssm-vault write /app/dev/DB_USERNAME -s
Enter text: dbuser

# Get a text value from stdin
$ echo -n "dbuser" | ssm-vault write /app/dev/DB_USERNAME -s
```

### Overwrite values

Use `ssm-vault write -f` to overwrite an existing parameter value:

```bash
$ ssm-vault write /app/dev/DB_PASSWORD
Enter secret: ********
ssm-vault: error: ParameterAlreadyExists: The parameter already exists. To overwrite this value, set the overwrite option in the request to true.
	status code: 400, request id: 5aaabe54-d0b4-4206-be2a-6a0fad4fdb58

$ ssm-vault write -f /app/dev/DB_PASSWORD
Enter secret: ********
```

## Renaming parameters

> :warning: AWS SSM Parameter Store does not support renaming.  SSM Vault just copies parameters and deletes old ones.  See the limitations below.

Rename by `ssm-vault rename`(alias `mv`):

```bash
# Rename a single parameter (exact match)
$ ssm-vault mv /app/dev/DB_USERNAME /app/dev/DB_USER
/app/dev/DB_USERNAME -> /app/dev/DB_USER

# Rename multiple parameters (prefix match)
$ ssm-vault mv /app/dev /app/staging
/app/dev/DB_PASSWORD -> /app/staging/DB_PASSWORD
/app/dev/DB_USER -> /app/staging/DB_USER
```

> **Limitations of parameter renames**
>
> - Instead of renaming, parameters are copied and deleted.
> - Currently the following attributes are not copied:
>   - Description
>   - KeyId
>   - Policies
>   - Tags
>   - Tier
>   - Versions
> - Historical values (versions) will be lost.  Only the latest value remains.
>
> If you need to keep old values, please copy parameters instead of renaming:
>
> ```bash
> $ ssm-vault read /app/dev/DB_USERNAME | ssm-vault write /app/dev/DB_USER
> ```

## Deleting parameters

Delete by `ssm-vault delete`(alias `rm`):

```bash
$ ssm-vault rm /app/dev/DB_USERNAME
Are you sure to delete /app/dev/DB_USERNAME (y/N)? y

# Delete without confirmation
$ ssm-vault rm -f /app/dev/DB_USERNAME
```

## Rendering template

Embed parameter values in template ([syntax](https://golang.org/pkg/text/template/)) by `ssm-vault render`:

```bash
$ cat my_cnf.template
[client]
user={{aws_ssm_parameter "/app/dev/DB_USERNAME"}}
password={{aws_ssm_parameter "/app/dev/DB_PASSWORD"}}

# Output to a file with mode 0600
$ ssm-vault render my_cnf.template -o ~/.my.cnf -m 0600

$ cat ~/.my.cnf
[client]
user=dbuser
password=MY-SUPER-SECRET
```

Template string from stdin:

```bash
$ ssm-vault render -o ~/.my.cnf <<EOT
[client]
user={{aws_ssm_parameter "/app/dev/DB_USERNAME"}}
password={{aws_ssm_parameter "/app/dev/DB_PASSWORD"}}
EOT
```

Truncate path prefix by `--path`(or `-p`):

```bash
$ ssm-vault render -p /app/dev -o ~/.my.cnf <<EOT
[client]
user={{aws_ssm_parameter "DB_USERNAME"}}
password={{aws_ssm_parameter "DB_PASSWORD"}}
EOT
```

## Running command with environment variables

> :warning: This command should be used with extra care.  There is plenty of discussion about storing secrets in environment variables, like "[Environment Variables Considered Harmful for Your Secrets - Hacker News](https://news.ycombinator.com/item?id=8826024)".  On the other hand, having secrets in environment variables is sometimes convenient, especially when you run Docker containers for development.  It's up to you whether or not to use this feature.

`ssm-vault exec` runs a given command after exposing parameter values as environment variables:

```bash
# Any symbols in parameter names are converted to "_"
$ ssm-vault exec -- env | grep DB_
APP_DEV_DB_PASSWORD=MY-SUPER-SECRET
APP_DEV_DB_USERNAME=dbuser

# Truncate path prefix
$ ssm-vault exec -p /app/dev -- env | grep DB_
DB_PASSWORD=MY-SUPER-SECRET
DB_USERNAME=dbuser
```

Use `--safe` if you wish to expose only plain text values:

```bash
$ ssm-vault exec -p /app/dev --safe -- env | grep DB_
DB_USERNAME=dbuser
```

### Overwrite environment variables

By default, `ssm-vault exec` does not overwrite existing environment variables.  Set `--overwrite`(or `-f`) to change the behavior:

```bash
$ export DB_USERNAME=newuser

$ ssm-vault read /app/dev/DB_USERNAME
dbuser

# No overwrite for existing environment variables
$ ssm-vault exec -p /app/dev -- env | grep DB_USERNAME
DB_USERNAME=newuser

# Overwrite environment variables
$ ssm-vault exec -f -p /app/dev -- env | grep DB_USERNAME
DB_USERNAME=dbuser
```
