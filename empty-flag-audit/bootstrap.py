#!/usr/bin/env python3
"""Write the first version of spec.json by interrogating the CLI.

Which commands and flags exist is read from coverage.json, which is written
from the command tree itself. What each command needs in order to work is
found by running it: this starts from the command alone and reads the CLI's own
complaints to fix the invocation: `required flag(s) "flow" not set` adds
--flow, `accepts 2 arg(s), received 1` adds a positional, and so on. It stops
when the command works.

This is a one-off aid, not part of running the audit. Its output, spec.json, is
the thing that is kept and edited: an entry there says plainly what a command
needs, and correcting one command means editing one entry rather than teaching
a rule here to make an exception. Re-running this overwrites those edits.
"""

import argparse
import json
import re
import tempfile

from audit import (COVERAGE, GLOBALS, HOST, ORG, SPEC, TOKEN, expand,
                   fixtures, globals_for, normalise, reset_server, run)

# The real SHA256 of ARTIFACT_PATH, not an invented one. Two flags need them to
# agree: `--fingerprint` names an artifact the server must already hold, and
# `--artifact-type` makes the CLI compute the fingerprint from the file. An
# invented value satisfies neither, because the artifact the setup reports and
# the artifact the CLI fingerprints are then different things.
FINGERPRINT = "1bef738d0bb1e690500f99a5b57d958caf3a5eb3e00d9012e1f4369fc6812e01"
POLICY_FILE = "cmd/kosli/testdata/policy-files/test-policy.yml"
ARTIFACT_PATH = "cmd/kosli/testdata/person-schema.json"
ATTESTATION_SCHEMA = "cmd/kosli/testdata/person-schema.json"
# A policy that passes. A denying policy would make every `evaluate` run exit
# non-zero whatever its flags held, leaving nothing to compare an empty value to.
REGO_POLICY = "cmd/kosli/testdata/policies/allow-all.rego"

# Commands that answer a question rather than perform an action, and say no by
# exiting non-zero. "Artifact is not compliant" is the command working, so that
# non-zero exit is the baseline its flags are compared against.
VERDICT_COMMANDS = {
    "assert artifact",
    "assert snapshot",
    "assert status",
    "evaluate trail",
    "evaluate trails",
    "evaluate input",
}

# Commands that need a service we cannot reach. Recorded as skipped rather than
# guessed at, so the results never claim to know something they do not.
NEEDS_EXTERNAL = {
    "aws": ["snapshot ecs", "snapshot lambda", "snapshot s3"],
    "azure": ["snapshot azure"],
    "google cloud": ["snapshot cloud-run"],
    "kubernetes": ["snapshot k8s"],
    "github": ["attest pullrequest github", "assert pullrequest github"],
    "gitlab": ["attest pullrequest gitlab", "assert pullrequest gitlab"],
    "bitbucket": ["attest pullrequest bitbucket", "assert pullrequest bitbucket"],
    "azure devops": ["attest pullrequest azure", "assert pullrequest azure"],
    "jira": ["attest jira"],
    "sonarqube": ["attest sonar"],
    "snyk": ["attest snyk"],
    "a server that implements it": ["enable beta", "disable beta"],
    "a connected git provider": ["get repo"],
}

