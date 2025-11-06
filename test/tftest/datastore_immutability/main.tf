terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "1.0.0"
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

variable "s3_endpoint_id" {
  type    = string
  default = "test-s3-immut"
}

variable "datastore_name" {
  type = string
}

variable "datastore_path" {
  type = string
}

variable "s3_bucket" {
  type = string
}

variable "comment" {
  type    = string
  default = "Test S3 datastore"
}

# Create S3 endpoint first
resource "pbs_s3_endpoint" "test" {
  id         = var.s3_endpoint_id
  endpoint   = "https://s3.amazonaws.com"
  region     = "us-east-1"
  access_key = "test-access-key"
  secret_key = "test-secret-key"
}

# Create S3-backed datastore
resource "pbs_datastore" "s3_test" {
  name      = var.datastore_name
  path      = var.datastore_path
  s3_client = pbs_s3_endpoint.test.id
  s3_bucket = var.s3_bucket
  comment   = var.comment
  
  depends_on = [pbs_s3_endpoint.test]
}

output "datastore_name" {
  value = pbs_datastore.s3_test.name
}

output "datastore_path" {
  value = pbs_datastore.s3_test.path
}

output "datastore_s3_client" {
  value = pbs_datastore.s3_test.s3_client
}

output "datastore_s3_bucket" {
  value = pbs_datastore.s3_test.s3_bucket
}

output "datastore_comment" {
  value = pbs_datastore.s3_test.comment
}

output "s3_endpoint_id" {
  value = pbs_s3_endpoint.test.id
}
