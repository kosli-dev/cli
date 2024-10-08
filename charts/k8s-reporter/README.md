---
title: Kubernetes Reporter Helm Chart
---

# k8s-reporter

![Version: 1.5.1](https://img.shields.io/badge/Version-1.5.1-informational?style=flat-square)

A Helm chart for installing the Kosli K8S reporter as a cronjob.
The chart allows you to create a Kubernetes cronjob and all its necessary RBAC to report running images to Kosli at a given cron schedule.

## Prerequisites

- A Kubernetes cluster (minimum supported version is `v1.21`)
- Helm v3.0+
- Create a secret for the Kosli API token which will be used for reporting. You can create a secret by running: `kubectl create secret generic <secret-name> --from-literal=<secret-key>=<your-api-key>`

## Installing the chart

To install this chart via the Helm chart repository:

```shell
helm repo add kosli https://charts.kosli.com/
helm repo update
helm install [RELEASE-NAME] kosli/k8s-reporter -f [VALUES-FILE-PATH]
```

> Chart source can be found at https://github.com/kosli-dev/cli/tree/main/charts/k8s-reporter

## Upgrading the chart

```shell
helm upgrade [RELEASE-NAME] kosli/k8s-reporter
```

## Uninstalling chart

```shell
helm uninstall [RELEASE-NAME]
```

## Configurations
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cronSchedule | string | `"*/5 * * * *"` | the cron schedule at which the reporter is triggered to report to Kosli   |
| fullnameOverride | string | `""` | overrides the fullname used for the created k8s resources. It has higher precedence than `nameOverride` |
| image.pullPolicy | string | `"IfNotPresent"` | the kosli reporter image pull policy |
| image.repository | string | `"ghcr.io/kosli-dev/cli"` | the kosli reporter image repository |
| image.tag | string | `"v2.10.13"` | the kosli reporter image tag, overrides the image tag whose default is the chart appVersion. |
| kosliApiToken.secretKey | string | `"key"` | the name of the key in the secret data which contains the Kosli API token |
| kosliApiToken.secretName | string | `"kosli-api-token"` | the name of the secret containing the kosli API token |
| nameOverride | string | `""` | overrides the name used for the created k8s resources. If `fullnameOverride` is provided, it has higher precedence than this one |
| podAnnotations | object | `{}` |  |
| reporterConfig.dryRun | bool | `false` | whether the dry run mode is enabled or not. In dry run mode, the reporter logs the reports to stdout and does not send them to kosli. |
| reporterConfig.httpProxy | string | `""` | the http proxy url |
| reporterConfig.kosliEnvironmentName | string | `""` | the name of Kosli environment that the k8s cluster/namespace correlates to |
| reporterConfig.kosliOrg | string | `""` | the name of the Kosli org |
| reporterConfig.namespaces | string | `""` | the namespaces which represent the environment. It is a comma separated list of namespace name regex patterns. e.g. `^prod$,^dev-*` reports for the `prod` namespace and any namespace that starts with `dev-` leave this unset if you want to report what is running in the entire cluster |
| resources.limits.cpu | string | `"100m"` | the cpu limit |
| resources.limits.memory | string | `"256Mi"` | the memory limit |
| resources.requests.memory | string | `"64Mi"` | the memory request |
| serviceAccount.annotations | object | `{}` | annotations to add to the service account |
| serviceAccount.create | bool | `true` | specifies whether a service account should be created |
| serviceAccount.name | string | `""` | the name of the service account to use. If not set and create is true, a name is generated using the fullname template |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)

