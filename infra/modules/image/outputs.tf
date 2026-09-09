output "repositories" {
  description = "Map of ECR repository URLs for the built services"
  value       = { for s, _ in var.services : s => aws_ecr_repository.repo[s].repository_url }
}

output "image_tags" {
  description = "Map of image tags for the built services"
  value       = { for s, t in var.services : s => var.is_manual_deploy ? local.sha1_hashes[s] : t }
}

output "image_uris" {
  description = "Map of full ECR image URIs (repo_url:tag)"
  value       = { for s, t in var.services : s => "${aws_ecr_repository.repo[s].repository_url}:${var.is_manual_deploy ? local.sha1_hashes[s] : t}" }
}
