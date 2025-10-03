resource "pbs_s3_endpoint" "example" {
  id = "my-s3-endpoint"
  
  # S3 configuration
  access_key = "your-access-key"
  secret_key = "your-secret-key"
  endpoint   = "https://s3.amazonaws.com"  # S3 service endpoint
  region     = "us-west-2"
  
  # Optional settings
  fingerprint = ""  # SSL certificate fingerprint if needed
  port        = 443
  path_style  = false
}