# Decision: may a flag ever be given an empty value?

I want to make "this flag was given an empty string" an error across the CLI, so
that an unset shell variable cannot quietly change what a command does.

## TL;DR

- An unset shell variable can silently change what a Kosli command does.
- An audit easily found nineteen cases so far:
  - seventeen of them exiting 0 and printing what success prints.
  - eleven of them change a compliance answer.
  - one 500s an org's environment listing. 
  - one 500s an attest override when reported without a commit.
  - one sends commit author and message to Kosli after being told to redact them.
  - nineteen is a floor, not a total. Every round of looking so far has found more.
  - the space is too large to check case by case.
- The findings are three kinds of problem:
  - the CLI does not notice an empty value it was given.
  - the CLI passes an empty value on where it should send nothing.
  - the server accepts an empty value it should refuse.
- Proposal: an empty value is an error on every flag that is given one. A flag
  whose default is empty is untouched.
- Cost: two `--description` flags can no longer clear a description, and adding
  `--clear-description` gives that back.
- With `--clear-description` alongside, this is a bug fix and can ship in a 2.x
  release. Without it, two real workflows lose a capability and it needs a major
  version and a warning period.

## Why we cannot establish what every empty value does

It is impractical to work out what every command does with every empty flag
value. The CLI has 94 commands declaring 165 distinct flag names between them,
which is already 700 command-and-flag combinations, and that count is the floor
rather than the ceiling:

- **The answer is not in the flag's type.** A flag may be refused by a
  required-flag check, by a value allowlist in `PreRunE`, by a mutual-exclusion
  rule, by the server, or by a third party - or not refused at all. `--page` is
  refused everywhere and `--fingerprint` is accepted on some commands, and no one
  decided either; it fell out of how each was declared.
- **The same flag name behaves differently on different commands.** 23 of the
  165 are refused on some commands and accepted on others.
- **Some differences never reach the output.** `kosli attest generic` prints
  `generic attestation 'unit-test' is reported to trail: my-trail` and exits 0
  whether the attestation went to the artifact or to the trail. The difference is
  only visible in what was stored.
- **The rules change inside CI.** A flag is checked because it is marked
  required, and that marking is skipped when the flag's default comes from a CI
  environment variable. Four combinations we know of stop being checked by the
  CLI inside GitHub Actions. Appendix 1 shows this happening.
- **Flags interact.** `kosli attest generic --fingerprint ""` silently attests to
  the trail; add `--artifact-type file` and the same empty value is refused with
  `only one of --fingerprint, --artifact-type is allowed`. One flag decides
  whether another is checked, so the real space is much larger than 700.
- **Some commands need what we do not have.** 288 of the 700 combinations are on
  20 commands needing AWS, Azure, Google Cloud, Kubernetes, a connected git
  provider, Jira, SonarQube or Snyk. What the CLI does
  with them can still be seen; what the service does cannot.

## What was measured instead

An audit of a slice of that space, taken by running the CLI against a local
server (rather than reasoning about it). The audit code is in the `empty-flag-audit/` 
directory of this repository, its README says how to run it. 

The audit slice is thorough. Each command has a baseline: a working
invocation the audit found. To measure one flag, that flag is taken out of the
baseline and the command is run three times:

- with a real value
- with the flag left out
- with an empty value

It then compares the three sets of:
- exit code
- output
- state left on the server afterwards


It ran all 700 command+flag combinations:
- For 412 of them the command can succeed against the local server alone, so the
whole story is visible: what the CLI did with the empty value, and what the
server did with whatever the CLI sent on.
- For 288 the command needs a service the audit has no credentials for, so only
the CLI's half is visible: its own flag checks run before it contacts the
service.
- The whole set runs twice, once as on a laptop and once with the
environment variables GitHub Actions sets.


### The 412 that can succeed against the local server

| What happens to an empty value | On a laptop | Inside GitHub Actions |
|---|---|---|
| the CLI refuses it | 223 | 219 |
| the CLI accepts it and the server refuses it | 5 | 9 |
| nothing refuses it | 183 | 183 |
| nothing can be said | 1 | 1 |
| **total** | **412** | **412** |

The one that says nothing is `get attestation --fingerprint`, where all three
runs fail the same way, so nothing about the failure is the empty value's doing.

