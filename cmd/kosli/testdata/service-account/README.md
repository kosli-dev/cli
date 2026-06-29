# Service Account response fixtures

These JSON files are the canonical example response bodies returned by the
Service Account endpoints (the accounts themselves and their `api-keys`). The
command tests in `cmd/kosli/apiKey_test.go` and `cmd/kosli/serviceAccount_test.go`
stub the API (`httpfake`) with these fixtures instead of inline strings, so the
response contract lives in one place.

### Service account `api-keys` endpoints

| Fixture | Endpoint / response |
|---------|---------------------|
| `created_api_key.json` | `POST .../{name}/api-keys` → `201` (create) |
| `rotated_api_key.json` | `POST .../{name}/api-keys/{key_id}/rotate` → `201` (rotate; includes `grace_period_expires_at`) |
| `listed_api_keys.json` | `GET .../{name}/api-keys` → `200` (list) |
| `revoke_success.json` | `DELETE .../{name}/api-keys/{key_id}` → `200` (bare string) |

### Service account management endpoints

| Fixture | Endpoint / response |
|---------|---------------------|
| `created_service_account.json` | `POST /service-accounts/{org}` → `201` (create) |
| `listed_service_accounts.json` | `GET /service-accounts/{org}` → `200` (list; one account per privilege: member, admin, snapshotter, reader) |
| `service_account.json` | `GET /service-accounts/{org}/{name}` → `200` (get) |
| `updated_service_account.json` | `PATCH /service-accounts/{org}/{name}` → `200` (update) |
| `delete_success.json` | `DELETE /service-accounts/{org}/{name}` → `200` (bare `"OK"`) |

### Shared

| Fixture | Endpoint / response |
|---------|---------------------|
| `error_*.json` | error envelope `{ "message": string }` (`403`/`404`) |

> **Note:** these fixtures only exercise CLI logic (flag parsing, output
> formatting, error handling). They do **not** verify the live API contract —
> if the API changes field names/types, the stubbed tests still pass. Keeping
> the bodies here makes it possible to validate them against the published
> OpenAPI schema in a separate step:
> https://app.kosli.com/api/v2/doc/
