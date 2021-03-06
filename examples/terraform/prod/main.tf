terraform {
  backend "s3" {
    bucket  = "my-terraform-prod-state"
    key     = "terraform.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}

provider "aws" {
  region = "us-east-1"
}