Of the 183 that nothing refuses, 166 do exactly what omitting the flag does.
That is not the same as doing nothing: omitting a flag has a meaning of its own,
and it is rarely the meaning someone had in mind when they wrote that flag with
a variable after it. Appendix 2 groups the 165 flag names by what they are for
and says how many of each kind the CLI already refuses. Appendix 3 is the same
thing flag by flag.

The other 17 (183-166) do something omitting the flag does not. 13 differ only by the deprecation notice a flag prints for being passed at all, leaving 4:

- `attach-policy --environment` attaches the policy to no environment (in Findings table below)
- `detach-policy --environment` detaches it from none (in Findings below)
- `list environments --tag` answers "No environments were found", which is also
  what a real no-match says (in Findings below).
- `update service-account --description` clears the description, the one case where the empty value does what someone would have asked for.

### The 288 that need credentials the audit does not have

| What the CLI does with an empty value | On a laptop | Inside GitHub Actions |
|---|---|---|
| refuses it | 164 | 161 |
| does not react, so it reaches the service | 105 | 108 |
| lets it through, exit 0 | 19 | 19 |
| **total** | **288** | **288** |

The middle row is the one to look at. Those 105 are empty values that survive every
check the CLI has and are handed to someone else. Among them are the credentials:
`--aws-key-id` and `--aws-secret-key` are refused on none of their three commands,
`--jira-pat` on neither of its, and `--registry-password` and
`--registry-username` on 6 of their 17 - so the same registry credential is
checked on six commands and not on the other eleven.

Most credential flags are refused, though. `--github-token`, `--gitlab-token`,
the `--bitbucket-*` and `--azure-*` pair, `--jira-api-token`, `--sonar-api-token`
and `--kubeconfig` are all refused everywhere they appear. Which are and which
are not follows no pattern anyone chose.

What those services then do with the empty values they are handed is the only
thing here that stays unknown, and no amount of work on this machine will show it
- the release plan below is what turns customers' runs into that measurement.

## Findings

One slice produced the nineteen findings below. Each was re-run on its own, 
outside the audit, and its result read from the
server - not taken from the audit's classification, and not inferred by reading
the CLI's source. `$VAR` means a variable that is unset, which is all it takes.

| Command | With the variable unset | |
|---|---|---|
| `kosli create flow F --template-file "$VAR"` | the flow is created with the default template, so it requires none of the attestations intended | changes compliance |
| `kosli begin trail T --flow F --template-file "$VAR"` | the trail requires nothing, so it is compliant whatever is attested to it | changes compliance |
| `kosli attest generic --fingerprint "$VAR"` (and `custom`, `decision`, `junit`) | the attestation is recorded against the trail instead of the artifact. The artifact you meant to attest has no such attestation | changes compliance |
| `kosli attest override --fingerprint "$VAR"` | the override is recorded against the trail. The artifact keeps the compliance status you meant to override | changes compliance |
| `kosli assert artifact --fingerprint "$F" --flow "$VAR"` | the verdict is about whichever flow the fingerprint is found in, not the flow you asked about | changes compliance |
| `kosli fingerprint D --artifact-type dir --exclude "$VAR"` | nothing is excluded, so the fingerprint is a different one. Everything attested, allowlisted or asserted with it then refers to an artifact that was never built. Ten combinations, including `snapshot path` and the attest commands | changes compliance |
| `kosli attest generic --commit "$VAR"` | no git provenance is recorded against the attestation - `git_commit_info` is null where a commit would have been | changes compliance |
| `kosli evaluate input --params "$VAR"` | the policy falls back to its own defaults instead of the parameters you passed. A score of 5 is ALLOWED at `threshold:3` and denied at the default of 10 | changes compliance |
| `kosli evaluate trails T --attestations "$VAR"` | every attestation is evaluated instead of the ones named, so a policy written against a particular set sees a different input | changes compliance |
| `kosli attach-policy P --environment "$VAR"` | the policy is attached to no environment. Anything deployed there is judged without it | changes compliance |
| `kosli detach-policy P --environment "$VAR"` | the policy is detached from no environment, so it stays in force | changes compliance |
| `kosli attest generic --repo-id "$VAR" --repo-url "$VAR" --repository "$VAR"` | `repo_info` is not recorded at all, so the attestation carries no repository provenance | records less than it was told to |
| `kosli attest generic --redact-commit-info "$VAR"` | the commit author and message are sent in the clear - the very data the flag exists to withhold | sends data meant to be withheld |
| `kosli create environment E --type logical --included-environments "$VAR"` | the record cannot be read back, and `list environments` returns HTTP 500 for every environment in the org until it is removed. Omitting the flag entirely does the same, so this one is not about empty values at all | **outright bug**, written up in `../../docs/handover/2026-08-13-logical-environment-500s-the-org-listing.md` |
| `kosli attest override --commit HEAD`, overriding an attestation that was reported without a commit | the server 500s, the CLI retries it three times and gives up, and the override does not happen. No empty value is involved | **outright bug**, written up in `../../docs/handover/2026-08-15-override-500-when-attestation-has-no-commit.md` |
| `kosli begin trail T --flow F` with no `--description` at all | the description the trail already had is replaced by an empty one, even though the server's update is written to leave an absent field alone | **an inconsistency**, not an empty value at all, written up in `../../docs/handover/2026-08-13-upsert-overwrites-unmentioned-fields.md` |
| `kosli list environments --tag "$VAR"` | answers "No environments were found", identical to a real no-match, exit 0 | wrong answer |
| `kosli tag flow F --unset "$VAR"` | the tag is not removed and stays in force. The CSV split of an empty value yields no elements, so `remove_tags` is sent empty, and the command answers "No tags were applied", exit 0 | removes nothing |
| `kosli get attestation NAME --flow F --trail "$VAR"` | `Error: Get "": GET  giving up after 1 attempt(s): Get "": unsupported protocol scheme ""`. Omitting the flag says `at least one of --trail, --fingerprint is required when using ATTESTATION-NAME`, so the empty value defeats the check that produces the good message and the user is shown the plumbing instead | unusable error |

