#!/bin/bash

set -euxo pipefail

prev_tag=$(git tag --sort version:refname | tail -n 2 | head -n 1)
while read -r item; do
  name=$(jq -r '.name' <<< "$item")
  path=$(jq -r '.path' <<< "$item")
  version=$(jq -r '.version' <<< "$item")

  if [[ $name == "app" ]]; then
    continue
  fi

  if [[ $(git diff --name-only $prev_tag $GITHUB_REF $path 2>/dev/null | wc -l) <= 0 ]]; then
    echo "Skip building $name, no source code updates in $path"
    continue
  fi

  docker buildx build -t $REGISTRY/mangalodon/$name:$version --push $path
done < <(jq -c '.[]' <<< "$SERVICES")
