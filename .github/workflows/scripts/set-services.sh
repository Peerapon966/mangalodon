#!/bin/bash

set -euxo pipefail

services_template='[
  {
    "name": "app",
    "path": "src",
    "versionFile": "Chart.yaml",
    "versionRegex": "s/.*-appVersion:\\s*([^+]*)\\+.*/\\1/p"
  },
  {
    "name": "frontend",
    "path": "src/frontend",
    "versionFile": "values.yaml",
    "versionRegex": "s/.*frontend:[^-]*-\\s*tag:\\s*([^+]*)\\+.*/\\1/p"
  },
  {
    "name": "apiservice",
    "path": "src/apiservice",
    "versionFile": "values.yaml",
    "versionRegex": "s/.*apiservice:[^-]*-\\s*tag:\\s*([^+]*)\\+.*/\\1/p"
  },
  {
    "name": "scrapeservice",
    "path": "src/scrapeservice",
    "versionFile": "values.yaml",
    "versionRegex": "s/.*scrapeservice:[^-]*-\\s*tag:\\s*([^+]*)\\+.*/\\1/p"
  }
]'

services=$(echo "$services_template" | jq \
  --arg app "$(yq eval '.appVersion' helm-chart/Chart.yaml)" \
  --arg fe "$(yq eval '.frontend.image.tag' helm-chart/values.yaml)" \
  --arg api "$(yq eval '.apiservice.image.tag' helm-chart/values.yaml)" \
  --arg scr "$(yq eval '.scrapeservice.image.tag' helm-chart/values.yaml)" \
  '(.[] | select(.name == "app")).version = $app |
  (.[] | select(.name == "frontend")).version = $fe |
  (.[] | select(.name == "apiservice")).version = $api |
  (.[] | select(.name == "scrapeservice")).version = $scr')

echo "services=$services" >> $GITHUB_OUTPUT