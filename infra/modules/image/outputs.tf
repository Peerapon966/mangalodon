output "repositories" {
  description = "Map of ECR repository URLs for the built services"
  value       = { for s in var.services : s => aws_ecr_repository.repo[s].repository_url }
}

output "image_tags" {
  description = "Map of computed image tags (SHA1) for the built services"
  value       = local.image_tags
}

output "image_uris" {
  description = "Map of full ECR image URIs (repo_url:tag)"
  value       = { for s in var.services : s => "${aws_ecr_repository.repo[s].repository_url}:${local.image_tags[s]}" }
}
