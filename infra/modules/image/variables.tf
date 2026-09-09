variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "aws_profile" {
  description = "AWS profile for local-exec commands"
  type        = string
  default     = null
}

variable "services" {
  description = "Map of services and version tags that require image build"
  type        = map(string)
}

variable "is_manual_deploy" {
  description = "Whether to ignore the version tags specified in services and use hashes generated from source code as image tags instead"
  type        = bool
  default     = false
}
