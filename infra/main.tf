module "image" {
  source = "./modules/image"

  project     = var.project
  environment = var.environment
  aws_profile = var.aws_profile
  services    = var.services
}
