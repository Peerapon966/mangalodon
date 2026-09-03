#!/bin/bash

set -eux

current_version=$(git tag --sort version:refname | tail -n 1)
new_version=$(yq eval '.version' helm-chart/Chart.yaml)

if [[ $(semver $new_version | wc -l) == 0 ]]; then
  echo "Invalid Helm chart version: $new_version"
  exit 1
fi

# semver tool sorts all input versions in asc order
# pass both new and previous version tags and expect the last row = new tag
if [[ $(semver "$current_version" "$new_version" | tail -n 1) != "$new_version" ]]; then
  echo "Helm chart version is decreasing ($new_version < $current_version)"
  echo "New release must have chart version increased from the previous releases"
  exit 1
fi
