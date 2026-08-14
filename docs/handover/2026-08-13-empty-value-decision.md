# Decision: may a flag ever be given an empty value?

I want to make "this flag was given an empty string" an error across the CLI, so
that an unset shell variable cannot quietly change what a command does.

## TL;DR

- An unset shell variable can silently change what a Kosli command does.
- An audit found ten cases, all exiting 0 and printing what success prints.
- Seven of them change a compliance answer. 
- One 500s an org's environment listing.
- The space is too large to check case by case.
- Proposal: an empty value is always an error, rolled out in four steps.
- The first roll out step is logging warnings to app.kosli.com.
- Cost: two `--description` flags can no longer clear a description.

## Why we cannot establish what every empty value does

It is impractical to work out what every command does with every empty flag
value. The CLI has 91 commands declaring 152 distinct flag names between them,
which is already 653 command-and-flag combinations, and that count is the floor
rather than the ceiling:

- **The answer is not in the flag's type.** A flag may be refused by a
  required-flag check, by a value allowlist in `PreRunE`, by a mutual-exclusion
  rule, by the server, or by a third party - or not refused at all. `--page` is
  refused everywhere and `--fingerprint` is accepted on some commands, and no one
  decided either; it fell out of how each was declared.
- **The same flag name behaves differently on different commands.** 21 of the
  152 are refused on some commands and accepted on others.
- **Some differences never reach the output.** `kosli attest generic` prints
  `generic attestation 'unit-test' is reported to trail: my-trail` and exits 0
  whether the attestation went to the artifact or to the trail. The difference is
  only visible in what was stored.
- **The rules change inside CI.** A flag is checked because it is marked
  required, and that marking is skipped when the flag's default comes from a CI
  environment variable. Two combinations we know of stop being checked by the CLI
  inside GitHub Actions. Appendix 1 shows this happening.
- **Flags interact.** `kosli attest generic --fingerprint ""` silently attests to
  the trail; add `--artifact-type file` and the same empty value is refused with
  `only one of --fingerprint, --artifact-type is allowed`. One flag decides
  whether another is checked, so the real space is much larger than 653.
- **Some commands need what we do not have.** 279 of the 653 combinations are on
  21 commands needing AWS, Azure, Google Cloud, Kubernetes, a connected git
  provider, Jira, SonarQube, Snyk, or a credentials store. What the CLI does
  with them can still be seen; what the service does cannot.

## What was measured instead

One slice of that space, taken by running the CLI rather than reasoning about it.
The audit is in `hack/empty-flag-audit`. The slice is: one flag emptied at a
time, every other flag held at a value that works, against a local server. The
commands needing a service we cannot reach are run too - they fail, but their own
checks run before they get that far, so what the CLI does with an empty value is
still visible.

Within that slice it is thorough. Each command is run three times - with the flag
left out, with a real value, and with an empty value - and the exit codes, the
output, and the state left on the server afterwards are compared. The whole set
runs twice, once as on a laptop and once with the environment variables GitHub
Actions sets.

It ran all 653. For 374 of them the command works here, so the whole story is
visible:

| What happens to an empty value | On a laptop | Inside GitHub Actions |
|---|---|---|
| the CLI refuses it | 203 | 201 |
| the CLI accepts it and the server refuses it | 5 | 7 |
| nothing refuses it | 166 | 166 |

Of the 166 that nothing refuses, 162 do exactly what omitting the flag does. That
is not the same as doing nothing: omitting a flag has a meaning of its own, and it
is rarely the meaning someone had in mind when they wrote that flag with a
variable after it.

Appendix 2 groups the 152 flag names by what they are for and says how many of
each kind the CLI already refuses. Appendix 3 is the same thing flag by flag.

The other 279 are on commands needing AWS, Azure, a git provider and the rest,
which cannot succeed here. Their own checks still run first, though, so the half
this decision is about is visible even without the credentials:

| What the CLI does with an empty value | On a laptop | Inside GitHub Actions |
|---|---|---|
| refuses it | 173 | 172 |
| does not react, so it reaches the service | 88 | 89 |
| lets it through, exit 0 | 18 | 18 |

