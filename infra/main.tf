module "image" {
  source = "./modules/image"

  project          = var.project
  environment      = var.environment
  aws_profile      = var.aws_profile
  services         = var.services
  is_manual_deploy = var.is_manual_deploy
}
