variable "aws_region" {
  description = "The AWS region to deploy to"
  type        = string
  default     = "ap-southeast-1"
}

variable "environment" {
  description = "The logical environment name (e.g., dev, prod)"
  type        = string

  validation {
    condition     = contains(["dev", "prod"], var.environment)
    error_message = "The environment variable must be strictly one of 'dev' or 'prod'."
  }
}

variable "aws_profile" {
  description = "AWS CLI profile to use for local deployment"
  type        = string
  default     = null
}

variable "project" {
  description = "Project name"
  type        = string
  default     = "mangalodon"
}

variable "services" {
  description = "List of services that require image build"
  type        = list(string)
}