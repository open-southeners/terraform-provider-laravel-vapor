terraform {
  required_providers {
    laravelvapor = {
      source  = "terraform.local/local/laravelvapor"
      version = "1.0.0"
    }
  }
}

provider "laravelvapor" {
  # Token set at LARAVEL_VAPOR_TOKEN envvar
}

data "laravelvapor_account" "me" {}

output "test_account_id" {
  value = data.laravelvapor_account.me.id
}
