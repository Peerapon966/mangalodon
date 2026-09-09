#!/bin/bash

set -euo pipefail

root_dir="$(dirname -- $(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd))"
project="mangalodon"
env="dev"

pushd "${root_dir}/infra" > /dev/null 2>&1
terraform init -backend-config=init/backend-${env}.hcl -reconfigure
terraform workspace select -or-create $env
terraform apply -var-file=tfvars/${env}.tfvars -auto-approve
repos=$(terraform output -json ecr_repositories)
image_tags=$(terraform output -json image_tags)
popd

helm upgrade --install $project charts/${project} \
  --namespace ${project}-${env} \
  --create-namespace \
  --set ingress.host=${project}-${env}.com \
  --set postgres.vaultSecretPath=kv/${project}/${env}/postgres/login-secret \
  --set postgres.instanceCount=1 \
  --set frontend.image.repository=$(echo $repos | jq -r '.frontend'),frontend.image.tag=$(echo $image_tags | jq -r '.frontend') \
  --set apiservice.image.repository=$(echo $repos | jq -r '.apiservice'),apiservice.image.tag=$(echo $image_tags | jq -r '.apiservice') \
  --set scrapeservice.image.repository=$(echo $repos | jq -r '.scrapeservice'),scrapeservice.image.tag=$(echo $image_tags | jq -r '.scrapeservice')