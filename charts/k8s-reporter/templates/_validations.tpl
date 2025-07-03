{{/*
Validate that namespacesRegex is not used with namespace-scoped permissions
*/}}
{{- define "k8s-reporter.validateNamespacesRegex" -}}
{{- if and (eq .Values.serviceAccount.permissionScope "namespace") (ne .Values.reporterConfig.namespacesRegex "") -}}
{{- fail "namespacesRegex cannot be used with namespace-scoped permissions (serviceAccount.permissionScope: namespace). namespacesRegex requires cluster-wide permissions." -}}
{{- end -}}
{{- end -}} 