# Plausible values, by flag name. A value that is wrong for a flag shows up as
# a baseline failure, which this reports rather than hides.
VALUES = {
    "fingerprint": FINGERPRINT,
    "artifact-type": "file",
    "commit": "HEAD",
    # What `report artifact` calls the flag the attest commands call --commit.
    # Both are resolved against the repository the audit runs in, so both need
    # a reference that exists rather than an invented name.
    "git-commit": "HEAD",
    "origin-url": "http://example.com",
    "build-url": "http://example.com",
    "commit-url": "http://example.com",
    # Not gitlab.com. A run of `attest pullrequest gitlab` sends it some eighty
    # requests carrying a token that was never valid, and after about twenty-six
    # of them gitlab.com starts answering "temporarily unavailable due to
    # elevated error rates". Its client retries that with a backoff, so every
    # later call costs four seconds, which was most of the two minutes that
    # command took. Sending them to the local server instead is answered
    # definitively and at once, and asks nothing of a service outside.
    #
    # A dead port is worse than either: the client retries a refused connection
    # as well, and takes twelve seconds to give up. What makes this fast is an
    # answer, not the absence of one.
    "gitlab-base-url": HOST,
    "repo-root": ".",
    "output": "json",
    # Values the CLI or the server states outright, taken from what they say
    # when refusing an invented one. Each is here because the run with a real
    # value is the control the empty value is compared against, and a control
    # that fails leaves the comparison resting on one leg. `--redact-commit-info
    # probe-redact-commit-info` was refused on seven commands, and the audit
    # recorded the comparison anyway.
    "annotate": "probe=annotate",
    # An expression naming a range of snapshots. Any index does; a made-up word
    # is refused with "is not a valid snapshot index". --end names one end of
    # the same kind of interval.
    "interval": "1",
    "end": "1",
    # One of the columns the server will sort on, which it lists when refusing
    # anything else.
    "sort": "name",
    # A tag is a pair, and the CLI says so: "must be in the format of key=value".
    "flow-tag": "probe=flow-tag",
    # public or private, which is what the API schema says this field accepts.
    # The flag is deprecated and still parsed, so it is still measured.
    "visibility": "public",
    # A file holding a schema, not a name. The audit has one already.
    "schema": ATTESTATION_SCHEMA,
    # A rule in jq format, which is what the flag is for. The schema this audit
    # uses describes a person, so a rule about their age is one the type will
    # take.
    "jq": ".age >= 18",
    # --start and --end name ends of a snapshot interval; --start-ts and --end-ts
    # are epoch seconds, parsed as a number.
    "start": "1",
    "start-ts": "1",
    "end-ts": "4102444800",
    # Hours, parsed as an integer.
    "grace-period-hours": "24",
    # Another pair the CLI spells out: "must be formatted as key=value".
    "link": "probe=http://example.com",
    "set": "probe=set",
    "external-url": "probe=http://example.com",
    "external-fingerprint": f"probe={FINGERPRINT}",
    "redact-commit-info": "author",
    "repo-provider": "github",
    "repo-url": "http://example.com",
    "sort-direction": "asc",
    "provider": "github",
    "params": "{}",
    "attachments": ARTIFACT_PATH,
    # An epoch timestamp, which is one of the two spellings the CLI names when
    # it refuses anything else. A fixed one, so two runs stay comparable.
    "expires-at": "4102444800",
    "page": "1",
    "page-limit": "5",
    "reason": "probe",
    "original-attestation-type": "generic",
    "type": "env",
    "privilege": "reader",
    "digests-source": "logs",
    "physical": "probe-physical",
    "logical": "probe-logical",
    "path": ARTIFACT_PATH,
    "paths": ".",
    "paths-file": "empty-flag-audit/paths.yml",
    "template-file": "cmd/kosli/testdata/valid_template.yml",
    "scan-results": "cmd/kosli/testdata/snyk_scan_example.json",
    "user-data": ARTIFACT_PATH,
    "input-file": ARTIFACT_PATH,
    "attestation-data": ARTIFACT_PATH,
    "results-dir": "cmd/kosli/testdata/junit",
    "artifact-name": "probe-artifact",
    "policy-file": POLICY_FILE,
}

# Values that mean different things on different commands. --type is an
# environment type on `create environment` and a policy type on `create
# policy`, so one dictionary cannot serve both.
COMMAND_VALUES = {
    # The global flags are measured here, and a global flag's working value is
    # the one this audit is already using. An invented org or token is refused
    # by the server, which leaves the empty value compared against a run that
    # never worked - and reads as the CLI having refused something.
    "archive flow": {"api-token": TOKEN, "org": ORG, "host": HOST,
                     "max-api-retries": "3"},
    "create environment": {"type": "K8S"},
    "create policy": {"type": "env"},
    "list environments": {"type": "K8S"},
    # --policy names a Rego file to read, not the policy fixture the name
    # suggests.
    "evaluate trail": {"policy": REGO_POLICY},
    "evaluate trails": {"policy": REGO_POLICY},
    "evaluate input": {"policy": REGO_POLICY},
    # The audited fingerprint has provenance by the time this runs, and an
    # artifact with provenance cannot be allowlisted.
    "allow artifact": {"fingerprint": "a" * 64},
    "attest custom": {"type": "{name}"},
    "attest junit": {"results-dir": "cmd/kosli/testdata/junit"},
    "attest decision": {"compliant": "true"},
    # Joining needs two environments that both exist, and the logical one has to
    # have been created as logical.
    "join environment": {"physical": "{env}", "logical": "{logical}"},
    # --watch waits for filesystem changes and never returns, so "true" is not a
    # value any run of this audit can wait for.
    "snapshot path": {"watch": "false"},
    "snapshot paths": {"watch": "false"},
}

