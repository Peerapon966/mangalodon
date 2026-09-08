#!/bin/bash

set -euo pipefail

services_template='[
  {
    "name": "app",
    "path": "src",
    "versionFile": "charts/mangalodon/Chart.yaml",
    "versionRegex": "s/.*-appVersion:\\s*([^+]*)\\+.*/\\1/p"
  },
  {
    "name": "frontend",
    "path": "src/frontend",
    "versionFile": "charts/mangalodon/values.yaml",
    "versionRegex": "s/.*frontend:.*tag:[[:space:]]*([^+]*)\+.*/\1/p"
  },
  {
    "name": "apiservice",
    "path": "src/apiservice",
    "versionFile": "charts/mangalodon/values.yaml",
    "versionRegex": "s/.*apiservice:.*tag:[[:space:]]*([^+]*)\+.*/\1/p"
  },
  {
    "name": "scrapeservice",
    "path": "src/scrapeservice",
    "versionFile": "charts/mangalodon/values.yaml",
    "versionRegex": "s/.*scrapeservice:.*tag:[[:space:]]*([^+]*)\+.*/\1/p"
  }
]'

services=$(echo "$services_template" | jq \
  --arg app "$(yq eval '.appVersion' charts/mangalodon/Chart.yaml)" \
  --arg fe "$(yq eval '.frontend.image.tag' charts/mangalodon/values.yaml)" \
  --arg api "$(yq eval '.apiservice.image.tag' charts/mangalodon/values.yaml)" \
  --arg scr "$(yq eval '.scrapeservice.image.tag' charts/mangalodon/values.yaml)" \
  '(.[] | select(.name == "app")).version = $app |
  (.[] | select(.name == "frontend")).version = $fe |
  (.[] | select(.name == "apiservice")).version = $api |
  (.[] | select(.name == "scrapeservice")).version = $scr')

echo "services=$(echo $services | jq -c '.')" >> $GITHUB_OUTPUT