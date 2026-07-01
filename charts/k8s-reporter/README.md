---
title: Kubernetes Reporter Helm Chart
---

# k8s-reporter

![Version: 2.3.1](https://img.shields.io/badge/Version-2.3.1-informational?style=flat-square)

A Helm chart for installing the Kosli K8S reporter as a CronJob.
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
```shell
helm repo add kosli https://charts.kosli.com/ && helm repo update
```

2. Create a secret for the Kosli API token
```shell
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

```shell
helm install kosli-reporter kosli/k8s-reporter -f values.yaml
```

> Chart source can be found at https://github.com/kosli-dev/cli/tree/main/charts/k8s-reporter

> See all available [configuration options](#configurations) below.

## Upgrading the chart

If upgrading from v1.x to v2.0.0, migrate your values to the **environments** list format (see above). Then:

```shell
helm upgrade kosli-reporter kosli/k8s-reporter -f values.yaml
```

## Uninstalling chart

```shell
helm uninstall kosli-reporter
```

## Running behind a TLS-inspecting proxy (corporate / custom CA bundle)

If your network sits behind a TLS-inspecting appliance (Zscaler, Netskope, Palo Alto, etc.) that re-signs HTTPS traffic with a corporate CA certificate, the reporter will fail with `x509: certificate signed by unknown authority`. To fix this, make the appliance's CA bundle available to the reporter.

The chart offers two ways to do this. Use whichever fits your deployment flow.

### Option 1 — `customCA` convenience wrapper (recommended for the common case)

1. Create a Secret containing the corporate CA certificate (PEM format, single cert or bundle):

```shell
kubectl create secret generic corporate-ca-bundle --from-file=ca.crt=/path/to/corporate-ca.crt
```

2. Enable the wrapper in `values.yaml`:

```yaml
customCA:
  enabled: true
  secretName: corporate-ca-bundle
  key: ca.crt
```

The chart mounts the certificate as a single file at `/etc/ssl/certs/kosli-custom-ca.crt` using `subPath`. Go's standard library on Linux loads CA roots in two independent passes — it reads the system bundle file (e.g. `/etc/ssl/certs/ca-certificates.crt`) and **also** scans `/etc/ssl/certs/` for additional certificate files. The mounted file is picked up by the directory scan and added to the trust store alongside the system roots, so no `SSL_CERT_FILE` env var is needed.

The wrapper deliberately does **not** set `SSL_CERT_FILE`. Setting it would replace the system bundle entirely with the customer's file, breaking trust for any public CAs the bundle does not include.

### Option 2 — generic `extraVolumes` / `extraVolumeMounts` / `extraEnvVars`

Use these when you need a non-default mount path, a ConfigMap instead of a Secret, multiple volumes, or any other shape the wrapper does not cover:

```yaml
extraVolumes:
  - name: corporate-ca
    secret:
      secretName: corporate-ca-bundle

extraVolumeMounts:
  - name: corporate-ca
    mountPath: /etc/ssl/certs/corporate
    readOnly: true
```

Note: if you mount the CA outside `/etc/ssl/certs/` and set `SSL_CERT_FILE` via `extraEnvVars`, your bundle must include the public CAs you also need to trust — Go uses only that file when `SSL_CERT_FILE` is set.

### Pod Security Standards

Both options use `secret`-backed volumes, which are permitted under the Pod Security Standards `restricted` profile. `hostPath` mounts are not permitted under that profile and should not be used here.

### Cluster-wide alternative

If you already run [cert-manager's trust-manager](https://cert-manager.io/docs/trust/trust-manager/) to distribute a corporate CA bundle into a well-known ConfigMap in every namespace, point `extraVolumes` / `extraVolumeMounts` at that ConfigMap instead of creating a per-namespace Secret.

## Running on EKS with Karpenter (or another node autoscaler)

By default the reporter runs as a CronJob every 5 minutes. On clusters that use [Karpenter](https://karpenter.sh) for node autoscaling, this frequent scheduling can prevent nodes from being **consolidated** (scaled down).

The cause is Karpenter's `consolidateAfter` timer: Karpenter only consolidates a node once it has seen no pod scheduling activity on it for the configured window. A reporter pod arriving every 5 minutes keeps resetting that timer, so any node whose `consolidateAfter` is longer than the reporter interval never becomes eligible for consolidation (see [karpenter#1921](https://github.com/kubernetes-sigs/karpenter/issues/1921)). This is Karpenter working as designed, not a reporter bug.

Frequent snapshots are what let Kosli surface drift or an unauthorized change quickly, so the best fix keeps the 5-minute cadence and moves the reporter out of Karpenter's way. Widening the interval trades away that detection speed and should be a last resort.

### 1. Pin the reporter to a stable node group (recommended)

If you run a stable managed node group that Karpenter does not manage, schedule the reporter there so it never disturbs Karpenter-managed nodes. Use `nodeSelector`, and `tolerations` if that node group is tainted:

```yaml
nodeSelector:
  eks.amazonaws.com/nodegroup: system   # your managed node group

tolerations:
  - key: dedicated
    operator: Equal
    value: system
    effect: NoSchedule
```

To steer the reporter *away* from Karpenter-managed nodes instead, use `affinity` (a plain `nodeSelector` cannot express "not on these nodes"):

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: karpenter.sh/nodepool
              operator: DoesNotExist
```

### 2. Run the reporter out of the cluster

For zero footprint on cluster nodes, run `kosli snapshot k8s` on a schedule outside the cluster (for example a CI cron job) with kubeconfig access, keeping your reporting cadence without placing a pod on the cluster's nodes. See the [Kubernetes environment reporting tutorial](https://docs.kosli.com/tutorials/report_k8s_envs/).

### 3. Widen the report interval (last resort)

Only if you cannot pin the reporter or move it out of cluster: set `cronSchedule` longer than your NodePool's `consolidateAfter` so nodes get quiet windows long enough to consolidate. This works, but a longer interval widens the window in which a change can go unreported, so prefer the options above.

```yaml
cronSchedule: "*/15 * * * *"
```

> Note: `karpenter.sh/do-not-disrupt: "true"` is **not** a fix here. It prevents Karpenter from disrupting the pod, which protects a mid-run report from interruption but makes consolidation of that node *less* likely, not more. Likewise `cluster-autoscaler.kubernetes.io/safe-to-evict` only affects the Kubernetes Cluster Autoscaler and is ignored by Karpenter.

## Configurations
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | affinity rules for scheduling the reporter pod. Supports nodeAffinity, podAffinity and podAntiAffinity. |
| concurrencyPolicy | string | `"Replace"` | specifies how to treat concurrent executions of a Job that is created by this CronJob |
| cronSchedule | string | `"*/5 * * * *"` | the cron schedule at which the reporter is triggered to report to Kosli |
| customCA | object | `{"enabled":false,"key":"ca.crt","secretName":""}` | convenience wrapper for mounting a corporate / custom CA bundle. See the "Running behind a TLS-inspecting proxy" section of the README for usage. |
| customCA.enabled | bool | `false` | enable mounting a corporate/custom CA bundle into the trust store |
| customCA.key | string | `"ca.crt"` | key within the Secret that holds the PEM-formatted CA certificate (single cert or multi-cert PEM bundle) |
| customCA.secretName | string | `""` | name of an existing Secret in the same namespace containing the CA bundle |
| env | object | `{}` | map of plain environment variables to inject into the reporter container. For a single-tenant Kosli instance, set `KOSLI_HOST` to `https://INSTANCE_NAME.kosli.com`. |
| extraEnvVars | list | `[]` | additional environment variables to inject into the reporter container. List of `{name, value}` or `{name, valueFrom}` entries, rendered verbatim into the container env. Supports plain values and valueFrom (`secretKeyRef` / `configMapKeyRef`). Note: entries here are appended after the chart's own env entries; on duplicate names the later entry wins. |
| extraVolumeMounts | list | `[]` | additional container-level volumeMounts for the reporter container. Rendered verbatim into the container spec alongside the chart's own mounts. |
| extraVolumes | list | `[]` | additional Pod-level volumes to attach to the reporter pod. Rendered verbatim into the Pod spec alongside the chart's own volumes. Use together with `extraVolumeMounts` to mount Secrets, ConfigMaps, or other volumes into the container. |
| failedJobsHistoryLimit | int | `1` | specifies the number of failed finished jobs to keep |
| fullnameOverride | string | `""` | overrides the fullname used for the created k8s resources. It has higher precedence than `nameOverride` |
| image.pullPolicy | string | `"IfNotPresent"` | the kosli reporter image pull policy |
| image.repository | string | `"ghcr.io/kosli-dev/cli"` | the kosli reporter image repository |
| image.tag | string | `""` | the kosli reporter image tag, overrides the image tag whose default is the chart appVersion. |
| kosliApiToken.secretKey | string | `"key"` | the name of the key in the secret data which contains the Kosli API token |
| kosliApiToken.secretName | string | `"kosli-api-token"` | the name of the secret containing the kosli API token |
| nameOverride | string | `""` | overrides the name used for the created k8s resources. If `fullnameOverride` is provided, it has higher precedence than this one |
| nodeSelector | object | `{}` | node labels for scheduling the reporter pod. On EKS with Karpenter, use this to pin the reporter to a stable managed node group (e.g. `eks.amazonaws.com/nodegroup: <name>`) so it does not interfere with node consolidation. See the "Running on EKS with Karpenter" section of the README. |
| podAnnotations | object | `{}` | annotations to add to the CronJob object itself. For pod-level annotations (added to each reporter pod), use `podTemplateAnnotations` instead. |
| podLabels | object | `{}` | custom labels to add to pods |
| podTemplateAnnotations | object | `{}` | annotations to add to the reporter pod template (applied to each Job pod that the CronJob creates) |
| reporterConfig.dryRun | bool | `false` | whether the dry run mode is enabled or not. In dry run mode, the reporter logs the reports to stdout and does not send them to kosli. |
| reporterConfig.environments | list | `[]` | List of Kosli environments to report to. Each entry has required 'name' and optional namespace selectors. Use one entry to report a single environment; use multiple entries to report to multiple environments with different selectors. Per entry: name (required), namespaces, namespacesRegex, excludeNamespaces, excludeNamespacesRegex (optional). Leave namespace fields unset for an entry to report the entire cluster to that environment. |
| reporterConfig.httpProxy | string | `""` | the http proxy url |
| reporterConfig.kosliOrg | string | `""` | the name of the Kosli org |
| reporterConfig.securityContext | object | `{"allowPrivilegeEscalation":false,"runAsNonRoot":true,"runAsUser":1000}` | the security context for the reporter cronjob. Set to null or {} to disable security context entirely (not recommended). For OpenShift with SCC, explicitly set runAsUser to null to let OpenShift assign the UID from the allowed range. Simply omitting runAsUser from your values override will not work because Helm deep-merges with these defaults. Example OpenShift override:   securityContext:     allowPrivilegeEscalation: false     runAsNonRoot: true     runAsUser: null |
| reporterConfig.securityContext.allowPrivilegeEscalation | bool | `false` | whether to allow privilege escalation |
| reporterConfig.securityContext.runAsNonRoot | bool | `true` | whether to run as non root |
| reporterConfig.securityContext.runAsUser | int | `1000` | the user id to run as. For OpenShift environments with SCC, set to null (runAsUser: null) to allow automatic UID assignment. Simply omitting this field will not work due to Helm's deep merge with chart defaults. |
| resources.limits.cpu | string | `"100m"` | the cpu limit |
| resources.limits.memory | string | `"256Mi"` | the memory limit |
| resources.requests.memory | string | `"64Mi"` | the memory request |
| serviceAccount.annotations | object | `{}` | annotations to add to the service account |
| serviceAccount.create | bool | `true` | specifies whether a service account should be created |
| serviceAccount.name | string | `""` | the name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceAccount.permissionScope | string | `"cluster"` | specifies whether to create a cluster-wide permissions for the service account or namespace-scoped permissions. allowed values are: [cluster, namespace] |
| successfulJobsHistoryLimit | int | `3` | specifies the number of successful finished jobs to keep |
| tolerations | list | `[]` | tolerations for scheduling the reporter pod, e.g. to run on a dedicated or tainted node group. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)

