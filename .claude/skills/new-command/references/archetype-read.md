# Archetype: read (single and list)

Two sub-shapes; choose based on whether the command returns one object or many.

## read-single

Canonical example: `cmd/kosli/getFlow.go` — read it in full and adapt.

### Deltas

**Options struct**
- Include an `output string` field; no `Payload`.

**`PreRunE`**
- `RequireGlobalFlags(global, []string{"Org", "ApiToken"})` wrapped in `ErrorBeforePrintingUsage`.
- No other required-flag validation unless the command has additional required flags.

**`Args`**
- `cobra.ExactArgs(1)` — the single positional name (e.g. the flow name).

**Flags**
- `cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)` — no other shared helpers.

**`run` method**
- Build the URL: `url.JoinPath(global.Host, "api/v2/<resource>", global.Org, args[0])`.
- `GET` via `kosliClient.Do(&requests.RequestParams{Method: http.MethodGet, URL: url, Token: global.ApiToken})`.
- Print with `output.FormattedPrint(response.Body, o.output, out, 0, map[string]output.FormatOutputFunc{"table": print<Noun>AsTable, "json": output.PrintJson})`.
- Define a `print<Noun>AsTable(raw string, out io.Writer, page int) error` helper that unmarshals and renders.

---

## read-list

Two patterns exist; pick based on whether the endpoint is paginated.

- **Paginated (most lists):** canonical `cmd/kosli/listArtifacts.go` (or `listTrails.go`). Embed `listOptions` in your options struct and call `addListFlags(cmd, &o.listOptions)` (defined in `flags.go:84`) — it adds `--output`, `--page`, and `--page-limit`. Pass an optional custom page-limit as a third arg, e.g. `addListFlags(cmd, &o.listOptions, 20)` (see `listTrails.go`).
- **Simple (non-paginated):** canonical `cmd/kosli/listFlows.go` — adds `--output` (and filter flags like `--name`, `--ignore-case`) directly with `StringVarP`/`BoolVarP`, no `addListFlags`.

Read whichever canonical file matches and adapt.

### Deltas vs read-single

**`Args`**
- `cobra.NoArgs` (no positional argument).

**Flags**
- Paginated: `addListFlags(cmd, &o.listOptions)` as above.
- Simple: `--output` plus any filter flags added directly.

**`run` method**
- Build the base URL, then append query params via `url.Values{}` and `params.Encode()`.
- Same `GET` + `output.FormattedPrint` pattern as read-single.
- The `print<Noun>sAsTable` helper unmarshals to a `[]map[string]interface{}` and handles the empty-list case with `logger.Info("No <nouns> were found.")`.

**`RunE` signature**
- `return o.run(out)` (no `args` needed when there are no positional args).

---

## What stays the same across both sub-shapes

- `new<VerbNoun>Cmd(out io.Writer) *cobra.Command` factory with description consts.
- `ErrorBeforePrintingUsage` wrapping for `RequireGlobalFlags` errors.
- Imports: `"net/http"`, `"net/url"`, `"io"`, `output`, `requests`, `cobra`.

## Where to look next

- Endpoint path and query-param suggestions: `references/openapi.md`.
- Registration, flag constants, lifecycle annotations, test skeleton: `references/wiring.md`.
