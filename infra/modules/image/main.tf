data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

locals {
  sha1_hashes = { for s, _ in var.services : s => sha1(join("", [for f in fileset("../src/${s}", "**") : filesha1("../src/${s}/${f}")])) }
  profile_arg = var.aws_profile != null && var.aws_profile != "" ? "--profile ${var.aws_profile}" : ""
}

# 1. ECR Repository
resource "aws_ecr_repository" "repo" {
  for_each = var.services

  name                 = strcontains(var.environment, "prod") ? "${var.project}/${each.key}" : "${var.project}/${var.environment}/${each.key}"
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_ecr_lifecycle_policy" "policy" {
  for_each   = var.services
  repository = aws_ecr_repository.repo[each.key].name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 5 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 5
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# 2. Docker Build & Push (Null Resource)
resource "null_resource" "docker_build" {
  for_each = var.services
  triggers = {
    # Re-run when the code changes
    hash = local.sha1_hashes[each.key]
  }

  provisioner "local-exec" {
    command = <<EOF
docker buildx build --platform linux/amd64 --provenance=false --no-cache -t ${aws_ecr_repository.repo[each.key].repository_url}:${each.value} --build-arg ENV=${var.environment} --push ../src/${each.key}
EOF
  }
}
