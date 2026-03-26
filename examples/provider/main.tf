terraform {
  required_providers {
    luckperms = {
      source = "digitaldrugstech/luckperms"
    }
  }
}

provider "luckperms" {
  base_url = "http://localhost:8080"
  api_key  = var.luckperms_api_key
  timeout  = 30
}

variable "luckperms_api_key" {
  description = "API key for LuckPerms"
  type        = string
  sensitive   = true
}
