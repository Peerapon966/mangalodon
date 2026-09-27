{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mangalodon.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Pod security context
*/}}
{{- define "mangalodon.podSecurityContext" -}}
fsGroup: 65532
runAsUser: 65532
runAsGroup: 65532
runAsNonRoot: true
seccompProfile:
  type: RuntimeDefault
{{- end }}

{{/*
Container security context
*/}}
{{- define "mangalodon.containerSecurityContext" -}}
allowPrivilegeEscalation: false
capabilities:
  drop:
    - ALL
privileged: false
readOnlyRootFilesystem: true
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