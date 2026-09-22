environment   = "prod"
vault_address = "https://192.168.1.10:8200"
granularity   = "service"

services = {
  frontend = {
    allow_secret_paths = ["kv/data/mangalodon/prod/frontend/*"]
  }
  apiservice = {
    allow_secret_paths = ["kv/data/mangalodon/prod/apiservice/*", "kv/data/mangalodon/prod/postgres/*", "kv/data/mangalodon/prod/rabbitmq/*"]
  }
  scrapeservice = {
    allow_secret_paths = ["kv/data/mangalodon/prod/scrapeservice/*", "kv/data/mangalodon/prod/postgres/*", "kv/data/mangalodon/prod/rabbitmq/*"]
  }
  scrapescheduler = {
    allow_secret_paths = []
  }
  postgres = {
    allow_secret_paths = ["kv/data/mangalodon/prod/postgres/*"]
  }
  rabbitmq = {
    allow_secret_paths    = ["kv/data/mangalodon/prod/rabbitmq/*"]
    service_account_names = ["rabbitmq-server", "messaging-topology-operator"]
  }
}
