terraform {
  required_providers {
    pbs = {
      source = "registry.terraform.io/micah/pbs"
    }
  }
}

provider "pbs" {
  endpoint  = "https://pbs.example.com:8007"
  api_token = "root@pam:api-token-name=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  # Alternatively use username/password:
  # username = "root@pam"
  # password = "your-password"
  
  # Optional settings
  insecure = false  # Set to true to skip TLS verification
  timeout  = 30     # API timeout in seconds
}