environment        = "dev"
vault_address      = "https://192.168.1.10:8200"
vault_ca_cert_file = "vault-root-ca.crt"
granularity        = "environment"

services = {
  frontend = ["kv/data/mangalodon/dev/frontend/*"]
  api      = ["kv/data/mangalodon/dev/api/*", "kv/data/mangalodon/dev/postgres/*"]
  scraper  = ["kv/data/mangalodon/dev/scraper/*", "kv/data/mangalodon/dev/api/*", "kv/data/mangalodon/dev/postgres/*"]
  postgres = []
  rabbitmq = []
  cronjob  = []
}
