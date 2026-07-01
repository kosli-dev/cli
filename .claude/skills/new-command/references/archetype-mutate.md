# Archetype: create-mutate / generic-action

Two sub-shapes. Use create-mutate for straightforward PUT/POST commands that send a JSON payload. Use generic-action when the command gathers data (e.g. from the local environment or another service) before calling the API and does not fit the read or attest shapes.

## create-mutate

Canonical example: `cmd/kosli/createFlow.go` — read it in full and adapt.

### Deltas

**Payload struct**
- Define a `<Noun>Payload` struct with fields and `json:` tags derived from the OpenAPI request body schema (see `references/openapi.md`).
- Embed it in the options struct: `payload <Noun>Payload`.

**Options struct**
- Holds `payload <Noun>Payload` and any extra fields that are command-local (e.g. a file path that is processed before being added to the payload).

**`PreRunE`**
- `RequireGlobalFlags(global, []string{"Org", "ApiToken"})`.
- Add `MuXRequiredFlags` calls for any mutually exclusive flag pairs (see `createFlow.go` for the template/template-file example).
- Add `RequireFlags` for any flags that are always required (beyond global flags).

**Flags**
- Bind each payload field to a flag: `cmd.Flags().StringVar(&o.payload.FieldName, "flag-name", "", flagConstant)`.
- Call `addDryRunFlag(cmd)` — dry-run support is expected on all mutate commands.

**`run` method**
- Build the URL: `url.JoinPath(global.Host, "api/v2/<resource>", global.Org, ...)`.
- Send with `kosliClient.Do(&requests.RequestParams{Method: http.MethodPut, URL: url, Payload: o.payload, DryRun: global.DryRun, Token: global.ApiToken})`.
- For multipart form uploads (e.g. when a file is involved), build a `[]requests.FormItem` and use `Form: form` instead of `Payload`.
- Log success: `if err == nil && !global.DryRun { logger.Info("<noun> '%s' was created", ...) }`.

**Imports**
- `"net/http"`, `neturl "net/url"` (alias to avoid collision with `url` variable), `requests`, `cobra`, `"io"`.

---

## generic-action fallback

Use when the command does significant local work (snapshot, diff, report) before or after the API call. There is no single canonical file — look at the closest example in `cmd/kosli/` for the verb you are implementing.

Key differences from create-mutate:
- May have no `Payload` struct if the body is built dynamically.
- `PreRunE` may be heavier (e.g. `ConditionallyRequiredFlags`, environment-specific validation).
- `run` may call multiple internal packages before building `reqParams`.

For everything else (global flags, dry-run, logging, error handling), follow the create-mutate pattern.

---

## What stays the same

- `new<VerbNoun>Cmd(out io.Writer) *cobra.Command` factory with description consts.
- `ErrorBeforePrintingUsage` wrapping for `RequireGlobalFlags` errors.
- `addDryRunFlag(cmd)` is always called.

## Where to look next

- Endpoint path and payload field derivation: `references/openapi.md`.
- Registration, flag constants, lifecycle annotations, test skeleton: `references/wiring.md`.
