{{/*
Validate reporterConfig.environments: non-empty, each entry has name, no duplicate names,
and no environment combines include (namespaces/namespacesRegex) with exclude (excludeNamespaces/excludeNamespacesRegex).
Regex pattern validity is still checked by the CLI when it parses the config file (validateK8SSnapshotConfig in snapshotK8S.go).
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
{{/* Duplicate environment names */}}
{{- range $idx2, $e2 := $envs -}}
{{- if and (ne $idx $idx2) (eq $e.name $e2.name) -}}
{{- fail (printf "reporterConfig.environments: duplicate environment name '%s'" $e.name) -}}
{{- end -}}
{{- end -}}
{{/* Include vs exclude mutual exclusion per environment */}}
{{- $hasInclude := or (gt (len ($e.namespaces | default list)) 0) (gt (len ($e.namespacesRegex | default list)) 0) -}}
{{- $hasExclude := or (gt (len ($e.excludeNamespaces | default list)) 0) (gt (len ($e.excludeNamespacesRegex | default list)) 0) -}}
{{- if and $hasInclude $hasExclude -}}
{{- fail (printf "reporterConfig.environments: environment '%s' cannot combine include (namespaces/namespacesRegex) with exclude (excludeNamespaces/excludeNamespacesRegex)" $e.name) -}}
{{- end -}}
{{- end -}}
{{- end -}}