# Flags a command needs but does not ask for, because its default is wrong for
# this machine rather than missing.
COMMAND_EXTRA_FLAGS = {
    "attest junit": ["results-dir"],
    # Without --input-file the command reads stdin, which the audit does not
    # feed, and it fails on end-of-file before reaching its own flags.
    "evaluate input": ["input-file", "policy"],
    # By the time this runs, the audit has created hundreds of environments, and
    # asking for all of them times out. A page limit keeps the request small
    # enough to answer, which is all these flags need.
    "list environments": ["page-limit"],
}

# The environment a command reports to must be of the type it reports. An
# environment created as K8S refuses a docker snapshot.
ENV_TYPES = {
    "snapshot docker": "docker",
    "snapshot path": "server",
    "snapshot paths": "server",
    "snapshot server": "server",
    "assert snapshot": "server",
    "diff snapshots": "server",
    "get snapshot": "server",
    "list snapshots": "server",
    "log environment": "server",
    "allow artifact": "server",
}

# Which fixture a command's positional argument names. `archive flow <name>`
# must be given the flow that setup created, not an unrelated string.
NOUNS = [
    ("service-account", "account"),
    ("environment", "env"),
    ("snapshot", "env"),
    ("policy", "policy"),
    ("control", "control"),
    ("trail", "trail"),
    ("flow", "flow"),
]

# Commands whose positionals the noun rule cannot reach: an argument that is a
# file, or two arguments naming the same fixture.
COMMAND_POSITIONALS = {
    # One of four shells, and nothing else is accepted, so an invented name
    # cannot get this command as far as running.
    "completion": ["bash"],
    "create policy": ["{policy}", POLICY_FILE],
    "diff snapshots": ["{env}", "{env}"],
    "fingerprint": [ARTIFACT_PATH],
    "tag": ["flow", "{flow}"],
    "search": [FINGERPRINT],
    "snapshot path": ["{env}"],
    "update default-org": [ORG],
    "get artifact": ["{flow}@" + FINGERPRINT],
    "get attestation": ["probe-attestation"],
    # The server generates the key's id, so setup captures it and it arrives
    # here as a placeholder like any fixture the audit named itself.
    "get api-key": ["{apikey}"],
    "rotate api-key": ["{apikey}"],
    "delete api-key": ["{apikey}"],
}


def positional_value(command, index):
    """Return the value for the index'th positional argument of command."""
    override = COMMAND_POSITIONALS.get(command, [])
    if index < len(override):
        return override[index]
    if index > 0:
        return "{name}-" + str(index)
    for noun, key in NOUNS:
        if noun in command:
            return "{" + key + "}"
    return "{name}"


def value_for(flag, command):
    """Return the value to give flag, as a placeholder where it names a fixture."""
    override = COMMAND_VALUES.get(command, {}).get(flag)
    if override:
        return override
    key = {"environment": "env", "service-account": "account"}.get(flag, flag)
    if key in fixtures("", ""):
        return "{" + key + "}"
    return VALUES.get(flag, f"probe-{flag}")


