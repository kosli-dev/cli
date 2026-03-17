---
title: Kubernetes Reporter Helm Chart
---

# k8s-reporter

![Version: 2.1.0](https://img.shields.io/badge/Version-2.1.0-informational?style=flat-square)

A Helm chart for installing the Kosli K8S reporter as a cronjob.
The chart allows you to create a Kubernetes cronjob and all its necessary RBAC to report running images to Kosli at a given cron schedule.

Configuration is done via **reporterConfig.environments**: a list of Kosli environments to report to. Each entry has a required `name` and optional namespace selectors. Use one entry for a single environment, or multiple entries to report to different environments with different selectors.

## Breaking change in v2.0.0

Version 2.0.0 removes the previous single-environment mode (`kosliEnvironmentName` and the `namespaces` / `namespacesRegex` / `excludeNamespaces` / `excludeNamespacesRegex` flags). You now configure one or more environments only via **reporterConfig.environments**. To report a single environment, use a list with one entry.

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

Configure **reporterConfig.environments** (required). Each entry has required `name` and optional `namespaces`, `namespacesRegex`, `excludeNamespaces`, `excludeNamespacesRegex`. Omit namespace fields for an entry to report the entire cluster to that environment.

**One environment, entire cluster:**

```yaml
# values.yaml
reporterConfig:
  kosliOrg: <your-org>
  environments:
    - name: <your-env-name>
```

**One environment, specific namespaces:**

```yaml
reporterConfig:
  kosliOrg: <your-org>
  environments:
    - name: <your-env-name>
      namespaces: [namespace1, namespace2]
```

**Multiple environments with different selectors:**

```yaml
reporterConfig:
  kosliOrg: <your-org>
  environments:
    - name: prod-env
      namespaces: [prod-ns1, prod-ns2]
    - name: staging-env
      namespacesRegex: ["^staging-.*"]
    - name: infra-env
      excludeNamespaces: [prod-ns1, prod-ns2, default]
```

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter -f values.yaml
```

> Chart source can be found at https://github.com/kosli-dev/cli/tree/main/charts/k8s-reporter

> See all available [configuration options](#configurations) below.

## Upgrading the chart

If upgrading from v1.x to v2.0.0, migrate your values to the **environments** list format (see above). Then:

```shell {.command}
helm upgrade kosli-reporter kosli/k8s-reporter -f values.yaml
```

## Uninstalling chart

```shell {.command}
helm uninstall kosli-reporter
```

## Configurations
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| concurrencyPolicy | string | `"Replace"` | specifies how to treat concurrent executions of a Job that is created by this CronJob |
| cronSchedule | string | `"*/5 * * * *"` | the cron schedule at which the reporter is triggered to report to Kosli |
| failedJobsHistoryLimit | int | `1` | specifies the number of failed finished jobs to keep |
| fullnameOverride | string | `""` | overrides the fullname used for the created k8s resources. It has higher precedence than `nameOverride` |
| image.pullPolicy | string | `"IfNotPresent"` | the kosli reporter image pull policy |
| image.repository | string | `"ghcr.io/kosli-dev/cli"` | the kosli reporter image repository |
| image.tag | string | `"v2.12.0"` | the kosli reporter image tag, overrides the image tag whose default is the chart appVersion. |
| kosliApiToken.secretKey | string | `"key"` | the name of the key in the secret data which contains the Kosli API token |
| kosliApiToken.secretName | string | `"kosli-api-token"` | the name of the secret containing the kosli API token |
| nameOverride | string | `""` | overrides the name used for the created k8s resources. If `fullnameOverride` is provided, it has higher precedence than this one |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` | custom labels to add to pods |
| reporterConfig.dryRun | bool | `false` |  |
| reporterConfig.environments | list | `[]` | List of Kosli environments to report to. Each entry has required 'name' and optional namespace selectors. Use one entry to report a single environment; use multiple entries to report to multiple environments with different selectors. Per entry: name (required), namespaces, namespacesRegex, excludeNamespaces, excludeNamespacesRegex (optional). Leave namespace fields unset for an entry to report the entire cluster to that environment. |
| reporterConfig.httpProxy | string | `""` | the http proxy url |
| reporterConfig.kosliOrg | string | `""` | the name of the Kosli org |
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
| successfulJobsHistoryLimit | int | `3` | specifies the number of successful finished jobs to keep |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)

