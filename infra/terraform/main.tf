terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.4"
    }
  }

  # Note: Backend should be configured for state storage (e.g., S3) in a real environment.
  # For now, we use local state.
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS Region to deploy to"
  type        = string
  default     = "ap-northeast-1"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "wakaba"
}
