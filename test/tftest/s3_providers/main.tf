terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "1.0.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

variable "pbs_endpoint" {
  type        = string
  description = "PBS server endpoint"
}

variable "pbs_username" {
  type        = string
  description = "PBS username"
}

variable "pbs_password" {
  type        = string
  description = "PBS password"
  sensitive   = true
}

variable "test_id" {
  type        = string
  description = "Unique test run identifier to avoid name conflicts"
  default     = "local"
}

provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

# Variables for S3 configuration
variable "s3_endpoint" {
  type        = string
  description = "S3 endpoint hostname (e.g., s3.us-west-2.amazonaws.com)"
}

variable "s3_region" {
  type        = string
  description = "S3 region"
}

variable "s3_access_key" {
  type        = string
  description = "S3 access key"
  sensitive   = true
}

variable "s3_secret_key" {
  type        = string
  description = "S3 secret key"
  sensitive   = true
}

variable "s3_bucket_name" {
  type        = string
  description = "S3 bucket name (must be unique)"
}

variable "s3_endpoint_id" {
  type        = string
  description = "PBS S3 endpoint ID"
}

variable "datastore_name" {
  type        = string
  description = "Name for the PBS datastore"
}

# AWS provider for bucket management
provider "aws" {
  region     = var.s3_region
  access_key = var.s3_access_key
  secret_key = var.s3_secret_key
}

# Create S3 bucket
resource "aws_s3_bucket" "test" {
  bucket        = var.s3_bucket_name
  force_destroy = true # Allow Terraform to delete non-empty bucket
  
  tags = {
    Name      = "PBS Test Bucket"
    TestID    = var.test_id
    ManagedBy = "Terraform"
    Purpose   = "PBS Provider Testing"
  }
}

# Wait for bucket to be available
resource "time_sleep" "bucket_creation" {
  depends_on      = [aws_s3_bucket.test]
  create_duration = "10s"
}

# Create PBS S3 endpoint
resource "pbs_s3_endpoint" "test" {
  depends_on = [time_sleep.bucket_creation]
  
  id         = var.s3_endpoint_id
  endpoint   = var.s3_endpoint
  region     = var.s3_region
  access_key = var.s3_access_key
  secret_key = var.s3_secret_key
  path_style = true # Required for PBS compatibility
}

# Create PBS datastore using the S3 bucket
resource "pbs_datastore" "test" {
  name      = var.datastore_name
  path      = "/datastore/${var.datastore_name}-cache"
  s3_client = pbs_s3_endpoint.test.id
  s3_bucket = aws_s3_bucket.test.bucket
  comment   = "Test S3 datastore for AWS provider"
}

# Outputs
output "bucket_name" {
  value = aws_s3_bucket.test.bucket
}

output "bucket_arn" {
  value = aws_s3_bucket.test.arn
}

output "s3_endpoint_id" {
  value = pbs_s3_endpoint.test.id
}

output "datastore_name" {
  value = pbs_datastore.test.name
}

output "datastore_path" {
  value = pbs_datastore.test.path
}
