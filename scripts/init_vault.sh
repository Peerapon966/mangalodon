#!/bin/bash

set -euo pipefail

usage() {
    echo "Usage: $(basename "$0") -e|--env <environment> [-h|--help]"
    echo ""
    echo "Options:"
    echo "  -e, --env      Specify the target environment (e.g., dev, staging, prod)"
    echo "  -h, --help     Show this help message and exit"
    exit 1
}

env=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        -e|--env)
            if [[ -z "${2:-}" || "${2:-}" == -* ]]; then
                echo "Error: Option '$1' requires an argument." >&2
                exit 1
            fi
            env="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        -*)
            echo "Error: Unknown option '$1'" >&2
            usage
            ;;
        *)
            echo "Error: Unexpected argument '$1'" >&2
            usage
            ;;
    esac
done

if [[ -z "$env" ]]; then
    echo "Error: The -e|--env option is required." >&2
    usage
fi

root_dir="$(dirname -- $(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd))"
pushd "${root_dir}/vault" > /dev/null 2>&1
terraform init -backend-config=init/backend-${env}.hcl -reconfigure
sleep 2

terraform workspace select -or-create $env
sleep 2

terraform apply -var-file=tfvars/${env}.tfvars -auto-approve
popd

echo "Vault init successfully"