Nineteen findings, from one slice of one CLI, and seventeen of them silent.
And nineteen is where the looking stopped, not where the findings did: 155 of the
combinations that accept an empty value are on commands that record or judge
compliance, and most have never been examined one at a time. Four rounds of
examining them have each produced more - two, then three, then one, then one -
and those rounds also turned up the redaction case, which is not about
compliance data at all, and the override 500, which needs no empty value and was
found by timing the audit rather than by reading its results. Nobody was looking
for either.

That is the argument. We cannot keep finding these one at a time, and a rule that
refuses an empty value everywhere removes the whole class rather than the
instances of it we happen to reach.

Three of them are worth filing on their own account, whatever is decided here,
and each has its own write-up alongside this one.

The last of those needs a word, because it is not what it first looked like.
`kosli create flow` is a create-or-update, and it reads its flags the way
`kubectl apply` reads a manifest: what you pass is what the flow becomes, so an
absent `--description` meaning "no description" is that model working, not a bug.
`begin trail` describes itself the same way, but its server side is written to
leave an absent description alone. Two commands of the same shape, two answers.
The bug is the disagreement, and it is worth settling on its own account: until
the CLI and server agree on what an absent flag means, "empty" and "absent"
cannot be given consistent meanings either.

## Three kinds of problem

The findings are three kinds of problem. They need different fixes, and only the
first is the rule proposed in this document.

**The CLI does not notice an empty value it was given.** `--fingerprint ""`
reaches the trail-scoped endpoint, `--template-file ""` leaves the default
template. This is the rule's own territory, and it is most of the nineteen.

**The CLI sends the empty value on, where it should have sent nothing.** `kosli
list environments --tag ""` puts `?tag=` in the request; omitting the flag sends
no parameter at all. So this is not a case of an empty value meaning what an
absent one means - the CLI makes a different request, and asks the server to
filter on nothing. An optional flag with nothing in it should be left out.

**The server accepts an empty value it should refuse.** Measured separately, in
`2026-08-15-auditing-empty-values-at-the-api.md`: 25 emptied fields and
parameters the API takes, including a flow template requiring an attestation
named "" and an artifact stored with no name. No CLI rule reaches these, because
a customer can send them without the CLI.

One example moves between the boxes on inspection. `kosli attach-policy P
--environment ""` looks like the server accepting an empty list, and it is not:

