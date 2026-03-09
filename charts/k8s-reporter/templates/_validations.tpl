{{/*
Validate that reporterConfig.environments is non-empty and each entry has a name
*/}}
{{- define "k8s-reporter.validateEnvironments" -}}
{{- $envs := .Values.reporterConfig.environments -}}
{{- if eq (len $envs) 0 -}}
{{- fail "reporterConfig.environments is required and must contain at least one entry. Each entry must have 'name'." -}}
{{- end -}}
{{- range $idx, $e := $envs -}}
{{- if not $e.name -}}
{{- fail (printf "reporterConfig.environments[%d]: 'name' is required for each entry." $idx) -}}
{{- end -}}
{{- end -}}
{{- end -}}