def setup_commands(command, flag):
    """Return the commands that create what command needs before it can run.

    A command that archives a flow needs a flow to archive. Prerequisites are
    matched on what the command talks about, which its own path already says,
    rather than listed one command at a time.
    """
    n = fixtures(command, flag)
    steps = []
    add = lambda *argv: steps.append({"argv": list(argv)})
    touches_trail = command.startswith(("attest ", "begin trail", "evaluate trail")) \
        or "trail" in command or "artifact" in command or "attestation" in command
    if (touches_trail or "flow" in command or command == "tag") \
            and not command.startswith("create flow"):
        add("create", "flow", n["flow"], "--use-empty-template")
    if touches_trail and not command.startswith("begin trail"):
        add("begin", "trail", n["trail"], "--flow", n["flow"])
    # A command that reads an artifact needs one reported into its trail. The
    # path is required even alongside --fingerprint, which is what the recorded
    # fingerprint is taken from.
    if command in ("get artifact", "get attestation", "assert artifact"):
        add("attest", "artifact", ARTIFACT_PATH, "--name", "probe-artifact",
            "--fingerprint", FINGERPRINT, "--flow", n["flow"], "--trail", n["trail"],
            "--build-url", "http://example.com", "--commit-url", "http://example.com")
    if command == "get attestation":
        add("attest", "generic", "--name", "probe-attestation",
            "--flow", n["flow"], "--trail", n["trail"])
    # An override needs an attestation to override.
    if command == "attest override":
        add("attest", "generic", "--name", n["name"],
            "--flow", n["flow"], "--trail", n["trail"])
    if ("environment" in command or "snapshot" in command or command in ENV_TYPES
            or command.startswith("join")) and not command.startswith("create environment"):
        add("create", "environment", n["env"], "--type", ENV_TYPES.get(command, "K8S"))
    # A command that reads a snapshot needs one to have been reported.
    if command in ("assert snapshot", "diff snapshots", "get snapshot",
                   "list snapshots", "log environment"):
        add("snapshot", "path", n["env"], "--path", ARTIFACT_PATH,
            "--name", "probe-artifact")
    if (command.startswith("attest custom") or "attestation-type" in command) \
            and not command.startswith("create attestation-type"):
        add("create", "attestation-type", n["name"], "--schema", ATTESTATION_SCHEMA)
    # Joining puts one environment inside another, so both have to exist. The
    # logical one is created around a seed environment rather than empty: a
    # logical environment with no included environments cannot be read back, and
    # `list environments` then returns 500 for every environment in the org,
    # which this audit measures later in the same run. The seed is not the
    # environment being joined, so the join still has something to do.
    if command.startswith("join environment"):
        add("create", "environment", n["seed"], "--type", "K8S")
        add("create", "environment", n["logical"], "--type", "logical",
            "--included-environments", n["seed"])
    if "control" in command and not command.startswith("create control"):
        add("create", "control", n["control"], "--name", n["control"])
    if "policy" in command and not command.startswith("create policy"):
        add("create", "policy", n["policy"], POLICY_FILE)
    if ("service-account" in command or "api-key" in command) \
            and not command.startswith("create service-account"):
        add("create", "service-account", n["account"], "--privilege", "reader")
    if "api-key" in command and not command.startswith("create api-key"):
        steps.append({
            "argv": ["create", "api-key", "--service-account", n["account"],
                     "--description", "probe"],
            "capture": ("apikey", r"ID:\s+(\S+)"),
        })
    return steps


def verify_commands(command):
    """Return read-only commands that show what a command did to the server.

    Some differences never reach the CLI's own output. `attest generic` reports
    "reported to trail" whether the attestation landed on the artifact or on
    the trail itself, so an emptied --fingerprint looks identical from outside
    and shows only in the trail's state.

    Read-only commands get none of these: their output already is everything
    they did.
    """
    if command.startswith(("list ", "get ", "search", "assert ", "diff ",
                           "evaluate ", "log ", "fingerprint", "status",
                           "version")):
        return []
    steps = []
    if command.startswith(("attest ", "begin trail")) or "trail" in command:
        steps.append(["get", "trail", "{trail}", "--flow", "{flow}",
                      "--output", "json"])
    elif "flow" in command:
        steps.append(["get", "flow", "{flow}", "--output", "json"])
    if "environment" in command or "snapshot" in command:
        steps.append(["get", "environment", "{env}", "--output", "json"])
    if "control" in command:
        steps.append(["get", "control", "{control}", "--output", "json"])
    if "policy" in command:
        steps.append(["get", "policy", "{policy}", "--output", "json"])
    if "service-account" in command:
        steps.append(["get", "service-account", "{account}", "--output", "json"])
    return steps


