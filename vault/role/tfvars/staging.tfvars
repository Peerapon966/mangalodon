environment        = "staging"
vault_address      = "https://192.168.1.10:8200"
vault_ca_cert_file = "vault-root-ca.crt"
granularity        = "environment"

services = {
  frontend = ["kv/data/mangalodon/staging/frontend/*"]
  api      = ["kv/data/mangalodon/staging/api/*", "kv/data/mangalodon/staging/postgres/*"]
  scraper  = ["kv/data/mangalodon/staging/scraper/*", "kv/data/mangalodon/staging/api/*", "kv/data/mangalodon/staging/postgres/*"]
  postgres = ["kv/data/mangalodon/staging/postgres/*"]
  rabbitmq = []
  cronjob  = []
}
