---
title: Kubernetes Reporter Helm Chart
---

# k8s-reporter

![Version: 1.11.0](https://img.shields.io/badge/Version-1.11.0-informational?style=flat-square)

A Helm chart for installing the Kosli K8S reporter as a cronjob.
The chart allows you to create a Kubernetes cronjob and all its necessary RBAC to report running images to Kosli at a given cron schedule.

## Prerequisites

- A Kubernetes cluster (minimum supported version is `v1.21`)
- Helm v3.0+
- If you want to report artifacts from just one namespace, you need to have permissions to `get` and `list` pods in that namespace.
- If you want to report artifacts from multiple namespaces or entire cluster, you need to have cluster-wide permissions to `get` and `list` pods.

## Installing the chart

To install this chart via the Helm chart repository:

1. Add the Kosli helm repo
```shell {.command}
helm repo add kosli https://charts.kosli.com/ && helm repo update
```

2. Create a secret for the Kosli API token
```shell {.command}
kubectl create secret generic kosli-api-token --from-literal=key=<your-api-key>
```

3. Install the helm chart

A. To report artifacts running in entire cluster (requires cluster-wide read permissions):

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter \
    --set reporterConfig.kosliOrg=<your-org> \
    --set reporterConfig.kosliEnvironmentName=<your-env-name>
```

B. To report artifacts running in multiple namespaces (requires cluster-wide read permissions):

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter \
    --set reporterConfig.kosliOrg=<your-org> \
    --set reporterConfig.kosliEnvironmentName=<your-env-name> \
    --set reporterConfig.namespaces=<namespace1,namespace2>
```

C. To report artifacts running in one namespace (requires namespace-scoped read permissions):

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter \
    --set reporterConfig.kosliOrg=<your-org> \
    --set reporterConfig.kosliEnvironmentName=<your-env-name> \
    --set reporterConfig.namespaces=<namespace1> \
    --set serviceAccount.permissionScope=namespace
```

> Chart source can be found at https://github.com/kosli-dev/cli/tree/main/charts/k8s-reporter

> See all available [configuration options](#configurations) below.

## Upgrading the chart

```shell {.command}
helm upgrade kosli-reporter kosli/k8s-reporter ...
```

## Uninstalling chart

```shell {.command}
helm uninstall kosli-reporter
```

## Configurations
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cronSchedule | string | `"*/5 * * * *"` | the cron schedule at which the reporter is triggered to report to Kosli   |
| fullnameOverride | string | `""` | overrides the fullname used for the created k8s resources. It has higher precedence than `nameOverride` |
| image.pullPolicy | string | `"IfNotPresent"` | the kosli reporter image pull policy |
| image.repository | string | `"ghcr.io/kosli-dev/cli"` | the kosli reporter image repository |
| image.tag | string | `"v2.11.27"` | the kosli reporter image tag, overrides the image tag whose default is the chart appVersion. |
| kosliApiToken.secretKey | string | `"key"` | the name of the key in the secret data which contains the Kosli API token |
| kosliApiToken.secretName | string | `"kosli-api-token"` | the name of the secret containing the kosli API token |
| nameOverride | string | `""` | overrides the name used for the created k8s resources. If `fullnameOverride` is provided, it has higher precedence than this one |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` | custom labels to add to pods |
| reporterConfig.dryRun | bool | `false` | whether the dry run mode is enabled or not. In dry run mode, the reporter logs the reports to stdout and does not send them to kosli. |
| reporterConfig.excludeNamespaces | string | `""` | the namespaces to exclude from scanning and reporting. Cannot be combined with namespaces or namespacesRegex. It is a comma separated list of namespace names. Leave this and excludeNamespacesRegex unset if you want to report what is running in the entire cluster |
| reporterConfig.excludeNamespacesRegex | string | `""` | the namespaces Regex patterns to exclude from scanning and reporting. Does not have effect if excludeNamespaces is set. Cannot be combined with namespaces or namespacesRegex. Requires cluster-wide permissions. It is a comma separated list of namespace regex patterns. Leave this and excludeNamespaces unset if you want to report what is running in the entire cluster |
| reporterConfig.httpProxy | string | `""` | the http proxy url |
| reporterConfig.kosliEnvironmentName | string | `""` | the name of Kosli environment that the k8s cluster/namespace correlates to |
| reporterConfig.kosliOrg | string | `""` | the name of the Kosli org |
| reporterConfig.namespaces | string | `""` | the namespaces to scan and report. Cannot be combined with excludeNamespaces or excludeNamespacesRegex. It is a comma separated list of namespace names. Leave this and namespacesRegex unset if you want to report what is running in the entire cluster |
| reporterConfig.namespacesRegex | string | `""` | the namespaces Regex patterns to scan and report. Does not have effect if namespaces is set. Requires cluster-wide permissions. Cannot be combined with excludeNamespaces or excludeNamespacesRegex. It is a comma separated list of namespace regex patterns. Leave this and namespaces unset if you want to report what is running in the entire cluster |
| reporterConfig.securityContext | object | `{"allowPrivilegeEscalation":false,"runAsNonRoot":true,"runAsUser":1000}` | the security context for the reporter cronjob Set to null or {} to disable security context entirely (not recommended) For OpenShift, you can omit runAsUser to let OpenShift assign the UID |
| reporterConfig.securityContext.allowPrivilegeEscalation | bool | `false` | whether to allow privilege escalation |
| reporterConfig.securityContext.runAsNonRoot | bool | `true` | whether to run as non root |
| reporterConfig.securityContext.runAsUser | int | `1000` | the user id to run as Omit this field for OpenShift environments to allow automatic UID assignment |
| resources.limits.cpu | string | `"100m"` | the cpu limit |
| resources.limits.memory | string | `"256Mi"` | the memory limit |
| resources.requests.memory | string | `"64Mi"` | the memory request |
| serviceAccount.annotations | object | `{}` | annotations to add to the service account |
| serviceAccount.create | bool | `true` | specifies whether a service account should be created |
| serviceAccount.name | string | `""` | the name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceAccount.permissionScope | string | `"cluster"` | specifies whether to create a cluster-wide permissions for the service account or namespace-scoped permissions. allowed values are: [cluster, namespace] |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)
    
