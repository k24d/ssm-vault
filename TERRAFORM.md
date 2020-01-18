# SSM Parameter Store from Terraform

- [Basic configurations](#basic-configurations)
- [Storing parameter values](#storing-parameter-values)
  - [Using parameter values](#using-parameter-values)
  - [Using stored secrets](#using-stored-secrets)
- [Best practices](#best-practices)
  - [ID in the code](#id-in-the-code)

## Basic configurations

In this document, we organize our Terraform repository into the following file hierarchy:

```
app/
  dev/
    main.tf
    parameters.tf
  prod/
    main.tf
    parameters.tf
```

(See [examples/terraform/](examples/terraform/) for full example.)

There are two application environments, "dev" and "prod", and they have separate directories.  Each environment has a dedicated AWS account.  We use AssumeRole to switch to each account.  Let's create `main.tf` for each environment:

```bash
# main.tf

terraform {
  backend "s3" {
    bucket  = "my-terraform-dev-state"
    key     = "terraform.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}

provider "aws" {
  region = "us-east-1"
}
```

In this example, we use the backend "s3" for state files.  **Terraform stores [sensitive data in state files](https://www.terraform.io/docs/state/sensitive-data.html)**, so access to this bucket must be restricted.

The provider "aws" specifies the region where we store parameters.

Terraform can be initialized as follows, using `aws-vault` to select a role:

```bash
$ cd app/dev

$ aws-vault exec dev -- terraform init
```

## Storing parameter values

Parameters are created by Terraform resource "aws_ssm_parameter".  You can store both plan text values (String) and encrypted values (SecureString):

```bash
# parameters.tf

# Store a username ("dbuser") as String (plain text)
resource "aws_ssm_parameter" "db_username" {
  name  = "/org/app/dev/DB_USERNAME"
  value = "dbuser"
  type  = "String"
}

# Generate a random password (16 characters)
resource "random_password" "db_password" {
  length           = 16
  special          = true
  override_special = "_%@"
}

# Store the password as SecureString (encrypted)
resource "aws_ssm_parameter" "db_password" {
  name  = "/org/app/dev/DB_PASSWORD"
  value = random_password.db_password.result
  type  = "SecureString"
}
```

Run `terraform apply` to apply changes:

```bash
$ aws-vault exec dev -- terraform apply
...
Plan: 3 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

random_password.db_password: Creating...
random_password.db_password: Creation complete after 0s [id=none]
aws_ssm_parameter.db_password: Creating...
aws_ssm_parameter.db_username: Creating...
aws_ssm_parameter.db_username: Creation complete after 0s [id=/org/app/dev/DB_USERNAME]
aws_ssm_parameter.db_password: Creation complete after 0s [id=/org/app/dev/DB_PASSWORD]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.
```

### Using parameter values

If you manage other resources, such as EC2 and RDS, use stored parameter values in the rest of your Terraform files:

```bash
# Create an RDS instance using stored parameter values
resource "aws_db_instance" "db" {
  instance_class = "db.t3.micro"
  ...
  username       = aws_ssm_parameter.db_username.value
  password       = aws_ssm_parameter.db_password.value
}
```

### Using stored secrets

In many cases, you already have some secrets and cannot generate random passwords.  You don't want to write them in plain text:

```bash
# THIS IS NOT SECURE!
resource "aws_ssm_parameter" "db_password" {
  name  = "/org/app/dev/DB_PASSWORD"
  value = "my-secret-password"
  type  = "SecureString"
}
```

Instead of using Terraform, I would just use `ssm-vault write` to store secrets:

```
$ ssm-vault write /org/app/dev/DB_PASSWORD
Enter secret: ********
```

Then, read the secrets from Terraform:

```
# Read a secret value from Parameter Store
data "aws_ssm_parameter" "db_password" {
  name  = "/org/app/dev/DB_PASSWORD"
  type  = "SecureString"
}

# Refer to the value
resource "aws_db_instance" "db" {
  instance_class = "db.t3.micro"
  ...
  username       = aws_ssm_parameter.db_username.value
  password       = data.aws_ssm_parameter.db_password.value
}
```

In this way, you can avoid storing secrets in your Terraform repository, while keeping parameter names in the code.
