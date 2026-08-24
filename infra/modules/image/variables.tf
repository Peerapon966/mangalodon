variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, prod)"
  type        = string
}

variable "services" {
  description = "List of services that require image build"
  type        = list(string)
}

variable "aws_profile" {
  description = "AWS profile for local-exec commands"
  type        = string
  default     = null
}
