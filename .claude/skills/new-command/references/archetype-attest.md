# Archetype: attest

Canonical example: `cmd/kosli/attestCustom.go` — read it in full and adapt.

## Deltas

**Payload struct**
- Embed `*CommonAttestationPayload` plus any type-specific fields.
- Include a `TypeName string \`json:"type_name"\`` field (or equivalent identifier for the attestation type).

  ```go
  type <Noun>AttestationPayload struct {
      *CommonAttestationPayload
      TypeName string `json:"type_name"`
      // ... type-specific fields
  }
  ```

**Options struct**
- Embed `*CommonAttestationOptions`; hold a `payload <Noun>AttestationPayload`.

  ```go
  type attest<Noun>Options struct {
      *CommonAttestationOptions
      payload <Noun>AttestationPayload
  }
  ```

**Factory initialisation**
- Initialise both embedded structs explicitly:
  ```go
  o := &attest<Noun>Options{
      CommonAttestationOptions: &CommonAttestationOptions{
          fingerprintOptions: &fingerprintOptions{},
      },
      payload: <Noun>AttestationPayload{
          CommonAttestationPayload: &CommonAttestationPayload{},
      },
  }
  ```

**`PreRunE`**
1. `CustomMaximumNArgs(1, args)` — attestation commands allow 0 or 1 positional artifact arg.
2. `RequireGlobalFlags(global, []string{"Org", "ApiToken"})` wrapped in `ErrorBeforePrintingUsage`.
3. `MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, false)` — only one of these may be supplied.
4. `ValidateSliceValues(o.redactedCommitInfo, allowedCommitRedactionValues)` if the type exposes `--redact-commit-info`.
5. `ValidateAttestationArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint)`.
6. `ValidateRegistryFlags(cmd, o.fingerprintOptions)`.

**Flags**
- Call `addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)` where `ci := WhichCI()` — this adds `--flow`, `--trail`, `--name`, `--fingerprint`, `--artifact-type`, commit flags, and more.
- Add type-specific flags after `addAttestationFlags`.
- `RequireFlags(cmd, []string{"flow", "trail", "name", ...})` for type-specific required flags.

**`RunE`**
- Capture `o.repoURLExplicit = cmd.Flags().Changed("repo-url")` before delegating.

**`run` method**
- Build the URL: `url.JoinPath(global.Host, "api/v2/attestations", global.Org, o.flowName, "trail", o.trailName, "<type-slug>")`.
- Call `o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)` to resolve fingerprint and commit info.
- Load any type-specific data (e.g. a JSON file) and populate `o.payload` fields.
- Call `prepareAttestationForm(o.payload, o.attachments)` to build the multipart form; handle `cleanupNeeded` with a `defer os.Remove`.
- POST: `kosliClient.Do(&requests.RequestParams{Method: http.MethodPost, URL: url, Form: form, DryRun: global.DryRun, Token: global.ApiToken})`.
- Log success: `logger.Info("<type>:%s attestation '%s' is reported to trail: %s", o.payload.TypeName, o.payload.AttestationName, o.trailName)`.
- Return `wrapAttestationError(err)` — never return `err` directly.

**No `Args:` field on the `cobra.Command`**
- The comment in `attestCustom.go` explains why: `CustomMaximumNArgs` handles this in `PreRunE` instead, so `Args` is intentionally omitted.

## Test setup

`SetupTest` must:
1. Set `global = &GlobalOpts{...}` as usual.
2. Call `CreateFlow(...)` to ensure the flow exists.
3. Call `BeginTrail(...)` to ensure the trail exists.

See `cmd/kosli/getFlow_test.go` for the global setup pattern and `cmd/kosli/testHelpers.go` for `CreateFlow`/`BeginTrail` helpers.

## Where to look next

- Registration, flag constants, lifecycle annotations, full test skeleton: `references/wiring.md`.
- Endpoint path: `references/openapi.md` (path pattern is `POST /attestations/{org}/{flow_name}/trail/{trail_name}/<type-slug>`).
