terraform {
  required_providers {
    admiral = {
      source  = "admiral-io/admiral"
      version = "~> 0.1"
    }
  }
}

provider "admiral" {
  host  = var.admiral_host
  token = var.admiral_token
}
