{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mangalodon.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "mangalodon.labels" -}}
helm.sh/chart: {{ include "mangalodon.chart" . }}
{{ include "mangalodon.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "mangalodon.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mangalodon.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Vault injector agent annotations
*/}}
{{- define "vault.annotations" -}}
vault.hashicorp.com/agent-inject: "true"
vault.hashicorp.com/service: {{ .Values.vault.server }}
vault.hashicorp.com/tls-secret: {{ .Chart.Name }}-vault-ca
vault.hashicorp.com/ca-cert: "/vault/tls/ca.crt"
vault.hashicorp.com/template-static-secret-render-interval: "5m"
{{- end }}