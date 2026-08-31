#!/bin/bash

set -euxo pipefail

versions="{
  \"tag_version\": \"$GITHUB_REF\",
  \"chart_version\": \"$(yq eval '.version' helm-chart/Chart.yaml)\"
}"

while read -r key value; do
  echo "Validating $key semver"
  if [[ $(semver $value | wc -l) == 0 ]]; then
    echo "$key invalid version: $value"
    exit 1
  fi

  export $key=$value
done < <(echo "$versions" | jq -r 'to_entries[] | "\(.key)\t\(.value)"')

if [[ "${GITHUB_REF#v}" != "$chart_version" ]]; then
  echo "Version specified in tag (${GITHUB_REF#v}) doesn't match chart version in Chart.yaml ($chart_version)"
  exit 1
fi

# semver tool sorts all input versions in asc order
# pass both new and previous version tags and expect the last row = new tag
prev_tag=$(git tag --sort version:refname | tail -n 2 | head -n 1)
if [[ $(semver "$prev_tag" "$GITHUB_REF" | tail -n 1) != "${GITHUB_REF#v}" ]]; then
  echo "Helm chart version is decreasing ($GITHUB_REF < $prev_tag)"
  echo "New release must have chart version increased from the previous releases"
  exit 1
fi
