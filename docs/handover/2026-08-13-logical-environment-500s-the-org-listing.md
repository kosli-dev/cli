# Bug: a logical environment with no included environments 500s the org's environment listing

`kosli create environment X --type logical` writes a record the server cannot
read back, unless `--included-environments` is given. From then on `list
environments` returns HTTP 500 for **every** environment in that organization,
not just the one just created, and `get environment X` returns 500 too.

The org stays in that state until the record is removed.

No empty value is needed. Creating a logical environment the plain way, with no
`--included-environments` at all, is enough.

## Reproducing it

Against a freshly reset local server:

```
$ kosli create environment inc-a --type server
environment inc-a was created

$ kosli create environment log-good --type logical --included-environments inc-a
environment log-good was created

$ kosli list environments
NAME      TYPE     LAST REPORT  LAST MODIFIED              TAGS  POLICIES
inc-a     server                2026-08-15T08:56:20+01:00        []      # fine

$ kosli create environment log-plain --type logical
environment log-plain was created                          # accepted, exit 0

$ kosli list environments
Error: [kosli list environments] Get ".../environments/docs-cmd-test-user-shared":
       giving up after 4 attempt(s)

$ kosli list environments --debug
[debug] GET .../api/v2/environments/docs-cmd-test-user-shared (status: 500): retrying in 1s (3 left)
[debug] GET .../api/v2/environments/docs-cmd-test-user-shared (status: 500): retrying in 2s (2 left)
[debug] GET .../api/v2/environments/docs-cmd-test-user-shared (status: 500): retrying in 4s (1 left)
```

`kosli get environment log-plain` returns 500 in the same way. Everything else
keeps working: `get environment log-good`, `get environment inc-a`, `list flows`
and the `create` commands are all unaffected.

An empty `--included-environments ""` does the same thing, because an empty
value and an absent flag send the same payload. That is how this was first
found, and it is why it was first written up as a bug about the flag.

## What is happening

From the server's log for the failing request:

```
File "/app/src/fastapi_app/common/environment.py", line 201, in list_environments
  return [_env_doc(env, compliance) for env in env_list]
File "/app/src/fastapi_app/common/environment.py", line 223, in _env_doc
  doc = env.json
File "/app/src/model/logical_environment.py", line 35, in json
  del doc[key]
KeyError: 'included_environments_ids'
```

The key is absent, not empty. Rendering a logical environment deletes
`included_environments_ids` from the document, and a logical environment created
without included environments never had it. The listing renders every
environment in the org in one comprehension, so one such record fails all of
them.

## Why it matters

- **It needs no mistake to trigger.** `--included-environments` is optional, so
  creating a logical environment without one is ordinary use, not a typo or an
  unset variable.
- **The damage is org-wide, not record-scoped.** One such record stops anyone in
  the organization listing any environment.
- **Nothing reports it at the time.** The create exits 0 with "environment
  log-plain was created".

## Suggested fixes

The empty-value rule proposed in
`../../empty-flag-audit/docs/2026-08-13-empty-value-decision.md` does **not**
fix this. Refusing an empty `--included-environments` closes one route to the
bad record and leaves the ordinary route open. This one is the server's, and
either of these settles it:

1. Rendering an environment tolerates the key being absent, so one record cannot
   take out the listing for everyone.
2. Creating a logical environment writes `included_environments_ids`, empty when
   nothing is included, so every record has the key the renderer removes.

## How it was found

By the empty-flag audit, in the `empty-flag-audit/` directory of the CLI
repository, which runs every command-and-flag combination with an empty value. It stood out because it was
the only case where an empty value broke a command other than the one being run.

The wider trigger was found later, by the audit's own retry check: `list
environments` was retrying a 500 to exhaustion in the middle of a run, long
before the empty-value combination for `create environment` was measured. Other
commands' setup steps create plain logical environments, so the org was already
poisoned. What had looked like a bug about one flag turned out to be a bug about
the default way to create a logical environment.
