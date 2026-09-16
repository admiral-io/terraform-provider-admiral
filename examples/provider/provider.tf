terraform {
  required_providers {
    admiral = {
      source  = "admiral-io/admiral"
      version = "~> 0.1"
    }
  }
}

provider "admiral" {
  # server  = "api.admiral.io:443"  # optional; or set ADMIRAL_SERVER
  # api_key = "admp_..."            # or set ADMIRAL_API_KEY (recommended)
}
