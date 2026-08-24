# Agent Co-working Rules

This document outlines the strict boundaries and rules of co-working between the human developer (User) and the AI Assistant (Agent) for this project.

## Division of Labor

### The Human (User) owns:
The User is solely responsible for writing, configuring, deploying, and maintaining the following components:
- **Kubernetes**: All `manifests/` and `helm/` charts.
- **Scraper Service**: The `src/scrapeservice/` codebase (Go).
- **Message Broker**: Any provisioning, configuration, or deployment related to RabbitMQ.
- **Architecture Design**: The high-level design of the Backend API and DB Schema.

### The AI (Agent) must NEVER:
1. **Kubernetes/Helm:** You must **NOT** add new components, resources, or significant configuration blocks to Kubernetes YAML manifests. Typo fixes or misconfiguration fixes are allowed, but you **MUST INFORM THE USER FIRST** and get approval before applying the fix.
2. **Go Scrapers:** Create, modify, or suggest changes to the Go scraper code.
3. **RabbitMQ Config:** Touch anything related to RabbitMQ provisioning or configuration (though using `amqplib` within the Hono app is allowed).

If the User asks the Agent to perform tasks in these restricted areas outside of these exceptions, the Agent must respectfully decline and remind the User of these boundaries.

## Agent Responsibilities

The Agent is permitted and expected to assist with:
- **Backend API**: Implementing the Hono API in `src/apiservice/` based on the User's designs.
- **Database**: Implementing the DB Schema and raw SQL integrations (Postgres.js) based on the User's designs.
- **Frontend**: The `src/frontend/` codebase (React/Vite).
- General project documentation (e.g., `README.md`).
- Terraform execution and Docker Image builds.
- Project scaffolding and creating boilerplate directory structures.
- General architectural discussions, explanations, or answering conceptual questions.
