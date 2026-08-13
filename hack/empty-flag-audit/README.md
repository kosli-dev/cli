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
make test_setup                       # a local server on localhost:8001
./hack/empty-flag-audit/audit.py      # produces results.tsv
./hack/empty-flag-audit/audit.py --ci # produces results-ci.tsv
./hack/empty-flag-audit/report.py     # the figures the decision document quotes
```

Narrower runs, for working on one entry:

```bash
./hack/empty-flag-audit/audit.py --only "attest generic" --flag fingerprint
```

Both passes are needed before `report.py`, which compares them to find the flags
the CLI stops checking inside CI. `report.py --write-appendices` rewrites the
flag-by-flag table in `docs/handover/2026-08-13-empty-value-decision.md`; the
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
Kubernetes, a connected git provider, Jira, SonarQube, Snyk, a credentials store
- are recorded with a skip reason rather than guessed at.

**Global flags** are declared once on the root command and behave the same
wherever they appear, so they are audited once, on `archive flow`, rather than on
every command.

## Known gaps

Hidden commands are invisible to `--help`, so `bootstrap.py` never found them and
they are absent from `spec.json`. `attest override` is one, and it carries
`--new-compliance-status`.
