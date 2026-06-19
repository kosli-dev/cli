---
name: new-command
description: Scaffold a new Kosli CLI command or subcommand (kosli <verb> <noun>) in this repo. Use when adding a CLI command, e.g. "add a command", "new CLI command", "scaffold a command", "create a kosli subcommand". Interviews for archetype, endpoint, flags, args, and beta/hidden status, then generates the command file, test skeleton, flag constants, and registration.
---

## Overview

This skill scaffolds a new Kosli CLI command following the repo's `kosli <verb> <noun> [flags]`
convention. It interviews you to collect the command path, archetype, API endpoint (where
applicable), positional args, flags, lifecycle status, and help-text descriptions, then
generates the command file, test skeleton, flag constants, and registration wiring - leaving
the codebase in a compiling, `--help`-renderable state with zero manual wiring remaining.

## Interview

Run these steps in order. Collect all answers before generating any files.

**Step 1 - Command path.** Ask for the verb and noun (e.g. `get trail`, `create deployment`).
Scan `cmd/kosli/` for an existing `new<Verb>Cmd` factory and the `AddCommand` block in
`root.go` to determine whether the verb already exists. If the verb is new, note that a
parent verb file (`cmd/kosli/<verb>.go`) and `root.go` wiring will also be created.

**Step 2 - Archetype.** Pre-suggest from the verb using this mapping:

| Verb(s) | Suggested archetype |
|---|---|
| `get` | read-single |
| `list` | read-list |
| `create`, `report`, `request` | create-mutate |
| `attest` | attest |
| `fingerprint`, `version`, `config` | local |
| anything else | prompt (default: generic-action if API, else local) |

The developer may override. Choosing `local` skips steps 3 (endpoint) and omits the
global-flag requirement (`RequireGlobalFlags`) entirely - no Org/ApiToken needed.

**Step 3 - API endpoint (API archetypes only; skip for `local`).** See
`references/openapi.md`. Fetch the relevant path from the OpenAPI spec, confirm the method
and path with the developer, derive `url.JoinPath` segments, and pre-fill the `Payload`
struct (for create-mutate/attest) and flag candidates (for read/list).

**Step 4 - Positional args.** Ask for names and count (mapped to `cobra.ExactArgs` or
`CustomMaximumNArgs`). For API archetypes, suggest from OpenAPI path params not covered by
global flags. For `local`, derive from the command's inputs.

**Step 5 - Flags.** For each flag collect: name, shorthand, Go type
(`String`/`Bool`/`StringSlice`/`StringToString`/`Int`...), default, requirement (required /
conditional / mutually-exclusive), and description. Suggest reusing shared helpers based on
archetype: `addDryRunFlag` (create-mutate), `addListFlags` (read-list),
`addAttestationFlags` + `addFingerprintFlags` (attest). Propose flags derived from the
OpenAPI request body or query params.

**Step 6 - Lifecycle.** Ask: beta? (adds `betaCLIAnnotation`); incubating/hidden? (adds
`Hidden: true` and `docgen.DocHiddenAnnotation`). Deprecated is out of scope for new
commands.

**Step 7 - Descriptions.** Draft `Short`, `Long`, and `Example` from the gathered
information. Use repo conventions: `^carets^` for inline code; `# title` and
backslash-continuation accordion format for examples. Present drafts for the developer to
edit before generating files.

## Routing table

After completing the interview, read the archetype-specific reference file, then `references/wiring.md`.

| Archetype | Read |
|---|---|
| local | `references/archetype-local.md` |
| read-single / read-list | `references/archetype-read.md` |
| create-mutate / generic-action | `references/archetype-mutate.md` |
| attest | `references/archetype-attest.md` |

For API archetypes, also read `references/openapi.md` for the endpoint/payload step.
For wiring (registration, flag constants, lifecycle annotations, test skeleton), read
`references/wiring.md`.

## Verification

After generating all files, run these checks in order:

1. `go build ./...` - must succeed with no errors.
2. `golangci-lint run ./cmd/kosli/...` - fallback: `go vet ./cmd/kosli/...`.
3. `go run . <verb> <noun> --help` - confirms wiring and that help renders (including the
   beta banner if marked beta).
4. Tests:
   - API archetypes: requires a local server (`make test_setup`); run
     `make test_integration_single TARGET=<Suite>`.
   - `local` archetype: typically runs without a server; run
     `go test ./cmd/kosli/ -run <Suite>` directly.

## Non-goals

- **No docs MDX.** Reference docs regenerate downstream in `kosli-dev/docs` on the next
  CLI release. Note that a release is required for docs to appear, and that beta/hidden
  commands surface via the lifecycle-notices mechanism.
- **No golden files.** Leave `golden: ""` / minimal assertions; capture golden output after
  the first real run against a server.
- **Never invent API endpoints or fields.** Consult the OpenAPI spec; if unreachable or the
  endpoint is not yet published, fall back to manual entry and explicitly flag the payload
  as unverified.
- **No deprecation handling.** Irrelevant for newly created commands.
