# Decision: may a flag ever be given an empty value?

## TL;DR

An unset shell variable can silently change what a Kosli CLI command does.  
An audit easily found nineteen cases so far:
  - seventeen of them exit 0 and print what success prints.
  - eleven of them change a compliance answer.
  - one 500s an org's environment listing. 
  - one 500s an attest override when reported without a commit.
  - one sends commit author and message to Kosli after being told to redact them.
  - nineteen is a floor, not a total. 
  - every round of looking so far has found more.

## There are three kinds of problem

1. the server accepts an empty value it should refuse.  
  **Proposal** - _not_ covered in this document:
    - refuse an empty `filename` and an empty `template` attestation name
    - `remove_tags` turned out not to need it, see below
    - decide empty query params
    - leave the 6 description fields alone
    - measured in
  `hack/empty-flag-audit/docs/2026-08-15-auditing-empty-values-at-the-api.md`
    - what was done, and the trap in the original wording, in "Refusing an empty
  value on the server" below

2. the CLI passes an empty value to the server where it should send nothing.  
  **Proposal** - _not_ covered in this document:
    - make an apply-vs-patch decision
    - `omitempty` on 173 fields (34 have it)
    - server and CLI must move together
    - written up in
  `hack/empty-flag-audit/docs/2026-08-13-upsert-overwrites-unmentioned-fields.md`.  
  
3. the CLI does not notice an empty flag value it was given.  
  **Proposal** - detailed in _this_ document:
    - an empty value is an error on every flag that is given one
    - a flag whose default is empty is untouched
    - clearing a description stops being possible, which is accepted
    - ship as a bug fix in a 2.x release

## Why we cannot establish what every empty value does

It is impractical to work out what every command does with every empty flag
value. The CLI has 94 commands declaring 165 distinct flag names between them,
which is already 700 command-and-flag combinations, and that count is the floor
rather than the ceiling:

- The answer is not in the flag's type.
- The same flag name can behave differently on different commands.
- Some differences never reach the output.
- The rules can change inside CI.
- Flags can interact.
- Some commands talk to external services.

## What was measured instead

An audit of a slice of that space, taken by running the CLI against a local
server (rather than reasoning about it). The audit code is in the `hack/empty-flag-audit/` 
directory of this repository, its README says how to run it. 

The audit slice is thorough. Each command has a baseline: a working
invocation the audit found. To measure one flag, that flag is taken out of the
baseline and the command is run three times:

- with a real value for the flag
- with the flag left out
- with an empty value for the flag

The audit then compares the three sets of:
- exit code
- output
- state left on the server afterwards


The audit ran all 700 command+flag combinations twice,
once as on a laptop and once with the environment variables GitHub Actions sets.
(There are small differences, as detailed in Appendix 1.)

- For 412 the command can succeed against the local server alone, so the
whole story is visible: what the CLI did with the empty value, and what the
server did with whatever the CLI sent on.
- For 288 the command needs a service the audit has no credentials for, so only
the CLI's half is visible: its own flag checks run before it contacts the
service.


## The 412 that can succeed against the local server

| What happens to an empty value | Count | 
|---|---|
| the CLI refuses it | 224 | 
| the CLI accepts it but the server refuses it | 5 | 
| nothing refuses it | 183 | 
| **total** | **412** | 

Of the 183 that nothing refuses, 165 do exactly what omitting the flag does.
That is not the same as doing nothing: omitting a flag has a meaning of its own.
But it is rarely, if ever, the meaning _intended_ when that flag was written with
a variable after it:

| What the flag names | How many | What omitting it does (== what `--flag ""` does) |
|---|---|---|
| what the command acts on | 72 | a different target, or none. `--fingerprint` sends the attestation to the trail rather than the artifact; `--commit`, `--repo-id` and `--repository` leave the provenance unrecorded |
| where to read from | 27 | the file or URL is never read, so what it held is simply absent: `--template-file`, `--params`, `--config-file`, `--repo-root` |
| which of many | 20 | no filter, so everything is in scope. `--exclude ""` excludes nothing, on 12 commands |
| a credential | 20 | nothing is sent. All 20 are `--registry-username` and `--registry-password` |
| text carried along | 18 | the field is left unset: `--description` on 12 commands, `--tag`, `--display-name`, `--comment` |
| how to present the answer | 7 | the default format or ordering: `--sort`, `--sort-direction`, `--interval`, `--schema` |
| a proxy | 1 | `--http-proxy`, so the request goes direct |
| **total** | **165** | |