```
$ kosli attach-policy my-policy --environment ""
policy 'my-policy' is attached to environments: []
```

No request is made at all. The CLI iterates an empty list, calls nothing, and
reports success. The API cannot refuse what it is never sent, so this one is the
CLI's alone.

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
the flag - which is 166 of the 183 - the server receives an identical payload
either way, and there is nothing empty in it to reject.

Those 166 fall into three kinds, by whose problem the payload turns out to be.
Here is one example of each, read with `--debug`:

| Command | What reaches the server | Whose problem |
|---|---|---|
| `kosli attest generic --fingerprint ""` | no fingerprint at all, sent to the trail-scoped endpoint `/attestations/{org}/{flow}/trail/{trail}/generic` | the CLI's. It picked that endpoint. Someone calling the API chooses the artifact endpoint or the trail endpoint deliberately, and is not misled. The rule settles this one |
| `kosli create environment --included-environments ""` | a logical environment with **no** `included_environments` field, which the server accepts with a 201 and then cannot read back. Omitting the flag sends the same thing | neither's, for this decision. It is a plain defect, being fixed on its own account |
| `kosli begin trail` | `"description": ""`, whether or not `--description` was passed | shared. The server's update is written to leave an absent description alone, and the CLI defeats that by always sending one. A direct caller that sends `""` gets the same result, so the CLI fix alone does not settle it |

So the server's half of this is not the CLI rule repeated. The CLI rule is
"refuse an empty value". The server's is: **an absent field means what the
command's verb says it means, and the CLI and the server must agree about
that.** Deciding it command by command would be ninety-four decisions nobody
will remember, which is how the CLI got inconsistent in the first place.
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

Empty is always an error, on every flag that is given one, with no exemptions to
remember. A flag whose default is empty is untouched: the rule is about what a
command was told, not about what it falls back on. That is also how it is
checked - pflag records whether a flag was set, and the check reads that.

It works because it separates two intents that are spelled identically today: "I
meant to pass a value and my variable was empty" and "I want no value". They look
the same, and no amount of checking can tell them apart after the fact.

Most breaks will be where a variable is genuinely unset, and in those pipelines
the command is already doing something its author did not ask for. There the
failure replaces silent wrong behaviour rather than correct behaviour.

Not every break is that, though. Someone may be passing an empty value
deliberately, with no variable involved. They lose nothing they cannot still get
by omitting the flag, but their pipeline fails until somebody edits it.

There is a good argument that this is not a breaking change at all but a bug
fix, and that it should ship as one: what starts failing is a command that was
already doing something its author did not ask for, and a release note nobody
can act on helps nobody. The measurements say that argument holds for all but
two of the 202 combinations that change. The two are below, and they decide the
question rather than qualify it.

## Releasing this

Commands that succeed today will start failing, and the first question is
whether that is a breaking change at all.

The case for shipping it as a bug fix, in a 2.x release, is strong. A command
that fails under this rule was passing an empty value it was never meant to
pass, and doing something its author did not ask for. Nobody is deprived of
behaviour they wanted. Held against that, a staged rollout keeps the compliance
holes open for the length of the stage.

The measurements support that for all but two of the 202 combinations. The
exceptions are `kosli update control --description ""` and `kosli update
service-account --description ""`, which clear a description on purpose, because
`update` is a patch and omitting the flag cannot clear anything. Those two are
not bad commands being caught. They are a capability being removed, and removing
it without a replacement is a regression however it is labelled.

So the question is not "breaking or not" but "what ships alongside":

- **With `--clear-description`**, nothing is taken away. Every remaining failure
  is a command that was already wrong, the change is a bug fix, and it can go
  out in a 2.x release without a staged rollout.
- **Without it**, two real workflows lose the only way they have to say what
  they mean, and that does need a major version and a warning period.

The first is better, and it is cheap: one flag on two commands. The rest of this
section is what the second would cost, and is worth keeping only if
`--clear-description` turns out to be harder than it looks.

Two things make the second sharper than usual:

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

### If it is staged: four steps

This is the shape the change takes if `--clear-description` does not ship with
it, or if we decide the impact needs measuring before the rule lands. At least
202 combinations change - 183 on the commands that work here, and 19 more on the
ones needing a service, where the CLI lets an empty value through before the
service is even reached - and staging is how that is made gradual.

