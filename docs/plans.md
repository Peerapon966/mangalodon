# Architecture Plan

The whole point of this project is to practice kubernetes and distributed systems. By keeping the actual code simple (just decrypting a file), we can focus entirely on the infrastructure side of things.

## Development approach

We're skipping docker-compose completely. 
1. Test code locally (npm run dev, go run, etc)
2. Build docker images and deploy straight to a local k8s cluster like minikube using standard manifests
3. Later we can add skaffold or tilt to speed up the loop

## The Components

### 1. Frontend
A simple web page for file uploads.
K8s concepts: Deployments, ConfigMaps for UI settings, Ingress for routing.

### 2. API (NestJS)
Receives the file, puts it in a shared volume, and sends a message to rabbitmq. 
NestJS is used because it has built-in microservice tools and feels like a real enterprise app.
K8s concepts: ServiceAccounts, RBAC, NetworkPolicies to restrict access.

### 3. Queue (RabbitMQ)
Message broker holding tasks for the workers. Kafka is too heavy for this, rabbitmq fits better.
K8s concepts: StatefulSets so it maintains state on restart.

### 4. Workers (Go)
Listens to rabbitmq, decrypts the file, and logs it. Written in go so we can use tiny alpine/scratch images for fast startup.
K8s concepts: InitContainers to wait for the db/queue to be ready, LivenessProbes to kill the pod if the queue connection drops.

### 5. Logging (ELK)
Collects worker logs.
K8s concepts: Sidecar pattern. The go worker writes to a file, a filebeat container in the same pod reads it and sends it to elasticsearch.

### 6. Database (PostgreSQL)
Tracks what's been decrypted.
K8s concepts: PersistentVolumes, PersistentVolumeClaims, Secrets for passwords.

### 7. File Sweeper
Deletes files older than 24 hours to save space.
K8s concepts: CronJobs.

### 8. Node Monitor
Background agent monitoring disk space.
K8s concepts: DaemonSets to run on every node, Taints/Tolerations to allow running on control plane nodes.

## Training Scenarios

Things to try once it's running:
- Backup etcd, delete the namespace, and try to restore it
- Flood the UI with traffic and try to drain/upgrade a node without dropping requests
- Delete the postgres pod to watch k8s re-attach the persistent volume
- Setup an HPA to auto-scale the go workers based on CPU usage
