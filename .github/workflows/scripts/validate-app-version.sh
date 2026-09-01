#!/bin/bash

set -euxo pipefail

# prev_tag should be empty when there's only 1 tag (first release)
prev_tag=""
if [[ $(git tag --sort version:refname | tail -n 2 | wc -l) == 2 ]]; then
  prev_tag=$(git tag --sort version:refname | tail -n 2 | head -n 1)
fi

while read -r item; do
  name=$(jq -r '.name' <<< "$item")
  path=$(jq -r '.path' <<< "$item")
  versionFile=$(jq -r '.versionFile' <<< "$item")
  versionRegex=$(jq -r '.versionRegex' <<< "$item")
  version=$(jq -r '.version' <<< "$item")

  echo "Validating $name semver"
  if [[ $(semver $version | wc -l) == 0 ]]; then
    echo "$name invalid version: $version"
    exit 1
  fi

  # check if source code in src/<service> dir has any update or not
  is_app_updated=false
  if [[ $prev_tag == "" || $(git diff --name-only $prev_tag $GITHUB_REF $path 2>/dev/null | wc -l) > 0 ]]; then
    is_app_updated=true
  fi
  if [[ "$name" == "app" ]]; then
    echo "is_app_updated=$is_app_updated" >> $GITHUB_OUTPUT
  fi

  prev_app_version=$(git diff $prev_tag $GITHUB_REF $versionFile 2>/dev/null | tr -d '\n' | sed -nE "$versionRegex" | tr -d '"' )
  is_app_version_updated=false
  if [[ $prev_tag == "" || $prev_app_version != "" ]]; then
    is_app_version_updated=true
  fi

  # validate 2 cases
  # 1. if src/<service> has updates => version value must also be updated & prev version < new version
  # 2. if src/<service> has no updates => version value must not be updated
  if [[ "$is_app_updated" == "true" && "$is_app_version_updated" == "false" ]]; then
    echo "Detected source code updates in $path but version in $versionFile is not bumping up"
    echo "Increase version in $versionFile and try again"
    exit 1
  fi

  if [[ "$is_app_updated" == "true" && "$is_app_version_updated" == "true" ]]; then
    # semver tool sorts all input versions in asc order
    # pass both new and previous versions and expect the last row = new version
    if [[ $(semver "$prev_app_version" "$version" | tail -n 1) != "${version#v}" ]]; then
      echo "$name version is decreasing ($version < $prev_app_version)"
      echo "Newer release must have version increased from the previous release"
      exit 1
    fi
  fi

  if [[ "$is_app_updated" == "false" && "$is_app_version_updated" == "true" ]]; then
    echo "$name version in $versionFile has changed without any source code updates in $path"
    echo "The workflow expects version to stay unchanged when no updates in source code present"
    exit 1
  fi
done < <(jq -c '.[]' <<< "$SERVICES")