The middle row is the one to look at. Those 88 are empty values that survive every
check the CLI has and are handed to someone else. Among them are the credentials:
`--aws-key-id` and `--aws-secret-key` are refused on none of their three commands,
`--jira-pat` on neither of its, and `--registry-password` and
`--registry-username` on 6 of their 16 - so the same registry credential is
checked on six commands and not on the other ten.

Most credential flags are refused, though. `--github-token`, `--gitlab-token`,
the `--bitbucket-*` and `--azure-*` pair, `--jira-api-token`, `--sonar-api-token`
and `--kubeconfig` are all refused everywhere they appear. Which are and which
are not follows no pattern anyone chose.

What those services then do with the empty values they are handed is the only
thing here that stays unknown, and no amount of work on this machine will show it
- the release plan below is what turns customers' runs into that measurement.

## What it found

That one slice was enough. In it, a single unset variable can make a flow require
none of the attestations it was meant to, put an attestation on the wrong thing,
answer a compliance question about the wrong flow, leave a policy governing
nothing, and take out an organization's environment listing with a 500. It also
turned up a bug that needs no empty value at all.

Ten of them, below. Each was re-run on its own, outside the audit, and its result
read from the server - not taken from the audit's classification, and not read
off the code. Every one exits 0 and prints what success prints. `$VAR` means a variable that is unset,
which is all it takes.

| Command | With the variable unset | |
|---|---|---|
| `kosli create flow F --template-file "$VAR"` | the flow is created with the default template, so it requires none of the attestations intended | changes compliance |
| `kosli begin trail T --flow F --template-file "$VAR"` | the trail requires nothing, so it is compliant whatever is attested to it | changes compliance |
| `kosli attest generic --fingerprint "$VAR"` (and `custom`, `decision`, `junit`) | the attestation is recorded against the trail instead of the artifact. The artifact you meant to attest has no such attestation | changes compliance |
| `kosli attest override --fingerprint "$VAR"` | the override is recorded against the trail. The artifact keeps the compliance status you meant to override | changes compliance |
| `kosli assert artifact --fingerprint "$F" --flow "$VAR"` | the verdict is about whichever flow the fingerprint is found in, not the flow you asked about | changes compliance |
| `kosli attach-policy P --environment "$VAR"` | the policy is attached to no environment. Anything deployed there is judged without it | changes compliance |
| `kosli detach-policy P --environment "$VAR"` | the policy is detached from no environment, so it stays in force | changes compliance |
| `kosli create environment E --type logical --included-environments "$VAR"` | the record cannot be read back, and `list environments` returns HTTP 500 for every environment in the org until it is removed | **outright bug**, written up in `2026-08-13-included-environments-500.md` |
| `kosli begin trail T --flow F` with no `--description` at all | the description the trail already had is replaced by an empty one, even though the server's update is written to leave an absent field alone | **an inconsistency**, not an empty value at all, written up in `2026-08-13-upsert-overwrites-unmentioned-fields.md` |
| `kosli list environments --tag "$VAR"` | answers "No environments were found", identical to a real no-match, exit 0 | wrong answer |

Ten findings, from one slice of one CLI, all of them silent. What the rest of the
space holds we do not know - and that is the argument. We cannot keep finding
these one at a time, and a rule that refuses an empty value everywhere removes
the whole class rather than the ten instances of it we happened to reach.

Two of them are worth filing on their own account, whatever is decided here, and
each has its own write-up alongside this one.

The last of those needs a word, because it is not what it first looked like.
`kosli create flow` is a create-or-update, and it reads its flags the way
`kubectl apply` reads a manifest: what you pass is what the flow becomes, so an
absent `--description` meaning "no description" is that model working, not a bug.
`begin trail` describes itself the same way, but its server side is written to
leave an absent description alone. Two commands of the same shape, two answers.
The bug is the disagreement, and it is worth settling on its own account: until
the CLI and server agree on what an absent flag means, "empty" and "absent"
cannot be given consistent meanings either.

## What this proposal cannot fix

A boolean flag is only safe when the variable is quoted. Unquoted, the shell
deletes the word before the CLI starts, so there is nothing left to refuse:

```
kosli attest generic --compliant "${NOPE}"   # Error: flag '--compliant' was given an empty value
kosli attest generic --compliant ${NOPE}     # "is_compliant": true
```

