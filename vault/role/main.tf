locals {
  # For service granularity, only create roles for services with non-empty permission arrays
  # For environment granularity, create a single entry keyed by the environment name
  service_policies = (
    var.granularity == "service"
    ? { for k, v in var.services : k => v if length(v) > 0 }
    : { (var.environment) = ["kv/data/${var.project}/${var.environment}/*"] }
  )
}

resource "vault_policy" "read_secrets_policy" {
  for_each = local.service_policies

  name = (
    var.granularity == "environment"
    ? "read-${var.project}-${var.environment}-secrets"
    : "read-${var.project}-${each.key}-${var.environment}-secrets"
  )

  policy = join("", [
    for path in each.value : <<EOT
path "${path}" {
  capabilities = ["read", "list"]
}

EOT
  ])
}

resource "vault_kubernetes_auth_backend_role" "read_secrets_role" {
  for_each = local.service_policies

  backend = "kubernetes"

  bound_service_account_names = (
    var.granularity == "environment"
    ? ["${var.project}-secret-store"]
    : ["${var.project}-${each.key}-secret-store"]
  )

  bound_service_account_namespaces = (
    var.granularity == "environment"
    ? ["${var.project}-${var.environment}"]
    : ["${var.project}-${each.key}-${var.environment}"]
  )

  role_name = (
    var.granularity == "environment"
    ? "${var.project}-${var.environment}"
    : "${var.project}-${each.key}-${var.environment}"
  )

  token_policies = [vault_policy.read_secrets_policy[each.key].name]
  token_ttl      = 3600
}
