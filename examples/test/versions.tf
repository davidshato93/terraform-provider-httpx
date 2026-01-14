terraform {
  required_version = ">= 1.0"
  
  required_providers {
    httpx = {
      source  = "registry.terraform.io/davidshato/httpx"
      # Version constraint is ignored for dev_overrides
      version = ">= 0"
    }
  }
}