The second is indistinguishable from someone deliberately typing `--compliant`.
It matters most on the flags carrying a verdict - `--compliant`,
`--new-compliance-status`, `--no-assert`. The only fix is to stop accepting the
bare form on those flags, requiring `--compliant=true`, which is a separate
decision.

## What a CLI rule cannot reach

Not every request arrives through the CLI. A customer calling the API directly
keeps all of this, and the warnings in steps 1 and 2 will never see them, so the
measurement they produce covers CLI traffic only.

The server cannot close that gap by copying the rule, because it mostly never
sees an empty value. When an empty value produces the same request as omitting
the flag - which is 162 of the 166 - the server receives an identical payload
either way, and there is nothing empty in it to reject.

Those 162 fall into three kinds, by whose problem the payload turns out to be.
One of each, read with `--debug`:

| Command | What reaches the server | Whose problem |
|---|---|---|
| `kosli attest generic --fingerprint ""` | no fingerprint at all, sent to the trail-scoped endpoint `/attestations/{org}/{flow}/trail/{trail}/generic` | the CLI's. It picked that endpoint. Someone calling the API chooses the artifact endpoint or the trail endpoint deliberately, and is not misled. The rule settles this one |
| `kosli create environment --included-environments ""` | a logical environment with **no** `included_environments` field, which the server accepts with a 201 and then cannot read back. Omitting the flag sends the same thing | neither's, for this decision. It is a plain defect, being fixed on its own account |
| `kosli begin trail` | `"description": ""`, whether or not `--description` was passed | shared. The server's update is written to leave an absent description alone, and the CLI defeats that by always sending one. A direct caller that sends `""` gets the same result, so the CLI fix alone does not settle it |

So the server's half of this is not the CLI rule repeated. The CLI rule is
"refuse an empty value". The server's is: **an absent field means what the
command's verb says it means, and the CLI and the server must agree about that.** Deciding it command by command would be ninety-one decisions
nobody will remember, which is how the CLI got inconsistent in the first place.
The verbs already promise it:

| Verb | What an absent field means | What the CLI does today |
|---|---|---|
| `create`, `begin` | apply - what you pass is what the object becomes | `create flow` replaces the description, as it should |
| `update` | patch - the fields you name change, the rest are left alone | `update control --name` leaves the description, and the record's `version` goes 1 to 2 |

`begin trail` is the one that does not fit. It is a `begin`, so apply, but its
server side leaves an absent description alone while the CLI sends one anyway.
Neither model is wrong; having both, one per layer, is.

Settling that covers every client, including the ones the warnings cannot see,
which makes the server work in step 2 more than a precondition for the CLI
change - it is the part that closes the hole for everyone.

It also explains the two flags this proposal costs. `update control
--description` and `update service-account --description` clear on purpose
because `update` is patch: omitting the flag cannot clear anything there, so an
empty value is the only way to say it. That is why those two need
`--clear-description` and nothing else does.

## What refusing an empty value would cost

Two flags, and only two, use an empty value to mean something you cannot say any
other way: `kosli update control --description ""` and `kosli update
service-account --description ""` clear a description on purpose. Refusing an
empty value takes that away.

Clearing would then need a flag that says so - `--clear-description` - rather
than an empty value that might be a mistake. That is straightforward to add, and
it keeps clearing something you cannot do by accident. It is not proposed here.

Everywhere else, nothing is lost: an empty value either does what omitting the
flag does, or does something nobody asked for.

## Proposal

Empty is always an error, on every flag, with no exemptions to remember.

It works because it separates two intents that are spelled identically today: "I
meant to pass a value and my variable was empty" and "I want no value". They look
the same, and no amount of checking can tell them apart after the fact.

Most breaks will be where a variable is genuinely unset, and in those pipelines
the command is already doing something its author did not ask for. There the
failure replaces silent wrong behaviour rather than correct behaviour.

Not every break is that, though. Someone may be passing an empty value
deliberately, with no variable involved. They lose nothing they cannot still get
by omitting the flag, but their pipeline fails until somebody edits it. That is a
real cost, and it is why this needs a release plan rather than just a fix.

## Releasing this

Commands that succeed today will eventually start failing, so the error itself
is a breaking change and cannot go out in a 2.x release. Two things make that
sharper than usual:

- Most of the breakage lands on flags where an empty value changed nothing worth
  noticing, simply because those are most of the flags. They need no decision,
  but they do need a release note.