def prepare(binary, command, flag, as_ci, home):
    """Create this invocation's fixtures, returning (failure, captured)."""
    captured = {}
    for step in setup_commands(command, flag):
        argv = [expand(a, command, flag, captured) for a in step["argv"]]
        code, msg, text = run(binary, argv + GLOBALS, as_ci, home)
        if code != 0:
            return f"setup '{' '.join(argv[:3])}' failed: {msg}", captured
        if step.get("capture"):
            key, pattern = step["capture"]
            m = re.search(pattern, text)
            if not m:
                return f"setup '{' '.join(argv[:3])}' printed no {key}", captured
            captured[key] = m.group(1)
    return None, captured


def find_invocation(binary, command, kinds, as_ci, home):
    """Find an invocation of command that works, by reading its complaints.

    Gives up after a bounded number of rounds and reports the last error, so a
    command that cannot be satisfied is visible rather than silently wrong.
    """
    failure, captured = prepare(binary, command, "baseline", as_ci, home)
    if failure:
        return {"args": [], "flags": {}, "baseline_ok": False, "error": failure}

    # A flag the CLI asks for has to be given a value its type accepts. A
    # boolean handed "probe-new-compliance-status" fails to parse, which looks
    # like the command refusing something when it is only the audit guessing.
    pick = lambda name: flag_value(name, kinds.get(name), command)
    extra = {name: pick(name)
             for name in COMMAND_EXTRA_FLAGS.get(command, [])}
    positionals = []
    settled = lambda code, text: {
        "args": positionals, "flags": extra, "baseline_ok": True,
        "baseline_exit": code,
        "baseline_output": normalise(text, command, [("baseline", captured)]),
    }
    for _ in range(12):
        grow = lambda v: expand(v, command, "baseline", captured)
        invocation = (command.split() + [grow(p) for p in positionals]
                      + globals_for(command))
        for name, val in extra.items():
            invocation.append(f"--{name}={grow(val)}")
        code, msg, text = run(binary, invocation, as_ci, home)
        if code == 0:
            return settled(0, text)

        # Cobra writes "required flag(s) ... not set". A command checking a flag
        # itself writes what its author typed, and `snapshot server` writes
        # "required flag" without the (s), so both spellings are read here.
        if re.search(r'required flag(?:\(s\))? "([^"]+)"', msg):
            for name in re.findall(r'"([^"]+)"', msg):
                extra[name] = pick(name)
            continue
        m = re.search(r"accepts (?:at most )?(\d+) arg\(s\), received (\d+)", msg)
        if m and int(m.group(2)) < int(m.group(1)):
            positionals.append(positional_value(command, len(positionals)))
            continue
        m = re.search(r"(?:requires|accepts) at least (\d+) arg\(s\)", msg)
        if m and len(positionals) < int(m.group(1)):
            positionals.append(positional_value(command, len(positionals)))
            continue
        m = re.search(r"accepts between (\d+) and \d+ arg\(s\), received (\d+)", msg)
        if m and int(m.group(2)) < int(m.group(1)):
            positionals.append(positional_value(command, len(positionals)))
            continue
        m = re.search(r"at least one of ([^\n]+) is required", msg)
        if m:
            name = m.group(1).split(",")[0].strip().lstrip("-")
            extra[name] = pick(name)
            continue
        if "requires either a positional" in msg or "argument is required" in msg:
            positionals.append(positional_value(command, len(positionals)))
            continue
        # "exactly one of the REPO-NAME argument or --repo-id must be provided"
        # asks for a positional, not a flag, so it is matched before the
        # flag-or-flag case below.
        if re.search(r"[A-Z-]+ argument or --", msg):
            positionals.append(positional_value(command, len(positionals)))
            continue
        # "either --artifact-type or --fingerprint must be specified".
        # Whichever of the two has a value here is chosen: --artifact-type
        # would also need a path to fingerprint, and --fingerprint stands alone.
        m = re.search(r"--([a-z][a-z0-9-]*) or --([a-z][a-z0-9-]*)", msg)
        if m and ("must be" in msg or "is required" in msg):
            choices = [n for n in m.groups() if n in VALUES and n not in extra]
            choices.sort(key=lambda n: n != "fingerprint")
            if choices:
                extra[choices[0]] = pick(choices[0])
                continue
        if "path is required" in msg:
            positionals.append(ARTIFACT_PATH)
            continue
        m = re.search(r"re-run with --([a-z][a-z0-9-]*)", msg)
        if m and m.group(1) not in extra:
            extra[m.group(1)] = "true"
            continue
        m = re.search(r"--([a-z][a-z0-9-]*) is required when", msg)
        if m and m.group(1) not in extra:
            extra[m.group(1)] = pick(m.group(1))
            continue
        # Nothing in the message asks for a better invocation, so this is as
        # well-formed as the command gets. For a command whose answer is a
        # verdict, that non-zero exit is the answer, not a broken baseline.
        if command in VERDICT_COMMANDS:
            return settled(code, text)
        return {"args": positionals, "flags": extra, "baseline_ok": False,
                "error": msg}
    return {"args": positionals, "flags": extra, "baseline_ok": False,
            "error": "gave up after 12 rounds"}


