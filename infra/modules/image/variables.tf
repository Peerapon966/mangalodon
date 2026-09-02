variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "services" {
  description = "Map of services and version tags that require image build"
  type        = map(string)
}

variable "aws_profile" {
  description = "AWS profile for local-exec commands"
  type        = string
  default     = null
}
