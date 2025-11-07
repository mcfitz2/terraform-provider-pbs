terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "1.0.0"
    }
    aws = {
      source  = "hashicorp/aws"
      # Use v4 because v5's aws_s3_bucket tries to read CORS/versioning/logging
      # which fails on S3-compatible services (Backblaze, Scaleway) with 404 errors
      version = "~> 4.0"
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

# Variables for S3 provider configuration
variable "s3_provider_name" {
  type        = string
  description = "Name of the S3 provider (AWS, Backblaze, Scaleway)"
}

variable "s3_endpoint" {
  type        = string
  description = "S3 endpoint hostname (e.g., s3.us-west-2.amazonaws.com)"
}

variable "s3_region" {
  type        = string
  description = "S3 region (used for AWS provider configuration)"
}

variable "pbs_s3_region" {
  type        = string
  description = "S3 region for PBS endpoint configuration (optional, defaults to s3_region)"
  default     = ""
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

variable "s3_provider_quirks" {
  type        = list(string)
  description = "Provider-specific quirks (e.g., skip-if-none-match-header for Backblaze)"
  default     = []
}

variable "datastore_name" {
  type        = string
  description = "Name for the PBS datastore"
}

# AWS provider for bucket management
# We dynamically configure the region based on the s3_region variable
provider "aws" {
  region     = var.s3_region
  access_key = var.s3_access_key
  secret_key = var.s3_secret_key

  # Provider-specific endpoint configuration
  endpoints {
    s3 = var.s3_provider_name == "AWS" ? null : "https://${var.s3_endpoint}"
  }

  # Force path-style for non-AWS providers
  s3_use_path_style = var.s3_provider_name != "AWS"

  skip_credentials_validation = var.s3_provider_name != "AWS"
  skip_region_validation      = var.s3_provider_name != "AWS"
  skip_requesting_account_id  = var.s3_provider_name != "AWS"
  skip_metadata_api_check     = var.s3_provider_name != "AWS"
}

# Create S3 bucket for AWS
resource "aws_s3_bucket" "test" {
  count = var.s3_provider_name == "AWS" ? 1 : 0
  
  bucket        = var.s3_bucket_name
  force_destroy = true # Allow Terraform to delete non-empty bucket
  
  tags = {
    Name        = "PBS Test Bucket"
    TestID      = var.test_id
    Provider    = var.s3_provider_name
    ManagedBy   = "Terraform"
    Purpose     = "PBS Provider Testing"
  }
}

# Create S3 bucket for S3-compatible services (Backblaze, Scaleway)
# These services don't support many AWS S3 features, so we use lifecycle ignore_changes
# to prevent the AWS provider from failing when it tries to read unsupported configurations
resource "aws_s3_bucket" "test_compat" {
  count = var.s3_provider_name != "AWS" ? 1 : 0
  
  bucket        = var.s3_bucket_name
  force_destroy = true # Allow Terraform to delete non-empty bucket
  
  # Don't set tags - Backblaze and Scaleway don't support PutBucketTagging
}

# Local values to simplify resource references
locals {
  # Select the appropriate bucket resource based on provider type
  bucket = var.s3_provider_name == "AWS" ? aws_s3_bucket.test[0] : aws_s3_bucket.test_compat[0]
}

# Wait for bucket to be available
resource "time_sleep" "bucket_creation" {
  depends_on      = [aws_s3_bucket.test, aws_s3_bucket.test_compat]
  create_duration = "10s"
}

# Create PBS S3 endpoint
resource "pbs_s3_endpoint" "test" {
  depends_on = [time_sleep.bucket_creation]
  
  id              = var.s3_endpoint_id
  endpoint        = var.s3_endpoint
  # Use pbs_s3_region if provided, otherwise fall back to s3_region
  region          = var.pbs_s3_region != "" ? var.pbs_s3_region : var.s3_region
  access_key      = var.s3_access_key
  secret_key      = var.s3_secret_key
  path_style      = true # Required for PBS compatibility
  provider_quirks = var.s3_provider_quirks
}

# Create PBS datastore using the S3 bucket
resource "pbs_datastore" "test" {
  name      = var.datastore_name
  path      = "/datastore/${var.datastore_name}"
  s3_client = pbs_s3_endpoint.test.id
  s3_bucket = local.bucket.bucket
  comment   = "Test S3 datastore for ${var.s3_provider_name}"
  
  # Note: Implicit dependency via pbs_s3_endpoint.test.id reference
  # Terraform should create endpoint first, then datastore
  # and destroy in reverse order (datastore first, then endpoint)
}

# Outputs
output "bucket_name" {
  value = local.bucket.bucket
}

output "bucket_arn" {
  value = local.bucket.arn
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
