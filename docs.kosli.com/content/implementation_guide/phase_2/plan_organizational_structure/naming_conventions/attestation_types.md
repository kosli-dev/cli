---
title: "Attestation Types"
bookCollapseSection: false
weight: 300
---

# Attestation Types

Use clear, descriptive names for custom attestation types to indicate what kind of evidence they represent.

Naming convention relates to `TYPE-NAME` in Kosli CLI command:

```bash
kosli create attestation-type TYPE-NAME [flags]
```

See [CLI documentation]({{< ref "client_reference/kosli_create_attestation-type" >}}) for more details.

**Name Convention:** `control objective`-`evidence type`-`[detail]`-`[version]`

- **control objective**: The high-level control or requirement the attestation supports (e.g., control id, code review, security scan, unit test)
- **evidence type**: The specific type of evidence being attested (e.g. tool-name, test-suite)
- **detail (Optional)**: Additional context or detail about the attestation (e.g., type, severity-level, environment, etc.)
- **version (Optional)**: The version of the attestation type or schema. Should follow semantic versioning (e.g., v1, v2)

{{% hint info %}}

- `detail` element may be repeated to add finer granularity if needed.
- You can skip `detail` and `version` if not needed for your use case.
- Kosli versions attestation types automatically, so `version` is often unnecessary. However, it can be useful for multiple version running at the same time, for example in shared pipelines.
{{% /hint %}}

{{< tabs "attestation-type-examples" >}}
{{< tab "snake_case" >}}

**Examples on `TYPE-NAME`:**
- `bc1-version_control-v1` (BC1 version control attestation, version 1)
- `code_review-github-pr` (basic code review attestation)
- `security_scan-snyk-high` (Custom schema for Snyk scan with high severity detail)
- `unit_test-junit-detail1-detail2-v2` (Multiple detail blocks with version)

**Regex:**

```bash
^[a-z][a-z0-9_]*-[a-z][a-z0-9_]*(-[a-z][a-z0-9_]*)*(-v[1-9][0-9]*)?$
```

{{< /tab >}}
{{< tab "camelCase" >}}
**Examples on `TYPE-NAME`:**
- `bc1-versionControl-v1` (BC1 version control attestation, version 1)
- `codeReview-github-pr` (basic code review attestation)
- `securityScan-snyk-high` (Custom schema for Snyk scan with high severity detail)
- `unitTest-junit-detail1-detail2-v2` (Multiple detail blocks with version)

**Regex:**

```bash
^[a-z][a-zA-Z0-9]*-[a-z][a-zA-Z0-9]*(-[a-z][a-zA-Z0-9]*)*(-v[1-9][0-9]*)?$
```
{{< /tab >}}
{{< tab "PascalCase" >}}

**Examples on `TYPE-NAME`:**
- `Bc1-VersionControl-V1` (BC1 version control attestation, version 1)
- `CodeReview-Github-Pr` (basic code review attestation)
- `SecurityScan-Snyk-High` (Custom schema for Snyk scan with high severity detail)
- `UnitTest-Junit-Detail1-Detail2-V2` (Multiple detail blocks with version)

**Regex:**

```bash
^[A-Z][a-zA-Z0-9]*-[A-Z][a-zA-Z0-9]*(-[A-Z][a-zA-Z0-9]*)*(-V[1-9][0-9]*)?$
```
{{< /tab >}}
{{< /tabs >}}