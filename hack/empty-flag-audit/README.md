# Empty-flag audit

Finds out what the CLI actually does when a flag is given an empty value, by
running every command-and-flag combination against a local Kosli server and
recording the result.

It exists because the answer cannot be read off the flag's type. A flag may be
refused by a required-flag check, by a value allowlist in `PreRunE`, by a
mutual-exclusion rule, by the server, or by a third party - or not refused at
all. Only running it settles which.

## Running it

```bash
make test_setup                    # a local server on localhost:8001
./hack/empty-flag-audit/audit.py        # produces results.tsv
./hack/empty-flag-audit/audit.py --ci   # produces results-ci.tsv
./hack/empty-flag-audit/report.py       # the figures the decision document quotes
./hack/empty-flag-audit/replay.py       # produces results-api.tsv
```

`audit.py` asks what the CLI does with an empty value. `replay.py` asks what the
server does with an emptied field, by capturing a request the CLI sent and
sending it again with one field emptied. It is a separate script rather than a
third mode because it measures something else: `--ci` produces rows that line up
with the plain run's, and these do not.

Narrower runs, for working on one entry:

```bash
./hack/empty-flag-audit/audit.py --only "attest generic" --flag fingerprint
```

Both passes are needed before `report.py`, which compares them to find the flags
the CLI stops checking inside CI. `report.py --write-appendices` rewrites the
flag-by-flag table in `docs/2026-08-13-empty-value-decision.md`; the
figures in that document's prose are printed but have to be copied in by hand.
Re-running the audit without re-running the report leaves the two disagreeing.

Run it from the repo root: the spec names files, such as template and policy
fixtures, by their path from there.

`spec.json` says how to make each command work and is the thing to edit. It was
written once by `bootstrap.py`, which interrogated the CLI to find a working
invocation per command. Re-running `bootstrap.py` overwrites hand edits for the
commands it covers, so correcting one command means editing its entry, not
teaching a rule in `bootstrap.py` to make an exception.

## What it measures

Each combination is run three times: with the flag left out, with a real value,
and with an empty value. Two comparisons come out of that, and they answer
different questions:

| Column | Question |
|---|---|
| `vs_omitted` | does `--flag ""` do the same as not writing the flag? |
| `vs_set` | does `--flag "$VAR"` still do its job when `VAR` is unset? |

`refused_by` says who refused an empty value: `cli` if nothing was sent, or
`downstream` if a request came back 4xx or 5xx. It is read from `--debug`, which
logs every request and its status, rather than guessed at from the wording of an
error.

An empty value counts as refused only when it failed **and** one of the other two
runs worked. Without that second condition, a command the audit could not give a
working value to would look as though it were refusing something, when in fact it
never ran properly at all.

`categories.json` says what each flag is for - identity, location, filter,
credentials, output, switch, metadata. That is a judgement, not something the
code records, and `report.py` groups by it.

## How it stays honest

**It resets the server first.** Every run empties the database, so fixture names
can be the same every time and two runs' results can be compared line by line.
The reset wipes data only - restarting the server would pull its image, which
means an AWS login, and no audit run should demand one.

**Every invocation owns its fixtures**, named after the command and the flag
under test. Two commands must not share, or one command's state decides
another's result. Two flags of one command must not share either: `create flow`
creates the flow the first time and updates it the second, and `rename flow`
renames the fixture out from under whatever runs next. The spec records `{flow}`
rather than a name, which is what lets one entry serve all of its flags.

**All three runs are set up before any of them runs**, so nothing changes
between them. A command that lists everything in the org would otherwise see the
next run's fixtures appear midway and report a difference the flag did not cause.

**Verify steps read back only this invocation's own fixtures.** A command's
output does not always show what it did - `attest generic` says "reported to
trail" whether the attestation landed on the artifact or on the trail - so
mutating commands are followed by a read whose output joins the comparison. That
read must be scoped to the fixture: `get service-account {account}`, never `list
service-accounts`. An org-wide read grows by one entry per run and reports every
flag as changing something.

**Volatile output is blanked before comparing:** uuids, timestamps, "2 minutes
ago" phrases, and the API key printed once on creation. Left in, they would
report every timestamped command as changed.

**A verdict is not a refusal.** `assert artifact` and `evaluate trail` answer a
question and say no by exiting non-zero, whatever their flags hold. Results are
compared against the other two runs rather than against exit 0, so their answers
are not mistaken for refusals.

**Commands needing a service we cannot reach** - AWS, Azure, Google Cloud,
Kubernetes, a connected git provider, Jira, SonarQube, Snyk - are recorded with a
skip reason rather than guessed at.

**Nothing may wait without explaining itself.** A run stops with a non-zero exit
when a combination takes over five seconds, or when a run retried a request until
it gave up. Every wait found so far was a defect: an AWS client waiting out a
metadata endpoint, gitlab.com throttling us, a keychain dialog waiting for a
person, and a server 500 the client kept retrying. Both checks report after the
results are written, so a run that trips one still keeps what it measured, and
both take named exceptions carrying their reason.

**Coverage is pinned to the command tree.** coverage.json is written from cobra's
own tree by a Go test, and audit.py refuses to run until spec.json covers it.
Reading --help instead used to miss whatever cobra does not list: deprecated
commands, hidden commands, and flags hidden by deprecation.

**Global flags** are declared once on the root command and behave the same
wherever they appear, so they are audited once, on `archive flow`, rather than on
every command.

## Known gaps

**The audit invents the values it gives flags.** A flag it knows nothing about
gets a made-up value, and for 114 of the 412 runnable combinations the run with
that value fails - so the empty run is compared against a control that did not
work. Written up in `docs/2026-08-15-the-audit-invents-its-input-values.md`.

**`replay.py` cannot reach a flag that is not a payload field.** The read
commands put their flags in the query string, and multipart requests are not
replayed as JSON. Of its 412 rows, 19 are an answer about the server. Written up
in `docs/2026-08-15-auditing-empty-values-at-the-api.md`.

**What a third-party service does with an empty value stays unknown.** The CLI's
own checks are visible without credentials; what AWS or Jira then does is not.
