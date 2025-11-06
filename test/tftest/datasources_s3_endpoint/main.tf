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
  id         = "tftest-s3-ds"
  endpoint   = "s3.us-east-1.amazonaws.com"
  region     = "us-east-1"
  access_key = "test-access-key"
  secret_key = "test-secret-key"
}

# Read it via data source
data "pbs_s3_endpoint" "test" {
  id = pbs_s3_endpoint.test.id
}

output "resource_id" {
  value = pbs_s3_endpoint.test.id
}

output "datasource_id" {
  value = data.pbs_s3_endpoint.test.id
}

output "datasource_endpoint" {
  value = data.pbs_s3_endpoint.test.endpoint
}
