environment   = "staging"
vault_address = "https://192.168.1.10:8200"
granularity   = "service"

services = {
  frontend = {
    allow_secret_paths = ["kv/data/mangalodon/staging/frontend/*"]
  }
  apiservice = {
    allow_secret_paths = ["kv/data/mangalodon/staging/apiservice/*", "kv/data/mangalodon/staging/postgres/*", "kv/data/mangalodon/staging/rabbitmq/*"]
  }
  scrapeservice = {
    allow_secret_paths = ["kv/data/mangalodon/staging/scrapeservice/*", "kv/data/mangalodon/staging/postgres/*", "kv/data/mangalodon/staging/rabbitmq/*"]
  }
  scrapescheduler = {
    allow_secret_paths = []
  }
  postgres = {
    allow_secret_paths = ["kv/data/mangalodon/staging/postgres/*"]
  }
  rabbitmq = {
    allow_secret_paths    = ["kv/data/mangalodon/staging/rabbitmq/*"]
    service_account_names = ["rabbitmq-server", "messaging-topology-operator"]
  }
}
