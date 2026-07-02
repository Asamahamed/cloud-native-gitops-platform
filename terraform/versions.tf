terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
  }

  # Remote state — uncomment and configure for team use / CI.
  # backend "s3" {
  #   bucket         = "asamahamed-tfstate"
  #   key            = "gitops-platform/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "cloud-native-gitops-platform"
      Environment = var.environment
      ManagedBy   = "Terraform"
      Owner       = "Asam Ahamed"
    }
  }
}
