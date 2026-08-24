output "ecr_repositories" {
  description = "The ECR repository URLs for each service"
  value       = module.image.repositories
}

output "image_tags" {
  description = "Map of computed image tags (SHA1) for the built services"
  value       = module.image.image_tags
}

output "image_uris" {
  description = "The complete ECR image URIs (with tags) to use in your Kubernetes deployments"
  value       = module.image.image_uris
}
