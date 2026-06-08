# Service Account API-key response fixtures

These JSON files are the canonical example response bodies returned by the
Service Account `api-keys` endpoints. The command tests in
`cmd/kosli/apiKey_test.go` stub the API (`httpfake`) with these fixtures
instead of inline strings, so the response contract lives in one place.

| Fixture | Endpoint / response |
|---------|---------------------|
| `created_api_key.json` | `POST .../api-keys` → `201` (create) |
| `rotated_api_key.json` | `POST .../api-keys/{key_id}/rotate` → `201` (rotate; includes `grace_period_expires_at`) |
| `listed_api_keys.json` | `GET .../api-keys` → `200` (list) |
| `revoke_success.json` | `DELETE .../api-keys/{key_id}` → `200` (bare string) |
| `error_*.json` | error envelope `{ "message": string }` (`403`/`404`) |

> **Note:** these fixtures only exercise CLI logic (flag parsing, output
> formatting, error handling). They do **not** verify the live API contract —
> if the API changes field names/types, the stubbed tests still pass. Keeping
> the bodies here makes it possible to validate them against the published
> OpenAPI schema in a separate step:
> https://app.kosli.com/api/v2/doc/
