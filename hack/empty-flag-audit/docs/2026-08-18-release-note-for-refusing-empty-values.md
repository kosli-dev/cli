# Release note: empty flag values are now refused

Text for the release that ships the refusal. The release flow drafts
`dist/release_notes.md` for a human to edit, and that file is build output, so
the wording lives here until it is needed.

## TL;DR

- Accepting an empty flag value was a bug. `--flag ""` never did what the
  command was asked to do, and usually reported success anyway.
- From this release it is an error. A pipeline that starts failing here was
  already producing a result nobody asked for: the failure is the CLI finally
  saying so, not new strictness.
- The empty value nearly always comes from a shell variable that is unset.
- The error names the flag. Give it a real value, or remove the flag: in almost
  every case an empty value did what leaving the flag out does.
- Leaving a flag out is unchanged. Defaults still apply, including the ones
  filled in from your CI environment.
- One capability goes: `--description ""` no longer clears a description on
  `update control` or `update service-account`.

## Why this is a fix, not a new restriction

A run that now fails was already doing something other than what was asked,
and the CLI could not tell you. Measured examples: `--exclude ""` excluded
nothing, so the fingerprint was one no artifact matched; `--fingerprint ""`
recorded an attestation against the trail instead of the artifact you named;
`--environment ""` attached a policy to no environment; `--redact-commit-info ""`
sent the commit author and message that the flag exists to withhold. Most of
these exited 0 and printed what success prints.

```
Error: flag '--exclude' was given an empty value
```

## What is refused

- `--flag ""` and `--flag=`
- an empty element in a comma list: `--exclude folder2,,folder3`
- a repeated flag where one value is empty: `--attachments "" --attachments a.json`
- the same values arriving from a `KOSLI_*` environment variable or `~/.kosli.yml`

## What has not changed

Leaving a flag out behaves exactly as before. The rule is about what a command
is told, not what it falls back on, so `--build-url`, `--commit-url` and
`--repository` still take their values from your CI environment when you omit
them.

## One case the CLI still cannot catch

A boolean flag written without quotes loses the empty value in the shell rather
than in the CLI: `--compliant ${UNSET}` arrives as `--compliant` with nothing
after it, which is indistinguishable from typing `--compliant` deliberately.
Quote the variable - `--compliant "${VAR}"` - and it is refused like any other
empty value.