- Anyone tracking the latest CLI gets the new behaviour without editing
  anything, so "breaking" here means pipelines failing with no change on their
  side.

We are on v2.36.5, and #1059 collects breaking changes for v3. This does not have
to join that batch, and I do not think it should:

- **A customer whose pipeline breaks should be able to read one release note and
  know why.** A major version carrying ten unrelated breaks cannot tell them
  that.
- **The v3 batch has been accumulating for a long time.** Tying this to it means
  the compliance holes above stay open until everything else in it is ready.
- **The right moment for this one is knowable on its own.** Step 2 reports how
  often empty values actually occur, so we can see when the impact has fallen
  far enough to flip the switch. That signal says nothing about whatever else is
  queued for v3.

So: this becomes its own major release, and the changes currently queued for v3
become the one after. Major versions are cheap; a release note nobody can act on
is not.

If we add `--clear-description` at the same time, the one capability this removes
comes back, and that part stops being breaking at all.

### Proposed: four steps

At least 184 combinations changing at once is a lot to ask of customers in one
upgrade - 166 on the commands that work here, and 18 more on the ones needing a
service, where the CLI lets an empty value through before the service is even
reached. So the rule arrives in stages.

#### Step 1: somewhere to put a warning, in app.kosli.com

Nobody reads warnings in a CI workflow run. A step that only prints one is not a
migration, it is a delay, and we would reach step 3 knowing no more than we do
now. So before the CLI warns about anything, there has to be somewhere for the
warning to go.

A warning goes to two places, and no more than two:

1. **The workflow run**, printed as now.
2. **app.kosli.com, at the org level.** A command that has `--org` and
   `--api-token` can send the warning whatever else it was doing, so this covers
   163 of the 166. The exception is `kosli fingerprint`, which is entirely local
   and needs no credentials.

This is work in app.kosli.com: somewhere to receive the warnings, and one place
per org to see them.

#### Step 2: the CLI warns, in a 2.x release, breaking nothing

Fix the two bugs above - the absent-flag disagreement and the
`--included-environments` 500 - and make every empty value print a warning naming
the flag, and report it. Nothing starts failing, and anyone whose pipeline has an
unset variable can see it and fix it before it costs them anything.

The description fix has an order to it: the server must accept a payload with no
`description` before the CLI stops sending one, or every `kosli create flow`
fails. Its write-up has the detail.

#### Step 3: the warning becomes the error, in a major release of its own

One guard, one migration, one release note, and nothing else breaking in the same
version. Ship `KOSLI_ALLOW_EMPTY_FLAG_VALUES=true` alongside it as an escape
hatch, so anyone caught out has a one-line unblock while they fix the pipeline,
and remove it in the next major release.

When to ship it is a question step 2 answers: when the reported warnings have
fallen far enough that the remaining breakage is small and known.

#### Step 4: delete what the guard replaced

The bespoke emptiness checks scattered through the commands become redundant once
one rule covers every flag. This is the step that is easiest to skip and the
reason the CLI is inconsistent today, so it belongs in the plan rather than in
someone's memory.

### Why steps 1 and 2 come first

Reporting warnings is not only a kindness to customers. It answers the question a
release note cannot: how much would step 3 actually break? Today that is an
argument. With this it becomes a number, per org, and we can tell the customers
who are affected before it lands rather than after.

It also reaches where this audit could not. For the 279 combinations on commands
needing AWS, Azure, a git provider and the rest, we can see what the CLI does but
not what the service does with the 88 empty values the CLI hands over. Customers
have the credentials we lack, so their warnings are the only place that half can
be found out.

Three conditions on the reporting. Sending a warning must never fail the command
that produced it: this is a report, not a check. It has to stay rare enough to be
worth looking at, which the numbers here suggest it will be. And it must go to
the `--host` the command was already given, through the same `--http-proxy` - not
to a hardcoded address.

That last one is what makes egress restrictions a non-issue. A command reporting
a snapshot from inside a locked-down network can already reach the host it
reports to, because that is what it is doing; a warning to the same host needs
nothing that the command did not already need. Posting somewhere else would break
exactly the customers who are most careful about what leaves their network.

---

## Appendix 1: why the CLI checks less inside CI

