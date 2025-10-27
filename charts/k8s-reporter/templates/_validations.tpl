{{/*
Validate that namespacesRegex is not used with namespace-scoped permissions
*/}}
{{- define "k8s-reporter.validateNamespacesRegex" -}}
{{- if and (eq .Values.serviceAccount.permissionScope "namespace") (ne .Values.reporterConfig.namespacesRegex "") -}}
{{- fail "namespacesRegex cannot be used with namespace-scoped permissions (serviceAccount.permissionScope: namespace). namespacesRegex requires cluster-wide permissions." -}}
{{- end -}}
{{- end -}}

{{/*
Validate that excludeNamespacesRegex is not used with namespace-scoped permissions
*/}}
{{- define "k8s-reporter.validateExcludeNamespacesRegex" -}}
{{- if and (eq .Values.serviceAccount.permissionScope "namespace") (ne .Values.reporterConfig.excludeNamespacesRegex "") -}}
{{- fail "excludeNamespacesRegex cannot be used with namespace-scoped permissions (serviceAccount.permissionScope: namespace). excludeNamespacesRegex requires cluster-wide permissions." -}}
{{- end -}}
{{- end -}}

{{/*
Validate that exclude options are not combined with include options
*/}}
{{- define "k8s-reporter.validateExcludeOptions" -}}
{{- if and (ne .Values.reporterConfig.namespaces "") (or (ne .Values.reporterConfig.excludeNamespaces "") (ne .Values.reporterConfig.excludeNamespacesRegex "")) -}}
{{- fail "excludeNamespaces and excludeNamespacesRegex cannot be combined with namespaces. Use either include (namespaces/namespacesRegex) or exclude (excludeNamespaces/excludeNamespacesRegex) options, but not both." -}}
{{- end -}}
{{- if and (ne .Values.reporterConfig.namespacesRegex "") (or (ne .Values.reporterConfig.excludeNamespaces "") (ne .Values.reporterConfig.excludeNamespacesRegex "")) -}}
{{- fail "excludeNamespaces and excludeNamespacesRegex cannot be combined with namespacesRegex. Use either include (namespaces/namespacesRegex) or exclude (excludeNamespaces/excludeNamespacesRegex) options, but not both." -}}
{{- end -}}
{{- end -}} 