Refusing an empty value takes none of this away. In every one of the 165 the
empty value already does what omitting the flag does, so a workflow that _really_
wants that behaviour can still have it by not writing the flag. What it takes away is
the ability to spell it `--flag ""`, which is the spelling an unset variable
produces, and which is why the two cannot be told apart today.

The other 18 (183-165) do something omitting the flag does not. 13 differ only by the deprecation notice a flag prints for being passed at all, leaving 5:

- `attach-policy --environment` attaches the policy to no environment (in Findings below)
- `detach-policy --environment` detaches it from no Environments (in Findings below)
- `list environments --tag` answers "No environments were found", which is also
  what a real no-match says (in Findings below).

The remaining two are `update control --description` and `update service-account --description` which clear the description. These two are the _only_ cases in the 
700 where an empty value does what was asked for, and the only capability this 
proposal removes. These are covered in the `The one capability this removes` section.

## The 288 that need credentials the audit does not have

| What the CLI does with an empty value | Count | 
|---|---|
| refuses it | 164 | 
| does not react, so it reaches the service | 105 | 
| lets it through, exit 0 | 19 | 
| **total** | **288** | 

The middle 105 are empty values that survive every check the CLI has and are handed
to a service. 
What those services then do with the empty values they are handed is unknown. 
There no discernible refusal pattern.
For example, credentials:
- `--aws-key-id` and `--aws-secret-key` are refused on _none_ of their commands
- `--registry-password` and `--registry-username` are refused on _some_ of their commands
- `--jira-pat`, `--github-token`, `--gitlab-token`, the `--bitbucket-*` and `--azure-*` pair, `--jira-api-token`, `--sonar-api-token` and `--kubeconfig` are refused on _all_ their commands. 


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
| `kosli attach-policy P --environment "$VAR"` | the policy is attached to no environment. Anything deployed there is judged without it. The CLI iterates an empty list and makes no request, so nothing on the server can refuse it | changes compliance |
| `kosli detach-policy P --environment "$VAR"` | the policy is detached from no environment, so it stays in force | changes compliance |
| `kosli attest generic --repo-id "$VAR" --repo-url "$VAR" --repository "$VAR"` | `repo_info` is not recorded at all, so the attestation carries no repository provenance | records less than it was told to |
| `kosli attest generic --redact-commit-info "$VAR"` | the commit author and message are sent in the clear - the very data the flag exists to withhold | sends data meant to be withheld |
| `kosli create environment E --type logical --included-environments "$VAR"` | the record cannot be read back, and `list environments` returns HTTP 500 for every environment in the org until it is removed. Omitting the flag entirely does the same, so this one is not about empty values at all | **outright bug**, [server issue 6503](https://github.com/kosli-dev/server/issues/6503) |
| `kosli attest override --commit HEAD`, overriding an attestation that was reported without a commit | the server 500s, the CLI retries it three times and gives up, and the override does not happen. No empty value is involved | **outright bug**, [server issue 6504](https://github.com/kosli-dev/server/issues/6504) |
| `kosli begin trail T --flow F` with no `--description` at all | the description the trail already had is replaced by an empty one, even though the server's update is written to leave an absent field alone | **an inconsistency**, not an empty value at all, written up in `hack/empty-flag-audit/docs/2026-08-13-upsert-overwrites-unmentioned-fields.md` |
| `kosli list environments --tag "$VAR"` | answers "No environments were found", identical to a real no-match, exit 0 | wrong answer |
| `kosli tag flow F --unset "$VAR"` | the tag is not removed and stays in force. The CSV split of an empty value yields no elements, so `remove_tags` is sent empty, and the command answers "No tags were applied", exit 0 | removes nothing |
| `kosli get attestation NAME --flow F --trail "$VAR"`, and the same with `--fingerprint "$VAR"` | `Error: Get "": GET  giving up after 1 attempt(s): Get "": unsupported protocol scheme ""`. Omitting either flag says `at least one of --trail, --fingerprint is required when using ATTESTATION-NAME`, so the empty value defeats the check that produces the good message and the user is shown the plumbing instead. Either flag on its own is enough | unusable error |

Nineteen findings, from one slice of one CLI, seventeen of them silent.
And nineteen is where the looking stopped, not where the findings did. Four rounds of
examining them have each produced more cases, as well as finding the redaction case,
and the attest-override 500 case. 

We cannot keep finding these one at a time. A rule that
refuses an empty value everywhere removes the whole class rather than the
instances we happen to stumble upon.

Three of them are worth filing on their own account, whatever is decided here.

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

## Refusing an empty value on the server (done 2026-08-19)

Problem 1's proposal above originally read "add schema `minLength: 1` on
`filename`, `template`, `remove_tags`". Two of the three landed; the wording of
the third was a trap worth recording, because it would have caused an outage.

**The trap: a model that validates a write may also validate a read.** A Pydantic
constraint is not a rule about incoming requests. It is a rule about every
construction of that model, and in this codebase several models are built both
from a request and from a stored document. Adding a constraint there does not
reject bad input; it rejects **data already in the database**, turning a bad
record into one that cannot be loaded at all. That is strictly worse: a wrong
requirement becomes a failed read, and one record can break a listing for
everyone in the org, which is the shape of
[server issue 6503](https://github.com/kosli-dev/server/issues/6503).

So the question to ask of any empty-value constraint is not "is this field
user-supplied?" but "is this class ever constructed from Mongo?".

| Field | Where the constraint went | Why |
|---|---|---|
| `filename` | `min_length=1` on `CreateArtifact` | Bound only as a FastAPI request body, constructed nowhere else. `ArtifactResponseBase` carries the same field and **is** built from stored artifacts, so it was left permissive on purpose. |
| template attestation name | a check on the write path in `common/flow.py`, not on the model | `Attestation` is built when a stored template is read back, and no shape difference distinguishes a write from reading a legacy record, so there was no model-level place to put it. |
| `remove_tags` | nothing | Measured: an empty entry removes no tag, so it does not reach the record. Refusing it is tidiness, not a fix. |

Both fixes carry a test that loads the bad value from the stored shape and
asserts it still works, and both were checked by restoring the constraint and
watching that test fail. A constraint of this kind that has not been checked that
way has not been checked.

Everything above is about newly submitted values. Records already holding an
empty value stay as they are, and reading them keeps working. Cleaning them up
would be a migration, which nobody has asked for.

Confirmed by measurement rather than by the tests alone: `replay.py` re-run
against a server carrying both fixes answers "the server refuses it" for all five
rows that reach `filename` and `template`, taking the acceptances from 25 to 22.
See "What the re-run changed" in
`2026-08-15-auditing-empty-values-at-the-api.md`.

## The one capability this removes

Two flags, and only two, use an empty value to mean something you cannot say any
other way: clear a description on purpose. Refusing an empty value takes that away.
- `kosli update control --description ""`
- `kosli update service-account --description ""`

They could say it because `update` is a _patch_: omitting the flag changes
nothing there, so an empty value was the only way to say "clear this". No other
command is in that position.

Nothing replaces it. Clearing a description is not behaviour the CLI has to
offer, so these two commands lose it: a description can be changed but not
emptied. That is a decision rather than a finding - the measurement only says
where the capability existed, not that it is worth keeping.

Both endpoints accept `description: null`, so a way to clear one can be added
later without server work if it is ever asked for.

Whether anyone relies on this capability is answerable before the change ships,
for controls at least. Every create and update of a control writes a version 
document carrying that version's description, so a control whose description goes 
from non-empty to empty between two adjacent versions has had one cleared, and 
the version says who did it and when. Service accounts keep no such history - the
update writes the field straight onto the record - so for those the question can
only be answered from a new measurement onwards.

Everywhere else, nothing is lost: an empty value either does what omitting the
flag does, or does something nobody asked for.

## Proposal

Empty is _always_ an error, on every flag that is given one. No exceptions. 
A flag whose default is empty is untouched: the rule is about what a
command was told, not about what it falls back on. That is also how it is
checked - every flag's value is wrapped, and the wrapper refuses what it is
handed rather than reading what the flag ends up holding. A default is never
handed to it, and an empty element of a multi-value flag is still there to be
seen.

It works because it separates two intents that are spelled identically today: "I
meant to pass a value and my variable was empty" and "I want no value". They look
the same, and no amount of checking can tell them apart after the fact.

It has to be the CLI, not the server. For 165 of the 183, an empty value produces
exactly the request that omitting the flag produces, so the server receives an
identical payload either way and has nothing empty to reject. Only the CLI can
still see the difference, because only the CLI saw the flag.

## Releasing this

A command that fails under this rule was passing an empty value it was never
meant to pass, and doing something its author did not ask for. 
- Nobody is deprived of behaviour they wanted.
- Everyone affected has an easy one-line fix to make.

The measurements support that for all but two of the 202 combinations.
The exceptions are:
- `kosli update control --description ""` 
- `kosli update service-account --description ""`

Those two are not bad commands being caught. They are a capability being
removed, and removing it is the decision: clearing a description is not
required behaviour, so nothing ships in its place.

This goes out as a bug fix in a 2.x release, with a release note.

---

## Appendix 0: the staged release, if it is needed

If we decide the impact needs measuring before refusing empty values lands, we
can stage the release.
At least 202 combinations change - 183 plus 19 more on the
ones needing a service. Staging is how that is made gradual.

Steps 1 and 2 are worth reading even if the empty-value refusal ships as a bug 
fix, because the warnings they describe are the only way to learn what an empty
value does at the services this audit cannot reach.

### Step 1: somewhere to put a warning, in app.kosli.com

Nobody reads warnings in a CI workflow run. A step that only prints one is not a
migration, it is a delay, and we would reach step 3 knowing no more than we do
now. So before the CLI warns about anything, there has to be somewhere for the
warning to go.

A warning goes to two places:

1. **The workflow run**, printed as now.
2. **app.kosli.com, at the org level.** A command that has `--org` and
   `--api-token` can send the warning whatever else it was doing, so this covers
   178 of the 183. The exception is `kosli fingerprint`, which is entirely local
   and needs no credentials.

This is work in app.kosli.com: somewhere to receive the warnings, and one place
per org to see them.

### Step 2: the CLI warns, in a 2.x release, breaking nothing

Fix the two bugs above - the absent-flag disagreement and the
`--included-environments` 500 - and make every empty value print a warning naming
the flag, and report it. Nothing starts failing, and anyone whose pipeline has an
unset variable can see it and fix it before it costs them anything.

The description fix has an order to it: the server must accept a payload with no
`description` before the CLI stops sending one, or every `kosli create flow`
fails. Its write-up has the detail.

### Step 3: the warning becomes the error, in a major release of its own

One guard, one migration, one release note, and nothing else breaking in the same
version. Ship `KOSLI_ALLOW_EMPTY_FLAG_VALUES=true` alongside it as an escape
hatch, so anyone caught out has a one-line unblock while they fix the pipeline,
and remove it in the next major release.

When to ship it is a question step 2 answers.

### Step 4: delete what the guard replaced

The bespoke emptiness checks scattered through the commands become redundant once
one rule covers every flag. This is the step that is easiest to skip and the
reason the CLI is inconsistent today, so it belongs in the plan rather than in
someone's memory.

### Why steps 1 and 2 come first

Reporting warnings is not only a kindness to customers. It answers the question a
release note cannot: how much would step 3 actually break? Today that is a guess.
With this it becomes a number, per org, and we can tell the customers
who are affected before it lands rather than after.

It also reaches where this audit could not. For the 288 combinations on commands
needing AWS, Azure, a git provider and the rest, we can see what the CLI does but
not what the service does with the 105 empty values the CLI hands over. Customers
have the credentials we lack, so their warnings are the only place that half can
be found out.

Three conditions on the reporting:
1. Sending a warning must never fail the command that produced it: this is a
report, not a check.
2. It has to stay rare enough to be worth looking at, which the numbers here
suggest it will be. 
3. It must go to the `--host` the command was already given, through the same `--http-proxy` - not to a hardcoded address.

That last one is what makes egress restrictions a non-issue. A command reporting
a snapshot from inside a locked-down network can already reach the host it
reports to, because that is what it is doing; a warning to the same host needs
nothing that the command did not already need. 

---

## Appendix 1: the CLI checks the same inside CI

A flag's default has nothing to do with whether an empty value is refused. The
wrapper refuses what a command is handed, so a default filled in from a CI
environment variable changes nothing:

```
# on a laptop
kosli attest artifact ... --build-url ""
Error: flag '--build-url' was given an empty value

# in GitHub Actions
kosli attest artifact ... --build-url ""
Error: flag '--build-url' was given an empty value
```

Four combinations used to differ here - `--build-url` and `--commit-url`, on
`attest artifact` and on `report artifact` - because `RequireFlags` marks a flag
required only when its default is empty, and inside CI those defaults are
filled. The audit run with CI variables set records all four as refused by the
CLI, along with the other 696. That closes cli#1088.

---

## Appendix 2: the flags grouped by what they are for

The question worth asking of each flag is what it means to the person running the
command, and whether an empty value means anything for that.

This counts flag names across the whole CLI, and asks how many of each kind the
CLI already refuses. It is not the earlier table of 165 combinations, which asks
what an empty value does where nothing refuses it; those use 43 flag names
between them, and the matching totals are coincidence.

Grouped that way, the 165 names are:

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
| `--annotate` | metadata | 13 | always |
| `--api-token` | global | 1 | always |
| `--archived` | filter | 1 | always |
| `--artifact-type` | identity | 17 | some commands |
| `--assert` | switch | 9 | always |
| `--assume-yes` | switch | 2 | always |
| `--attachments` | identity | 12 | always |
| `--attestation-data` | identity | 1 | always |
| `--attestation-id` | identity | 1 | always |
| `--attestations` | filter | 3 | never |
| `--aws-key-id` | credentials | 3 | never |
| `--aws-region` | location | 3 | some commands |
| `--aws-secret-key` | credentials | 3 | never |
| `--azure-client-id` | credentials | 1 | always |
| `--azure-client-secret` | credentials | 1 | always |
| `--azure-org-url` | location | 2 | always |
| `--azure-resource-group-name` | identity | 1 | always |
| `--azure-subscription-id` | identity | 1 | always |
| `--azure-tenant-id` | identity | 1 | always |
| `--azure-token` | credentials | 2 | always |
| `--bitbucket-access-token` | credentials | 2 | always |
| `--bitbucket-password` | credentials | 2 | always |
| `--bitbucket-username` | credentials | 2 | always |
| `--bitbucket-workspace` | location | 2 | always |
| `--bucket` | location | 1 | always |
| `--build-url` | location | 2 | always, but in CI only the server does |
| `--cluster` | identity | 1 | always |
| `--clusters` | filter | 1 | never |
| `--clusters-regex` | filter | 1 | never |
| `--comment` | metadata | 1 | never |
| `--commit` | identity | 18 | some commands |
| `--commit-url` | location | 2 | always, but in CI only the server does |
| `--compliant` | switch | 2 | always |
| `--config-file` | location | 2 | some commands |
| `--control` | identity | 1 | always |
| `--debug` | global | 1 | always |
| `--description` | metadata | 22 | some commands |
| `--digests-source` | identity | 1 | always |
| `--display-name` | metadata | 1 | never |
| `--dry-run` | switch | 56 | always |
| `--e` | identity | 2 | never |
| `--end` | identity | 1 | never |
| `--end-ts` | identity | 1 | always |
| `--environment` | identity | 4 | some commands |
| `--exclude` | filter | 23 | never |
| `--exclude-namespaces` | filter | 1 | never |
| `--exclude-namespaces-regex` | filter | 1 | never |
| `--exclude-regex` | filter | 4 | never |
| `--exclude-scaling` | filter | 1 | always |
| `--exclude-services` | filter | 1 | never |
| `--exclude-services-regex` | filter | 1 | never |
| `--expires-at` | identity | 2 | never |
| `--external-fingerprint` | identity | 14 | always |
| `--external-url` | location | 14 | always |
| `--fingerprint` | identity | 18 | some commands |
| `--flow` | identity | 23 | some commands |
| `--flow-tag` | filter | 1 | never |
| `--function-name` | identity | 1 | always |
| `--function-names` | filter | 1 | never |
| `--function-names-regex` | filter | 1 | never |
| `--function-version` | identity | 1 | always |
| `--git-commit` | identity | 1 | always |
| `--github-base-url` | location | 2 | always |
| `--github-org` | identity | 2 | always, and not at all in CI |
| `--github-token` | credentials | 2 | always |
| `--gitlab-base-url` | location | 2 | always |
| `--gitlab-org` | identity | 2 | always |
| `--gitlab-token` | credentials | 2 | always |
| `--grace-period-hours` | identity | 1 | always |
| `--host` | global | 1 | always |
| `--http-proxy` | global | 1 | never |
| `--ignore-branch-match` | switch | 1 | always |
| `--ignore-case` | switch | 1 | always |
| `--include` | filter | 2 | never |
| `--include-regex` | filter | 2 | never |
| `--include-scaling` | filter | 1 | always |
| `--included-environments` | filter | 1 | never |
| `--input-file` | location | 1 | always |
| `--interval` | output | 2 | never |
| `--jira-api-token` | credentials | 1 | always |
| `--jira-base-url` | location | 1 | always |
| `--jira-issue-fields` | identity | 1 | never |
| `--jira-pat` | credentials | 1 | never |
| `--jira-project-key` | identity | 1 | never |
| `--jira-secondary-source` | identity | 1 | never |
| `--jira-username` | credentials | 1 | always |
| `--jq` | location | 1 | always, some only by the server |
| `--kubeconfig` | credentials | 1 | always |
| `--link` | metadata | 1 | always |
| `--logical` | identity | 1 | always |
| `--max-api-retries` | global | 1 | always |
| `--max-wait` | output | 1 | always |
| `--name` | identity | 20 | some commands |
| `--namespaces` | filter | 1 | never |
| `--namespaces-regex` | filter | 1 | never |
| `--new-compliance-status` | switch | 1 | always |
| `--no-assert` | switch | 3 | always |
| `--org` | global | 1 | always |
| `--origin-url` | location | 13 | never |
| `--original-attestation-type` | identity | 1 | always |
| `--output` | output | 33 | some commands |
| `--page` | output | 8 | always |
| `--page-limit` | output | 8 | always |
| `--params` | location | 3 | never |
| `--path` | location | 1 | always |
| `--paths` | location | 1 | always |
| `--paths-file` | location | 1 | always |
| `--physical` | identity | 1 | always |
| `--policy` | identity | 4 | some commands |
| `--privilege` | identity | 2 | always, some only by the server |
| `--project` | identity | 3 | always |
| `--provider` | identity | 3 | never |
| `--pull-request` | identity | 1 | never |
| `--quiet` | global | 1 | always |
| `--reason` | metadata | 2 | always |
| `--redact-commit-info` | identity | 14 | never |
| `--region` | location | 1 | always |
| `--registry-password` | credentials | 17 | some commands |
| `--registry-provider` | identity | 17 | some commands |
| `--registry-username` | credentials | 17 | some commands |
| `--repo` | identity | 2 | never |
| `--repo-id` | identity | 17 | some commands |
| `--repo-provider` | identity | 14 | never |
| `--repo-root` | location | 15 | never |
| `--repo-url` | location | 14 | never |
| `--repository` | identity | 18 | some commands, and not at all in CI |
| `--require-provenance` | switch | 1 | always |
| `--resolve-names` | switch | 1 | always |
| `--results-dir` | location | 1 | always |
| `--reverse` | output | 2 | always |
| `--scan-results` | location | 1 | always |
| `--schema` | output | 1 | never |
| `--search` | filter | 2 | never |
| `--service-account` | identity | 5 | always |
| `--service-name` | identity | 1 | always |
| `--services` | filter | 1 | never |
| `--services-regex` | filter | 1 | never |
| `--set` | metadata | 2 | always |
| `--short` | output | 1 | always |
| `--show-input` | output | 3 | always |
| `--show-unchanged` | output | 1 | always |
| `--sonar-api-token` | credentials | 1 | always |
| `--sonar-ce-task-url` | location | 1 | always |
| `--sonar-project-key` | identity | 1 | never |
| `--sonar-revision` | identity | 1 | never |
| `--sonar-server-url` | location | 1 | never |
| `--sonar-working-dir` | location | 1 | always |
| `--sort` | output | 1 | never |
| `--sort-direction` | output | 3 | never |
| `--space-id` | filter | 1 | never |
| `--start` | identity | 1 | never |
| `--start-ts` | identity | 1 | always |
| `--tag` | metadata | 3 | never |
| `--template` | identity | 1 | always |
| `--template-file` | location | 2 | never |
| `--trail` | identity | 15 | some commands |
| `--type` | identity | 4 | some commands |
| `--unset` | metadata | 2 | never |
| `--upload-results` | switch | 2 | always |
| `--use-empty-template` | switch | 1 | always |
| `--user-data` | identity | 13 | never |
| `--visibility` | metadata | 1 | never |
| `--watch` | output | 2 | always |
| `--yes` | switch | 2 | always |
| `--zip` | output | 1 | always |
| **total** | **165 flags** | **700** | |
