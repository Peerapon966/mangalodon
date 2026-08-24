terraform {
  required_version = ">= 1.5.0"

  backend "s3" {}

  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "~> 5.0"
    }
  }
}

provider "vault" {
  address      = var.vault_address
  ca_cert_file = var.vault_ca_cert_file
}