def flag_value(flag, kind, command):
    """Return a real value for flag, for the run an empty value is compared to.

    A per-command override wins even for a boolean, because "true" is not always
    a value this audit can wait for: `snapshot path --watch=true` never returns.
    """
    override = COMMAND_VALUES.get(command, {}).get(flag)
    if override:
        return override
    if kind == "bool":
        return "true"
    return value_for(flag, command)


def every_command():
    """Return {command: {flag: what it takes}} for the whole CLI.

    Read from coverage.json rather than from --help. A help page is a listing
    of what the CLI chooses to show: cobra leaves out a command marked Hidden
    or Deprecated, and pflag leaves out a hidden flag, so reading help pages
    finds only the commands nobody minded showing. coverage.json is written
    from the command tree itself, where nothing is filtered.
    """
    return json.loads(COVERAGE.read_text())


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="./kosli", help="CLI to interrogate")
    parser.add_argument("--only", help="limit to commands containing this text")
    parser.add_argument("--ci", action="store_true",
                        help="run as if inside GitHub Actions")
    args = parser.parse_args()
    reset_server()
    home = tempfile.mkdtemp(prefix="kosli-audit-home-")

    commands = every_command()
    skip = {c: why for why, cs in NEEDS_EXTERNAL.items() for c in cs}
    # Start from what is already there, so re-running with --only rewrites one
    # command's entry and leaves everyone else's, including any edited by hand.
    spec = json.loads(SPEC.read_text()) if SPEC.exists() else {}
    for command, flags in sorted(commands.items()):
        if args.only and args.only not in command:
            continue
        reason = next((w for c, w in skip.items() if command.startswith(c)), None)
        entry = find_invocation(args.binary, command, flags, args.ci, home)
        if reason:
            # Recorded, not abandoned. The command cannot succeed here, but its
            # own checks run before it reaches the service it needs, so an empty
            # value can still be seen being refused - or getting past everything
            # the CLI has and dying at the service instead.
            entry["needs"] = reason
            entry["baseline_ok"] = False
        entry["flags_to_test"] = sorted(flags)
        # A real value for each flag, so an empty value can be compared with a
        # set variable as well as with an absent one.
        entry["flag_values"] = {
            name: flag_value(name, kind, command) for name, kind in sorted(flags.items())
        }
        # Freeze the setup steps into the spec as placeholders, so the audit
        # reads them rather than deriving them again.
        entry["setup"] = [
            dict(step, argv=[normalise(a, command, [("baseline", {})])
                             for a in step["argv"]])
            for step in setup_commands(command, "baseline")
        ]
        entry["verify"] = verify_commands(command)
        spec[command] = entry
        state = "ok  " if entry["baseline_ok"] else ("needs" if reason else "FAIL")
        print(f"{state}  {command}"
              + ("" if entry["baseline_ok"] else f"  {entry.get('error','')[:90]}"))
    SPEC.write_text(json.dumps(spec, indent=1))
    print(f"\nwrote {SPEC}")


if __name__ == "__main__":
    main()
