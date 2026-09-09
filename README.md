# Mangalodon

Mangalodon is a purposely over-engineered microservices manga scraper and reader application.

The core task is really simple: scraping images from manga sites. But instead of writing a 10-line python script to do it, we're building a full distributed system around it. The goal here is to have a playground to practice deploying and managing complex infrastructure in Kubernetes and AWS, without getting bogged down by complicated application code.

## Tech Stack

- **Web UI (Frontend)**: React + Vite Single Page Application (SPA).
- **Core API (Backend)**: Hono + Postgres.js running on Node.js (formerly NestJS). Handles CRUD, database transactions, and real-time WebSocket events.
- **Message Broker**: RabbitMQ holds ad-hoc scraping tasks for the workers.
- **Scraping Workers**: Go routines that pull tasks, scrape image assets, download them to local disk, and trigger webhooks.
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana) with Filebeat sidecars pushing logs.
- **Database**: PostgreSQL (managed via CloudNativePG) storing all manga and job metadata.
- **Infrastructure as Code**: Terraform is used to provision AWS ECR repositories and securely build/push Docker images.
- **Deployment**: Fully containerized and orchestrated via Kubernetes Helm charts, strictly enforcing unprivileged containers and read-only filesystems.

## Running Locally

Run the deploy_dev script to deploy app to local Kubernetes cluster

```bash
bash scripts/deploy_dev.sh
```
