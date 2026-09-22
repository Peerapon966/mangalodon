environment   = "dev"
vault_address = "https://192.168.1.10:8200"
granularity   = "service"

services = {
  frontend = {
    allow_secret_paths = ["kv/data/mangalodon/dev/frontend/*"]
  }
  apiservice = {
    allow_secret_paths = ["kv/data/mangalodon/dev/apiservice/*", "kv/data/mangalodon/dev/postgres/*", "kv/data/mangalodon/dev/rabbitmq/*"]
  }
  scrapeservice = {
    allow_secret_paths = ["kv/data/mangalodon/dev/scrapeservice/*", "kv/data/mangalodon/dev/postgres/*", "kv/data/mangalodon/dev/rabbitmq/*"]
  }
  scrapescheduler = {
    allow_secret_paths = []
  }
  postgres = {
    allow_secret_paths = ["kv/data/mangalodon/dev/postgres/*"]
  }
  rabbitmq = {
    allow_secret_paths         = ["kv/data/mangalodon/dev/rabbitmq/*"]
    service_account_names      = ["rabbitmq-server", "messaging-topology-operator"]
    service_account_namespaces = ["rabbitmq-system"]
  }
}