Steps 1 and 2 are worth reading even if the rule ships as a bug fix, because the
warnings they describe are the only way to learn what an empty value does at the
services this audit cannot reach.

#### Step 1: somewhere to put a warning, in app.kosli.com

Nobody reads warnings in a CI workflow run. A step that only prints one is not a
migration, it is a delay, and we would reach step 3 knowing no more than we do
now. So before the CLI warns about anything, there has to be somewhere for the
warning to go.

A warning goes to two places, and no more than two:

1. **The workflow run**, printed as now.
2. **app.kosli.com, at the org level.** A command that has `--org` and
   `--api-token` can send the warning whatever else it was doing, so this covers
   178 of the 183. The exception is `kosli fingerprint`, which is entirely local
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

It also reaches where this audit could not. For the 288 combinations on commands
needing AWS, Azure, a git provider and the rest, we can see what the CLI does but
not what the service does with the 105 empty values the CLI hands over. Customers
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

That moves four combinations - `--build-url` and `--commit-url`, on `attest
artifact` and on `report artifact` - from being refused by the CLI to being
refused by the server. All four are still refused here, because this server
happens to check them. The point is that the CLI stopped, in exactly the place
these bugs happen. That is cli#1088.

---

## Appendix 2: the flags grouped by what they are for

The question worth asking of each flag is what it means to the person running the
command, and whether an empty value means anything for that. Grouped that way,
the 165 names are:

| What the flag is for, with a few examples | Names | Does an empty value mean anything? | CLI always refuses | Only the server refuses | Refuses on some commands | Never refuses |
|---|---|---|---|---|---|---|
| identity and selection - `--flow`, `--trail`, `--fingerprint`, `--name` | 53 | no. There is no artifact called "" | 25 | 1 | 12 | 15 |
| location and input - `--template-file`, `--results-dir`, `--paths` | 27 | no. There is no file called "" | 18 | 1 | 2 | 6 |
| filters - `--exclude`, `--namespaces`, `--services`, `--attestations` | 24 | no. Filtering on "" filters on nothing | 3 | 0 | 0 | 21 |
| credentials - `--github-token`, `--aws-secret-key`, `--registry-password` | 17 | no. There is no token "" | 12 | 0 | 2 | 3 |
| output and paging - `--output`, `--sort`, `--page`, `--reverse` | 14 | no | 9 | 0 | 1 | 4 |
| behaviour switches - `--dry-run`, `--assert`, `--compliant` | 13 | no | 13 | 0 | 0 | 0 |
| free-text metadata - `--description`, `--comment`, `--reason`, `--tag` | 10 | **sometimes** | 4 | 0 | 1 | 5 |
| the global flags - `--org`, `--api-token`, `--host`, `--debug` | 7 | no | 6 | 0 | 0 | 1 |
| **total** | **165** | | **90** | **2** | **18** | **55** |

The credentials row is the one to look at twice. Twelve of its seventeen names
are refused by the CLI everywhere they appear. Three never are - `--aws-key-id`,
`--aws-secret-key` and `--jira-pat` - so an empty one is handed to AWS or to Jira,
and whether it is caught then depends on that service and not on us. The last two,
`--registry-password` and `--registry-username`, are refused on six commands and
not on the other eleven. Kosli's own
`--api-token` is not among them: it is a global flag, and it already refuses an
empty value with `--api-token is not set`.

---

## Appendix 3: every flag in the CLI

