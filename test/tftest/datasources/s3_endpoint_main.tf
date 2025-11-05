terraform {
  required_providers {
    pbs = {
      source = "registry.terraform.io/micah/pbs"
    }
  }
}

provider "pbs" {
  endpoint = var.pbs_endpoint
  insecure = var.pbs_insecure
}

variable "pbs_endpoint" {
  type = string
}

variable "pbs_insecure" {
  type    = bool
  default = true
}

# Create an S3 endpoint
resource "pbs_s3_endpoint" "test" {
  name       = "tftest-s3-ds"
  provider   = "aws"
  endpoint   = "https://s3.amazonaws.com"
  region     = "us-east-1"
  access_key = "test-access-key"
  secret_key = "test-secret-key"
  comment    = "Test S3 endpoint for data source"
}

# Read it via data source
data "pbs_s3_endpoint" "test" {
  name = pbs_s3_endpoint.test.name
}

output "resource_name" {
  value = pbs_s3_endpoint.test.name
}

output "datasource_name" {
  value = data.pbs_s3_endpoint.test.name
}

output "datasource_provider" {
  value = data.pbs_s3_endpoint.test.provider
}
