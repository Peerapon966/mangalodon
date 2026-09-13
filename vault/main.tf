locals {
  # For service granularity, only create roles for services with non-empty permission arrays
  # For environment granularity, create a single entry keyed by the environment name
  services = (
    var.granularity == "environment"
    ? { (var.environment) = object({
      allow_secret_paths = ["kv/data/${var.project}/${var.environment}/*"]
    }) }
    : { for k, v in var.services : k => v if length(v.allow_secret_paths) > 0 }
  )
}

resource "vault_policy" "read_secrets_policy" {
  for_each = local.services

  name = (
    var.granularity == "environment"
    ? "read-${var.project}-${var.environment}-secrets"
    : "read-${var.project}-${each.key}-${var.environment}-secrets"
  )

  policy = join("", concat(
    [<<EOT
path "sys/events/subscribe/*" {
  capabilities = ["read"]
}

EOT
    ],
    [for path in each.value.allow_secret_paths : <<EOT
path "${path}" {
  capabilities = ["read", "list", "subscribe"]
  subscribe_event_types = ["*"]
}

EOT
    ]
  ))
}

resource "vault_kubernetes_auth_backend_role" "read_secrets_role" {
  for_each = local.services

  backend = "kubernetes"
  role_name = (
    var.granularity == "environment"
    ? "${var.project}-${var.environment}"
    : "${var.project}-${each.key}-${var.environment}"
  )
  bound_service_account_names      = [var.granularity == "environment" ? "${var.project}" : coalesce(each.value.service_account_name, each.key)]
  bound_service_account_namespaces = ["${var.project}-${var.environment}"]
  token_policies                   = [vault_policy.read_secrets_policy[each.key].name]
  token_ttl                        = 3600
}
