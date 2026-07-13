# OpenAPI-driven endpoint and payload derivation

Use this file for all API archetypes (read-single, read-list, create-mutate, attest, generic-action). Skip for `local`.

---

## 1. Source

`https://app.kosli.com/api/v2/openapi.json` - public, no auth required, OpenAPI 3.1.0, ~80 paths.

**Fetch only what you need** - never dump the whole spec into context. Use `jq` to extract the single path or schema you care about.

---

## 2. Find the endpoint and method

Match the command to a spec path. Confirm the match with the developer before proceeding.

Example - extract a single path:

```bash
curl -s https://app.kosli.com/api/v2/openapi.json | jq '.paths["/flows/{org}"]'
```

List all paths to discover candidates:

```bash
curl -s https://app.kosli.com/api/v2/openapi.json | jq '.paths | keys[]'
```

Extract the schema for a specific request body:

```bash
curl -s https://app.kosli.com/api/v2/openapi.json \
  | jq '.paths["/flows/{org}"].put.requestBody.content["application/json"].schema'
```

---

## 3. Path mapping to `url.JoinPath`

Spec paths are relative to `/api/v2`. Map each segment:

| Spec path segment | CLI mapping |
|---|---|
| `/flows/{org}` | `url.JoinPath(global.Host, "api/v2/flows", global.Org)` |
| `{flow_name}` | positional arg or `--flow` flag value |
| `{trail_name}` | `--trail` flag value |
| `{environment_name}` | positional arg or `--environment` flag value |

Path params covered by global flags (`{org}` → `global.Org`, `{host}` → `global.Host`) are never exposed as CLI flags. All other path params become positional args or named flags - decide with the developer.

Full example for `PUT /flows/{org}`:

```go
url, err := url.JoinPath(global.Host, "api/v2/flows", global.Org)
```

---

## 4. Payload struct (create-mutate and attest)

Derive the `Payload` struct fields and `json:` tags directly from the request body's component schema. Do not invent field names.

If the schema references a `$ref`, resolve it:

```bash
curl -s https://app.kosli.com/api/v2/openapi.json \
  | jq '.components.schemas.FlowRequest'
```

Then write the struct with exact field names from the spec:

```go
type FooPayload struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

For the attest archetype the payload embeds `*CommonAttestationPayload` and adds a `type_name` field - see `references/archetype-attest.md` (canonical example `cmd/kosli/attestCustom.go`) for the exact shape.

---

## 5. Flag suggestions

- **read/list** - query params from the spec `parameters` array become flag candidates.
- **create-mutate** - body fields from the schema become flag candidates; required body fields map to required flags (no `[optional]` prefix).
- Shared helpers to prefer over custom flags: `addDryRunFlag`, `addListFlags`, `addAttestationFlags`, `addFingerprintFlags` - check `cmd/kosli/flags.go` before adding a custom flag.

---

## 6. Fallback - never fabricate

If the spec is unreachable or the endpoint is not yet published (common when the API change is still in flight):

1. Tell the developer the spec could not be reached or the path was not found.
2. Ask the developer to supply the endpoint path, method, and payload fields manually.
3. Mark every field in the generated `Payload` struct with a comment `// UNVERIFIED - confirm against spec`.
4. Do not invent field names, types, or JSON tags.

This upholds the project rule: never assume API response structures or field names.
