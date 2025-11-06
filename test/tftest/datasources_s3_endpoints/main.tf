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

# Create multiple S3 endpoints
resource "pbs_s3_endpoint" "test1" {
  id         = "tftest-s3-list-1"
  endpoint   = "s3.us-east-1.amazonaws.com"
  region     = "us-east-1"
  access_key = "test-key-1"
  secret_key = "test-secret-1"
}

resource "pbs_s3_endpoint" "test2" {
  id         = "tftest-s3-list-2"
  endpoint   = "s3.us-west-2.amazonaws.com"
  region     = "us-west-2"
  access_key = "test-key-2"
  secret_key = "test-secret-2"
}

# List all S3 endpoints
data "pbs_s3_endpoints" "all" {
  depends_on = [
    pbs_s3_endpoint.test1,
    pbs_s3_endpoint.test2
  ]
}

output "endpoint_count" {
  value = length(data.pbs_s3_endpoints.all.endpoints)
}

output "endpoint_ids" {
  value = [for e in data.pbs_s3_endpoints.all.endpoints : e.id]
}
