variable "project" {
  description = "Project name"
  type        = string
  default     = "mangalodon"
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "granularity" {
  description = "Granularity of vault roles and policies: 'environment' or 'service'"
  type        = string
  validation {
    condition     = contains(["environment", "service"], var.granularity)
    error_message = "The granularity must be either 'environment' or 'service'."
  }
}

variable "vault_address" {
  description = "Origin URL of the Vault server"
  type        = string
}

variable "vault_ca_cert_file" {
  description = "Path to a CA file on local disk (relative to the root module) that will be used to validate the certificate presented by the Vault server"
  type        = string
  default     = "assets/cacert/vault-root-ca.crt"
}

variable "services" {
  description = "Configuration for specific services and the exact Vault secret paths they can read"
  type = map(object({
    allow_secret_paths         = list(string)
    service_account_names      = optional(list(string))
    service_account_namespaces = optional(list(string))
  }))

  validation {
    condition     = toset(keys(var.services)) == toset(["frontend", "apiservice", "scrapeservice", "scrapescheduler", "postgres", "rabbitmq"])
    error_message = "var.services must contain exactly the following keys: frontend, apiservice, scrapeservice, scrapescheduler, postgres, rabbitmq."
  }
}