A flag refuses an empty value when it is marked required, and `RequireFlags` only
marks a flag required when its default is empty. Several defaults are filled in
from CI environment variables, so inside CI those flags are not marked required
and the CLI stops checking them:

```
# on a laptop
kosli attest artifact ... --build-url ""
Error: flag '--build-url' is required, but empty string was provided

# in GitHub Actions
kosli attest artifact ... --build-url ""
Error: Input payload validation failed: map[build_url:Input should be a valid URL, input is empty]
```

That moves two combinations - `attest artifact --build-url` and `attest artifact
--commit-url` - from being refused by the CLI to being refused by the server.
Both are still refused here, because this server happens to check them. The point
is that the CLI stopped, in exactly the place these bugs happen. That is cli#1088.

---

## Appendix 2: the flags grouped by what they are for

The question worth asking of each flag is what it means to the person running the
command, and whether an empty value means anything for that. Grouped that way,
the 152 names are:

| What the flag is for, with a few examples | Names | Does an empty value mean anything? | CLI always refuses | Only the server refuses | Refuses on some commands | Never refuses |
|---|---|---|---|---|---|---|
| identity and selection - `--flow`, `--trail`, `--fingerprint`, `--name` | 46 | no. There is no artifact called "" | 19 | 1 | 14 | 12 |
| location and input - `--template-file`, `--results-dir`, `--paths` | 26 | no. There is no file called "" | 17 | 1 | 3 | 5 |
| filters - `--exclude`, `--namespaces`, `--services`, `--attestations` | 22 | no. Filtering on "" filters on nothing | 1 | 0 | 0 | 21 |
| credentials - `--github-token`, `--aws-secret-key`, `--registry-password` | 17 | no. There is no token "" | 12 | 0 | 2 | 3 |
| output and paging - `--output`, `--sort`, `--page`, `--reverse` | 14 | no | 9 | 0 | 1 | 4 |
| behaviour switches - `--dry-run`, `--assert`, `--compliant` | 11 | no | 11 | 0 | 0 | 0 |
| free-text metadata - `--description`, `--comment`, `--reason`, `--tag` | 9 | **sometimes** | 4 | 0 | 1 | 4 |
| the global flags - `--org`, `--api-token`, `--host`, `--debug` | 7 | no | 6 | 0 | 0 | 1 |
| **total** | **152** | | **79** | **2** | **21** | **50** |

The credentials row is the one to look at twice. Twelve of its seventeen names
are refused by the CLI everywhere they appear. Three never are - `--aws-key-id`,
`--aws-secret-key` and `--jira-pat` - so an empty one is handed to AWS or to Jira,
and whether it is caught then depends on that service and not on us. The last two,
`--registry-password` and `--registry-username`, are refused on six commands and
not on the other ten. Kosli's own
`--api-token` is not among them: it is a global flag, and it already refuses an
empty value with `--api-token is not set`.

---

## Appendix 3: every flag in the CLI

All 152 flag names, what each is for, how many of its commands were run, and what
happened. Commands needing AWS, Azure, Google Cloud, Kubernetes, a connected git
provider, Jira, SonarQube, Snyk or a credentials store are included: they cannot
succeed here, but their own checks run first, so whether the CLI refuses an empty
value is still visible.

"Always" means every measured command refused it. "Some commands" means it was
refused on some commands and not on others. "Only the server" means the CLI sent
the empty value and the server rejected it, so the CLI itself does not check it.

"But in CI only the server does" means the CLI checks the flag on a laptop and
stops checking it inside CI, because the flag's default is then filled in from a
CI environment variable, which is what stops it being marked required. Two flags
are measured that way, and both are still caught, by the server. "And not at all
in CI" would mean nothing catches it there; no flag measured that way.

Categories are a judgement, not something the code records. `--tag` filters on
`list` commands and is metadata on `create control`, for instance, and it is
listed once.

