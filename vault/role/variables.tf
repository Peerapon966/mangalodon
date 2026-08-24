variable "project" {
  description = "Project name"
  type        = string
  default     = "mangalodon"
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vault_address" {
  description = "Origin URL of the Vault server"
  type        = string
}

variable "vault_ca_cert_file" {
  description = "Path to a file on local disk that will be used to validate the certificate presented by the Vault server"
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

variable "services" {
  description = "Configuration for specific services and the exact Vault secret paths they can read"
  type = object({
    frontend = list(string)
    api      = list(string)
    scraper  = list(string)
    postgres = list(string)
    rabbitmq = list(string)
    cronjob  = list(string)
  })
}
