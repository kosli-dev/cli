---
title: "Flows and Trails"
bookCollapseSection: false
weight: 200
---
# Flows and Trails Naming Conventions

This document outlines recommended naming conventions for Flows and Trails as they closely relate to each other in Kosli. Adopting these conventions will help maintain clarity and consistency across your organization.

## Flows

A clear naming convention transforms a simple ID into a meaningful identifier that everyone understands. This shared language ensures attestations go to the right place and you can track your releases from start to finish.

The naming convention relates to `FLOW-NAME` in Kosli CLI command:

```bash
kosli create flow FLOW-NAME [flags]
```

See [CLI documentation]({{< ref "client_reference/kosli_create_flow" >}}) for more details.

The following sections define conventions for the two main types of Flows in Kosli: Build Flows and Release Flows.

- **Build Flows**: Represent how code changes move from commit to artifact.
- **Release Flows**: Represent how artifacts move from binary repository to deployment.

### Build Flows

**Name Convention:** `org-unit`-`repo`-`[service]`

- **org-unit**: Your organizational unit, division or team name
- **repo**: Your repository name
- **service (Optional)**: The specific service or component that the artifact belongs to

{{% hint info %}}
You can skip `service` if your repository produces only one artifact, i.e. non-monorepo setups.
{{% /hint %}}

{{< tabs "build-flow-examples" >}}
{{< tab "snake_case" >}}

**Examples on `FLOW-NAME`:**
- `investment-web_app` (single artifact)
- `investment-web_app-frontend` (with service: frontend)
- `devops_team-mobile_app-backend` (with service: backend)

**Regex:**

```bash
^[a-z][a-z0-9_]*-[a-z][a-z0-9_]*(-[a-z][a-z0-9_]*)?$
```

{{< /tab >}}
{{< tab "camelCase" >}}

**Examples on `FLOW-NAME`:**
- `investment-webApp` (single artifact)
- `investment-webApp-frontend` (with service: frontend)
- `devopsTeam-mobileApp-backend` (with service: backend)

**Regex:**

```bash
^[a-z][a-zA-Z0-9]*-[a-z][a-zA-Z0-9]*(-[a-z][a-zA-Z0-9]*)?$
```
{{< /tab >}}
{{< tab "PascalCase" >}}

**Examples on `FLOW-NAME`:**
- `Investment-WebApp` (single artifact)
- `Investment-WebApp-Frontend` (with service: frontend)
- `DevOpsTeam-MobileApp-Backend` (with service: backend)

**Regex:**

```bash
^[A-Z][a-zA-Z0-9]*-[A-Z][a-zA-Z0-9]*(-[A-Z][a-zA-Z0-9]*)?$
```
{{< /tab >}}
{{< /tabs >}}

### Release Flows

**Name Convention:** `org-unit`-`repo`

- **org-unit**: Your organizational unit, division or team name
- **repo**: Your repository name

{{< tabs "release-flow-examples" >}}
{{< tab "snake_case" >}}

**Examples on `FLOW-NAME`:**
- `investment-web_app`
- `devops_team-mobile_app`

**Regex:**

```bash
^[a-z][a-z0-9_]*-[a-z][a-z0-9_]*$
```
{{< /tab >}}
{{< tab "camelCase" >}}
**Examples on `FLOW-NAME`:**
- `investment-webApp`
- `devopsTeam-mobileApp`

**Regex:**

```bash
^[a-z][a-zA-Z0-9]*-[a-z][a-zA-Z0-9]*$
```
{{< /tab >}}
{{< tab "PascalCase" >}}
**Examples on `FLOW-NAME`:**
- `Investment-WebApp`
- `DevOpsTeam-MobileApp`

**Regex:**

```bash
^[A-Z][a-zA-Z0-9]*-[A-Z][a-zA-Z0-9]*$
```
{{< /tab >}}
{{< /tabs >}}


## Trails

The naming convention for Trails depends on the type of Flow they are associated with: Build Flows or Release Flows and relates to `TRAIL-NAME` in Kosli CLI command:

```bash
kosli begin trail TRAIL-NAME \
  --flow FLOW-NAME \ # Build or Release Flow
  [other flags]
```

See [CLI documentation]({{< ref "client_reference/kosli_begin_trail" >}}) for more details.


### Trails associated with [Build Flows]({{< ref "#build-flows" >}})

Casing does not matter for SHA values, so we do not provide multiple casing options here.

**Name Convention:** `sha`

- **sha**: The git commit HEAD SHA that triggered the build.

**Examples on `TRAIL-NAME`:**
- `abcdef1234567890abcdef1234567890abcdef12` (full 40-char SHA)
- `abcdef123` (short SHA)

**Regex:**

```bash
^[a-f0-9]+$
```

### Trails associated with [Release Flows]({{< ref "#release-flows" >}})
**Name Convention:** `env`-`pr-number`

- **env**: The target deployment environment (e.g., staging, production)
- **pr-number**: The pull request or change request number associated with the deployment.

{{< tabs "release-trail-examples" >}}
{{< tab "snake_case" >}}
**Examples on `TRAIL-NAME`:**
- `staging-42`
- `production-108`

**Regex:**

```bash
^[a-z][a-z0-9_]*-[0-9]+$
```

{{< /tab >}}
{{< tab "camelCase" >}}
**Examples on `TRAIL-NAME`:**
- `staging-42`
- `production-108`

**Regex:**

```bash
^[a-z][a-zA-Z0-9]*-[0-9]+$
```

{{< /tab >}}
{{< tab "PascalCase" >}}
**Examples on `TRAIL-NAME`:**
- `Staging-42`
- `Production-108`

**Regex:**

```bash
^[A-Z][a-zA-Z0-9]*-[0-9]+$
```
{{< /tab >}}
{{< /tabs >}}
