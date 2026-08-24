environment        = "prod"
vault_address      = "https://192.168.1.10:8200"
vault_ca_cert_file = "vault-root-ca.crt"
granularity        = "service"

services = {
  frontend = ["kv/data/mangalodon/prod/frontend/*"]
  api      = ["kv/data/mangalodon/prod/api/*", "kv/data/mangalodon/prod/postgres/*"]
  scraper  = ["kv/data/mangalodon/prod/scraper/*", "kv/data/mangalodon/prod/api/*", "kv/data/mangalodon/prod/postgres/*"]
  postgres = []
  rabbitmq = []
  cronjob  = []
}