All 165 flag names, what each is for, how many of its commands were run, and what
happened. Commands needing AWS, Azure, Google Cloud, Kubernetes, a connected git
provider, Jira, SonarQube or Snyk are included: they cannot succeed without
credentials the audit does not have, but their own checks run first, so whether
the CLI refuses an empty value is still visible.

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
| `--artifact-type` | identity | 17 of 17 | some commands |
| `--assert` | switch | 9 of 9 | always |
| `--assume-yes` | switch | 2 of 2 | always |
| `--attachments` | identity | 12 of 12 | always |
| `--attestation-data` | identity | 1 of 1 | always |
| `--attestation-id` | identity | 1 of 1 | always |
| `--attestations` | filter | 3 of 3 | never |
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
| `--build-url` | location | 2 of 2 | always, but in CI only the server does |
| `--cluster` | identity | 1 of 1 | always |
| `--clusters` | filter | 1 of 1 | never |
| `--clusters-regex` | filter | 1 of 1 | never |
| `--comment` | metadata | 1 of 1 | never |
| `--commit` | identity | 18 of 18 | some commands |
| `--commit-url` | location | 2 of 2 | always, but in CI only the server does |
| `--compliant` | switch | 2 of 2 | always |
| `--config-file` | location | 2 of 2 | some commands |
| `--control` | identity | 1 of 1 | always |
| `--debug` | global | 1 of 1 | always |
| `--description` | metadata | 22 of 22 | some commands |
| `--digests-source` | identity | 1 of 1 | always |
| `--display-name` | metadata | 1 of 1 | never |
| `--dry-run` | switch | 56 of 56 | always |
| `--e` | identity | 2 of 2 | never |
| `--end` | identity | 1 of 1 | never |
| `--end-ts` | identity | 1 of 1 | always |
| `--environment` | identity | 4 of 4 | some commands |
| `--exclude` | filter | 23 of 23 | never |
| `--exclude-namespaces` | filter | 1 of 1 | never |
| `--exclude-namespaces-regex` | filter | 1 of 1 | never |
| `--exclude-regex` | filter | 4 of 4 | never |
| `--exclude-scaling` | filter | 1 of 1 | always |
| `--exclude-services` | filter | 1 of 1 | never |
| `--exclude-services-regex` | filter | 1 of 1 | never |
| `--expires-at` | identity | 2 of 2 | never |
| `--external-fingerprint` | identity | 14 of 14 | always |
| `--external-url` | location | 14 of 14 | always |
| `--fingerprint` | identity | 18 of 18 | some commands |
| `--flow` | identity | 23 of 23 | some commands |
| `--flow-tag` | filter | 1 of 1 | never |
| `--function-name` | identity | 1 of 1 | always |
| `--function-names` | filter | 1 of 1 | never |
| `--function-names-regex` | filter | 1 of 1 | never |
| `--function-version` | identity | 1 of 1 | always |
| `--git-commit` | identity | 1 of 1 | always |
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
| `--include-scaling` | filter | 1 of 1 | always |
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
| `--name` | identity | 20 of 20 | some commands |
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
| `--paths` | location | 1 of 1 | always |
| `--paths-file` | location | 1 of 1 | always |
| `--physical` | identity | 1 of 1 | always |
| `--policy` | identity | 4 of 4 | some commands |
| `--privilege` | identity | 2 of 2 | always, some only by the server |
| `--project` | identity | 3 of 3 | always |
| `--provider` | identity | 3 of 3 | never |
| `--pull-request` | identity | 1 of 1 | never |
| `--quiet` | global | 1 of 1 | always |
| `--reason` | metadata | 2 of 2 | always |
| `--redact-commit-info` | identity | 14 of 14 | never |
| `--region` | location | 1 of 1 | always |
| `--registry-password` | credentials | 17 of 17 | some commands |
| `--registry-provider` | identity | 17 of 17 | some commands |
| `--registry-username` | credentials | 17 of 17 | some commands |
| `--repo` | identity | 2 of 2 | never |
| `--repo-id` | identity | 17 of 17 | some commands |
| `--repo-provider` | identity | 14 of 14 | never |
| `--repo-root` | location | 15 of 15 | never |
| `--repo-url` | location | 14 of 14 | never |
| `--repository` | identity | 18 of 18 | some commands, and not at all in CI |
| `--require-provenance` | switch | 1 of 1 | always |
| `--resolve-names` | switch | 1 of 1 | always |
| `--results-dir` | location | 1 of 1 | always |
| `--reverse` | output | 2 of 2 | always |
| `--scan-results` | location | 1 of 1 | always |
| `--schema` | output | 1 of 1 | never |
| `--search` | filter | 2 of 2 | never |
| `--service-account` | identity | 5 of 5 | always |
| `--service-name` | identity | 1 of 1 | always |
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
| `--visibility` | metadata | 1 of 1 | never |
| `--watch` | output | 2 of 2 | always |
| `--yes` | switch | 2 of 2 | always |
| `--zip` | output | 1 of 1 | always |
