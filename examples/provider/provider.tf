terraform {
  required_providers {
    admiral = {
      source  = "admiral/admiral"
      version = "~> 1"
    }
  }
}

provider "admiral" {
  host  = var.admiral_host
  token = var.admiral_token
}