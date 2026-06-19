# Archetype: local (no-API)

Canonical example: `cmd/kosli/fingerprint.go` — read it in full and adapt.

## Deltas from the canonical example

These are the things that differ for a generic local command vs. `fingerprint.go`:

**Options struct**
- Define a `<verbNoun>Options` struct with only the fields the command needs locally (no `Payload`, no API-related fields).

**`PreRunE`**
- Do local flag validation only (e.g. `RequireFlags`, `ValidateRegistryFlags` if reading an image, `MuXRequiredFlags` for mutually exclusive flags).
- **No `RequireGlobalFlags`** — local commands do not need `Org` or `ApiToken`.

**`RunE` → `o.run(args, out)`**
- Delegate to `o.run(args, out)` exactly as in `fingerprint.go`.

**`run` method**
- Does local work: reads files, computes values, calls internal packages.
- Prints results via `logger.Info(...)` or writes to `out`.
- **No `url.JoinPath`**, **no `kosliClient.Do`**, **no `DryRun`**, **no `Payload`**.

**Flags**
- Add only the flags this command needs; no `addDryRunFlag`.
- Use `RequireFlags` for mandatory ones.

**Imports**
- Typically only `"io"` and `"github.com/spf13/cobra"` at a minimum; add internal packages as needed.
- No `"net/http"`, `"net/url"`, or `requests` import.

## What stays the same

- `new<VerbNoun>Cmd(out io.Writer) *cobra.Command` factory with description consts.
- `Args: cobra.ExactArgs(N)` or `cobra.NoArgs` as appropriate.
- String consts for `Short`, `Long`, `Example` above the factory.

## Where to look next

- Registration, flag constants, lifecycle annotations, test skeleton: `references/wiring.md`.
- For a simple example of the `run` pattern without API calls, read `cmd/kosli/fingerprint.go` alongside `cmd/kosli/version.go`.
