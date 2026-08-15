# Bug: an empty --included-environments makes `list environments` return 500 for the whole org

`kosli create environment` accepts an empty `--included-environments` and writes a
record that the server cannot read back. From then on `list environments` returns
HTTP 500 for **every** environment in that organization, not just the one just
created, and `get environment` on that one also returns 500.

The org stays in that state until the record is removed.

## Reproducing it

Against a freshly reset local server:

```
$ kosli create environment inc-a --type server
environment inc-a was created

$ kosli create environment log-good --type logical --included-environments inc-a
environment log-good was created

$ kosli list environments
NAME      TYPE     LAST REPORT  LAST MODIFIED  TAGS  POLICIES
...                                                              # fine

$ kosli create environment log-bad --type logical --included-environments ""
environment log-bad was created                                  # accepted, exit 0

$ kosli list environments
Error: [kosli list environments] Get ".../environments/docs-cmd-test-user-shared":
       giving up after 4 attempt(s)

$ kosli list environments --debug
[debug] GET .../api/v2/environments/docs-cmd-test-user-shared (status: 500): retrying in 1s (3 left)
[debug] GET .../api/v2/environments/docs-cmd-test-user-shared (status: 500): retrying in 2s (2 left)
[debug] GET .../api/v2/environments/docs-cmd-test-user-shared (status: 500): retrying in 4s (1 left)
```

`kosli get environment log-bad` returns 500 in the same way. Other endpoints are
unaffected: `get environment inc-a`, `list flows` and the `create` commands all
keep working.

## What is probably happening

`--included-environments` is a comma-separated list. An empty value produces a
list holding one empty string rather than an empty list, so the record names an
environment called "". Reading the environment back resolves the names it
includes, and resolving "" fails.

That is a guess from the outside; the server logs will say.

## Why it matters

- **An unset variable is enough to trigger it.** `--included-environments
  "${ENVS}"` with `ENVS` unset is all it takes, and the command reports success.
- **The damage is org-wide, not record-scoped.** One bad record stops anyone in
  the organization listing any environment.
- **Nothing reports it at the time.** The create exits 0 with "environment
  log-bad was created".

## Suggested fixes

Any one of these stops the org-wide failure; the first two are worth having anyway.

1. The CLI refuses an empty `--included-environments`. This is the general rule
   proposed in `../../empty-flag-audit/docs/2026-08-13-empty-value-decision.md`, and this bug is one case of it.
2. The server rejects an included-environment name that is empty, at write time.
3. Reading an environment tolerates a name that does not resolve, so one bad
   record cannot take out the listing for everyone.

## How it was found

By the empty-flag audit in `empty-flag-audit`, which runs every
command-and-flag combination with an empty value. This one stood out because it
was the only case where an empty value broke a command other than the one being
run.
