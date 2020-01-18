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