| Flag | What it is for | Commands measured | Refuses an empty value |
|---|---|---|---|
| `--annotate` | metadata | 13 of 13 | always |
| `--api-token` | global | 1 of 1 | always |
| `--archived` | filter | 1 of 1 | always |
| `--artifact-type` | identity | 16 of 16 | some commands |
| `--assert` | switch | 9 of 9 | always |
| `--assume-yes` | switch | 2 of 2 | always |
| `--attachments` | identity | 11 of 11 | always |
| `--attestation-data` | identity | 1 of 1 | always |
| `--attestation-id` | identity | 1 of 1 | never |
| `--attestations` | filter | 2 of 2 | never |
| `--aws-key-id` | credentials | 3 of 3 | never |
| `--aws-region` | location | 3 of 3 | some commands |
| `--aws-secret-key` | credentials | 3 of 3 | never |
| `--azure-client-id` | credentials | 1 of 1 | always |
| `--azure-client-secret` | credentials | 1 of 1 | always |
| `--azure-org-url` | location | 2 of 2 | always |
| `--azure-resource-group-name` | identity | 1 of 1 | always |
| `--azure-subscription-id` | identity | 1 of 1 | always |
| `--azure-tenant-id` | identity | 1 of 1 | always |
| `--azure-token` | credentials | 2 of 2 | always |
| `--bitbucket-access-token` | credentials | 2 of 2 | always |
| `--bitbucket-password` | credentials | 2 of 2 | always |
| `--bitbucket-username` | credentials | 2 of 2 | always |
| `--bitbucket-workspace` | location | 2 of 2 | always |
| `--bucket` | location | 1 of 1 | always |
| `--build-url` | location | 1 of 1 | always, but in CI only the server does |
| `--clusters` | filter | 1 of 1 | never |
| `--clusters-regex` | filter | 1 of 1 | never |
| `--comment` | metadata | 1 of 1 | never |
| `--commit` | identity | 18 of 18 | some commands |
| `--commit-url` | location | 1 of 1 | always, but in CI only the server does |
| `--compliant` | switch | 2 of 2 | always |
| `--config-file` | location | 2 of 2 | some commands |
| `--control` | identity | 1 of 1 | always |
| `--debug` | global | 1 of 1 | always |
| `--description` | metadata | 22 of 22 | some commands |
| `--digests-source` | identity | 1 of 1 | always |
| `--display-name` | metadata | 1 of 1 | never |
| `--dry-run` | switch | 54 of 54 | always |
| `--end` | identity | 1 of 1 | never |
| `--end-ts` | identity | 1 of 1 | always |
| `--environment` | identity | 4 of 4 | some commands |
| `--exclude` | filter | 21 of 21 | never |
| `--exclude-namespaces` | filter | 1 of 1 | never |
| `--exclude-namespaces-regex` | filter | 1 of 1 | never |
| `--exclude-regex` | filter | 4 of 4 | never |
| `--exclude-services` | filter | 1 of 1 | never |
| `--exclude-services-regex` | filter | 1 of 1 | never |
| `--expires-at` | identity | 2 of 2 | never |
| `--external-fingerprint` | identity | 14 of 14 | always |
| `--external-url` | location | 14 of 14 | always |
| `--fingerprint` | identity | 17 of 17 | some commands |
| `--flow` | identity | 21 of 21 | some commands |
| `--flow-tag` | filter | 1 of 1 | never |
| `--function-names` | filter | 1 of 1 | never |
| `--function-names-regex` | filter | 1 of 1 | never |
| `--github-base-url` | location | 2 of 2 | always |
| `--github-org` | identity | 2 of 2 | always, and not at all in CI |
| `--github-token` | credentials | 2 of 2 | always |
| `--gitlab-base-url` | location | 2 of 2 | always |
| `--gitlab-org` | identity | 2 of 2 | always |
| `--gitlab-token` | credentials | 2 of 2 | always |
| `--grace-period-hours` | identity | 1 of 1 | always |
| `--host` | global | 1 of 1 | always |
| `--http-proxy` | global | 1 of 1 | never |
| `--ignore-branch-match` | switch | 1 of 1 | always |
| `--ignore-case` | switch | 1 of 1 | always |
| `--include` | filter | 2 of 2 | never |
| `--include-regex` | filter | 2 of 2 | never |
| `--included-environments` | filter | 1 of 1 | never |
| `--input-file` | location | 1 of 1 | always |
| `--interval` | output | 2 of 2 | never |
| `--jira-api-token` | credentials | 1 of 1 | always |
| `--jira-base-url` | location | 1 of 1 | always |
| `--jira-issue-fields` | identity | 1 of 1 | never |
| `--jira-pat` | credentials | 1 of 1 | never |
| `--jira-project-key` | identity | 1 of 1 | never |
| `--jira-secondary-source` | identity | 1 of 1 | never |
| `--jira-username` | credentials | 1 of 1 | always |
| `--jq` | location | 1 of 1 | always, some only by the server |
| `--kubeconfig` | credentials | 1 of 1 | always |
| `--link` | metadata | 1 of 1 | always |
| `--logical` | identity | 1 of 1 | always |
| `--max-api-retries` | global | 1 of 1 | always |
| `--max-wait` | output | 1 of 1 | always |
| `--name` | identity | 19 of 19 | some commands |
| `--namespaces` | filter | 1 of 1 | never |
| `--namespaces-regex` | filter | 1 of 1 | never |
| `--new-compliance-status` | switch | 1 of 1 | always |
| `--no-assert` | switch | 3 of 3 | always |
| `--org` | global | 1 of 1 | always |
| `--origin-url` | location | 13 of 13 | never |
| `--original-attestation-type` | identity | 1 of 1 | always |
| `--output` | output | 33 of 33 | some commands |
| `--page` | output | 8 of 8 | always |
| `--page-limit` | output | 8 of 8 | always |
| `--params` | location | 3 of 3 | never |
| `--path` | location | 1 of 1 | always |
| `--paths-file` | location | 1 of 1 | always |
| `--physical` | identity | 1 of 1 | always |
| `--policy` | identity | 4 of 4 | some commands |
| `--privilege` | identity | 2 of 2 | always, some only by the server |
| `--project` | identity | 3 of 3 | always |
| `--provider` | identity | 3 of 3 | some commands |
| `--pull-request` | identity | 1 of 1 | never |
| `--quiet` | global | 1 of 1 | always |
| `--reason` | metadata | 2 of 2 | always |
| `--redact-commit-info` | identity | 14 of 14 | some commands |
| `--region` | location | 1 of 1 | always |
| `--registry-password` | credentials | 16 of 16 | some commands |
| `--registry-username` | credentials | 16 of 16 | some commands |
| `--repo` | identity | 2 of 2 | never |
| `--repo-id` | identity | 17 of 17 | some commands |
| `--repo-provider` | identity | 14 of 14 | some commands |
| `--repo-root` | location | 14 of 14 | never |
| `--repo-url` | location | 14 of 14 | some commands |
| `--repository` | identity | 18 of 18 | some commands, and not at all in CI |
| `--resolve-names` | switch | 1 of 1 | always |
| `--results-dir` | location | 1 of 1 | always |
| `--reverse` | output | 2 of 2 | always |
| `--scan-results` | location | 1 of 1 | always |
| `--schema` | output | 1 of 1 | never |
| `--search` | filter | 2 of 2 | never |
| `--service-account` | identity | 5 of 5 | always |
| `--services` | filter | 1 of 1 | never |
| `--services-regex` | filter | 1 of 1 | never |
| `--set` | metadata | 2 of 2 | always |
| `--short` | output | 1 of 1 | always |
| `--show-input` | output | 3 of 3 | always |
| `--show-unchanged` | output | 1 of 1 | always |
| `--sonar-api-token` | credentials | 1 of 1 | always |
| `--sonar-ce-task-url` | location | 1 of 1 | always |
| `--sonar-project-key` | identity | 1 of 1 | never |
| `--sonar-revision` | identity | 1 of 1 | never |
| `--sonar-server-url` | location | 1 of 1 | never |
| `--sonar-working-dir` | location | 1 of 1 | always |
| `--sort` | output | 1 of 1 | never |
| `--sort-direction` | output | 3 of 3 | never |
| `--space-id` | filter | 1 of 1 | never |
| `--start` | identity | 1 of 1 | never |
| `--start-ts` | identity | 1 of 1 | always |
| `--tag` | metadata | 3 of 3 | never |
| `--template` | identity | 1 of 1 | always |
| `--template-file` | location | 2 of 2 | never |
| `--trail` | identity | 15 of 15 | some commands |
| `--type` | identity | 4 of 4 | some commands |
| `--unset` | metadata | 2 of 2 | never |
| `--upload-results` | switch | 2 of 2 | always |
| `--use-empty-template` | switch | 1 of 1 | always |
| `--user-data` | identity | 13 of 13 | never |
| `--watch` | output | 2 of 2 | always |
| `--zip` | output | 1 of 1 | always |
