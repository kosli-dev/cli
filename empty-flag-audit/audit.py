#!/usr/bin/env python3
"""Find out what the CLI does when a flag is given an empty value.

Reads spec.json, which says for each command how to set that command up and
what a working invocation of it looks like, then runs each of the command's
flags with an empty value and records what changed. See README.md for why this
cannot be answered by reading the flags' types.

spec.json is the source of truth and is meant to be edited by hand, one command
at a time. bootstrap.py wrote the first version of it by interrogating the CLI.
"""

import argparse
import json
import os
import pathlib
import re
import subprocess
import tempfile

HERE = pathlib.Path(__file__).parent
REPO = HERE.parent
SPEC = HERE / "spec.json"
# Every command and flag the CLI has, written from the command tree itself by
# TestEmptyFlagAuditCoversEveryCommandAndFlag in cmd/kosli. Reading the tree is
# the only way to see all of them: cobra prints no listing for a hidden or a
# deprecated command, and pflag prints none for a hidden flag, so walking
# --help - which is how spec.json was first written - silently misses them.
COVERAGE = HERE / "coverage.json"
# One file per pass, so a --ci run does not overwrite the laptop run. report.py
# reads both and needs them side by side.
RESULTS = HERE / "results.tsv"
RESULTS_CI = HERE / "results-ci.tsv"

HOST = "http://localhost:8001"
# A shared org, not a personal one. Real customers do real work in a shared
# org, and some commands - service accounts, for one - are refused outright in
# a personal org, which would show up as a missing result rather than a real
# one.
ORG = "docs-cmd-test-user-shared"
TOKEN = (
    "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
    ".eyJpZCI6ImNkNzg4OTg5In0"
    ".e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"
)
GLOBALS = ["--host", HOST, "--org", ORG, "--api-token", TOKEN]

# Variables a CI system sets, which the CLI uses to fill in flag defaults. The
# audit sets them only when asked, so a result never depends on where it ran.
CI_ENV = {
    "GITHUB_RUN_NUMBER": "1",
    "GITHUB_SERVER_URL": "https://github.com",
    "GITHUB_REPOSITORY": "kosli-dev/cli",
    "GITHUB_RUN_ID": "99",
    "GITHUB_REPOSITORY_OWNER": "kosli-dev",
    # A real commit, because the CLI resolves this one against the repository
    # it is run in. A made-up SHA fails to resolve, which would stop commands
    # before they reached the flag being tested.
    "GITHUB_SHA": subprocess.run(
        ["git", "rev-parse", "HEAD"], cwd=REPO, capture_output=True, text=True,
    ).stdout.strip() or "0" * 40,
}


def environment(as_ci, home):
    """Build the environment for one CLI run.

    Everything is set explicitly. The CLI reads KOSLI_* variables, a config
    file under $HOME, and CI variables that fill in flag defaults, so
    inheriting the caller's environment would make a result depend on the
    machine it ran on and on the day it ran.
    """
    env = {"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "HOME": str(home)}
    if as_ci:
        env.update(CI_ENV)
    return env


# Cobra prints this above the output of a deprecated command, which still runs.
DEPRECATION_NOTICE = re.compile(r'^Command "[^"]+" is deprecated')


def noise(line):
    """Report whether line is something the CLI says alongside its result.

    A deprecated command still does its work, and a warning is not a result
    either. Reading either as the command's first meaningful line hides what
    the command actually did, and for a deprecated command it hides it behind
    a sentence that is the same whatever the flags were.
    """
    return "[warning]" in line or DEPRECATION_NOTICE.match(line)


def run(binary, args, as_ci=False, home="/tmp"):
    """Run the CLI once and return (exit code, first meaningful line, all output).

    A command that never returns is reported as a result rather than raised. One
    of them - `snapshot path --watch` waits for filesystem changes by design -
    would otherwise end a run of several hundred combinations at whichever one
    it reached.
    """
    try:
        proc = subprocess.run(
            [binary] + args, capture_output=True, text=True, timeout=120,
            env=environment(as_ci, home),
        )
    except subprocess.TimeoutExpired:
        return 124, "timed out after 120 seconds", "timed out after 120 seconds"
    text = (proc.stderr + proc.stdout).strip()
    lines = text.splitlines()
    first = next((l for l in lines if l.strip() and not noise(l)), "")
    return proc.returncode, first, text


def fixtures(command, flag):
    """Name the resources one invocation owns.

    Fixtures belong to a command-and-flag pair. Two commands must not share, or
    one command's state would decide another's result. Two flags of one command
    must not share either: `create flow` creates the flow the first time and
    updates it the second, and `rename flow` renames the fixture out from under
    whatever runs next.

    The names are the same on every run, because the audit starts by wiping the
    server. That is what makes two runs' results comparable line by line.
    """
    key = f"{command} {flag}"
    base = re.sub(r"[^a-z0-9]+", "-", key.lower()).strip("-")
    return {kind: f"probe-{base}"[:52].rstrip("-") + "-" + suffix
            for kind, suffix in (
                ("flow", "fl"), ("trail", "tr"), ("env", "en"),
                ("policy", "po"), ("control", "co"), ("account", "sa"),
                ("logical", "lg"), ("name", "nm"),
            )}


def expand(value, command, flag, captured):
    """Replace placeholders in a spec value with this invocation's names.

    The spec stores `{flow}` rather than a name, which is what lets one entry
    serve every flag of its command, each with its own fixtures. Substitution
    is by known name rather than str.format, so a value that is itself braces -
    the empty JSON document some flags take - survives untouched.
    """
    names = dict(fixtures(command, flag), **captured)
    return re.sub(r"\{(" + "|".join(names) + r")\}",
                  lambda m: names[m.group(1)], value)


# Things that differ between two runs of the same invocation for reasons that
# have nothing to do with the flag being tested: the clock, and ids the server
# mints. Left in, they would report every timestamped command as changed.
VOLATILE = [
    (re.compile(r"\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4,12}\b"),
     "{uuid}"),
    (re.compile(r"\b\w{3}, \d{1,2} \w{3} \d{4} \d{2}:\d{2}:\d{2} \w+"), "{timestamp}"),
    (re.compile(r"\d{4}-\d{2}-\d{2}T[\d:.+-]+"), "{timestamp}"),
    (re.compile(r"(about |less than |over |almost )?(an?|\d+) "
                r"(second|minute|hour|day|week|month|year)s? ago"), "{ago}"),
    # A created API key is printed once and is different every time. It is not
    # a uuid, so the pattern above does not reach it.
    (re.compile(r"(?m)^(Key:\s+)\S+$"), r"\1{secret}"),
    # The state read back by a verify step timestamps everything in epoch
    # seconds - created_at, last_modified_at, evaluated_at - which differ
    # between two runs however alike the runs were.
    (re.compile(r"\b1[6-9]\d{8}(\.\d+)?\b"), "{epoch}"),
]


def stabilise(output):
    """Blank out the parts of an output that change on their own."""
    for pattern, name in VOLATILE:
        output = pattern.sub(name, output)
    return output


def normalise(output, command, runs):
    """Put placeholders back where any of these runs' names appear.

    Two runs of one command differ in their fixture names, and in any id the
    server minted for them, so their outputs differ even when they say the same
    thing. Both runs' names are replaced in both outputs, not just each run's
    own: a command that lists everything in the org sees the other run's
    fixtures too, and leaving those as literal names would make two identical
    answers look different.

    Longer names are replaced first, so a name that contains another is not
    half-substituted.
    """
    pairs = []
    for flag, captured in runs:
        for key, name in dict(fixtures(command, flag), **captured).items():
            pairs.append((name, "{" + key + "}"))
    for name, placeholder in sorted(pairs, key=lambda p: -len(p[0])):
        output = output.replace(name, placeholder)
    return stabilise(output)


def prepare(binary, entry, command, flag, as_ci, home):
    """Create this invocation's fixtures.

    Returns (failure, captured). A fixture the server names rather than the
    audit - an API key gets a generated id - is read back out of the output of
    the step that created it, and joins the placeholders the invocation can use.
    """
    captured = {}
    for step in entry.get("setup", []):
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


def refused_by(binary, argv, as_ci, home):
    """Say who refused an empty value: the CLI itself, or something downstream.

    The two are worth telling apart. A flag the CLI checks is protected
    wherever it runs; a flag only the server rejects is unprotected in the CLI,
    and whether it is caught at all depends on someone else. `--build-url ""`
    is refused by the CLI on a laptop and reaches the server from inside GitHub
    Actions, where a filled-in CI default stops the flag being marked required.

    The answer is read from --debug, which logs every request and its status,
    rather than guessed at from the wording of the error.
    """
    _, _, text = run(binary, argv + ["--debug"], as_ci, home)
    if re.search(r"request made to .* and got status [45]\d\d", text):
        return "downstream"
    return "cli"


def same(a, b):
    """Say whether two runs did the same thing, as (exit code, what it did).

    The whole output is compared, not its first line. Two runs of a list
    command share a header and differ in the rows beneath it, so a first-line
    comparison would call an emptied filter unchanged when it had quietly
    changed the answer.
    """
    return "same" if a == b else "differs"


def observe(binary, entry, command, flag, captured, argv, as_ci, home):
    """Run one invocation and return (exit code, first line, what it did).

    What it did is the command's own output followed by the state the verify
    steps read back. Some commands leave no trace in their own output - an
    attestation says "reported to trail" whether it landed on the artifact or
    on the trail - so comparing only what was printed would call those
    unchanged. A verify step that fails is included as it stands: not being
    able to read the state back is itself a difference worth seeing.
    """
    code, first, text = run(binary, argv, as_ci, home)
    parts = [text]
    for step in entry.get("verify", []):
        step_argv = [expand(a, command, flag, captured) for a in step]
        _, _, seen = run(binary, step_argv + GLOBALS, as_ci, home)
        parts.append(seen)
    return code, first, "\n".join(parts)


def invocation_for(command, entry, key, flag, how, captured):
    """Build the spec's working command, with flag omitted, set, or emptied.

    `key` names the fixtures this run owns and differs between the three runs;
    `flag` is the flag under test and is the same in all three. Using one for
    the other builds `--fingerprint-set` and leaves the real flag in place.

    An empty value is worth comparing with both. Against the flag omitted it
    answers "does writing --flag '' do the same as not writing it?". Against
    the flag set it answers the question a pipeline actually asks: "does
    --flag "$VAR" still do its job when VAR is unset?". The two answers differ:
    an attestation with no --fingerprint at all and one with an empty
    --fingerprint both land on the trail, while a real fingerprint lands on the
    artifact.
    """
    grow = lambda v: expand(v, command, key, captured)
    argv = command.split() + [grow(a) for a in entry["args"]] + GLOBALS
    # The joined form works for every flag type, including the booleans a
    # separate token would leave dangling as a positional argument.
    for name, val in entry["flags"].items():
        if name != flag:
            argv.append(f"--{name}={grow(val)}")
    if how == "omitted":
        return argv
    if how == "empty":
        return argv + [f"--{flag}", ""]
    value = entry["flags"].get(flag) or entry.get("flag_values", {}).get(flag, "")
    return argv + [f"--{flag}={grow(value)}"]


SERVER_CONTAINER = "cli_kosli_server"


def reset_server():
    """Empty the database of the server that is already running.

    Without this the audit would meet the fixtures its last run left behind,
    and would have to invent a fresh name each time to avoid them.
    Deterministic names are worth more: they make two runs' results comparable
    line by line.

    This wipes data only. Restarting the server would pull its image, which
    means an AWS login, and no audit run should demand one. Start the server
    with `make test_setup` first; this refuses rather than starting one.
    """
    for script in ("clean_database.py", "create_standalone_test_users.py"):
        done = subprocess.run(
            ["docker", "exec", SERVER_CONTAINER, f"/app/test/{script}"],
            capture_output=True, text=True,
        )
        if done.returncode != 0:
            raise SystemExit(
                f"could not run {script} in {SERVER_CONTAINER}: "
                f"{(done.stderr or done.stdout).strip()}\n"
                "Is the local server running? Start it with: make test_setup"
            )


# Commands whose own results are fine but which leave the server unable to
# answer something else. `create environment --included-environments ""` writes a
# record that makes `list environments` return 500 for the whole org, so it is
# measured after everything that needs the server intact. Remove an entry here
# once the bug behind it is fixed.
MEASURE_LAST = ["create environment"]


def in_order(spec):
    """Return the commands to measure, poisoners last."""
    names = sorted(spec)
    deferred = [c for c in names if c in MEASURE_LAST]
    return [c for c in names if c not in deferred] + deferred


def merged(path, rows):
    """Fold new rows into whatever the file already holds.

    A run limited with --only or --flag measures a handful of combinations.
    Writing just those would throw away every other result in the file, so the
    rows it did measure replace their old selves and the rest are left alone.
    """
    header, fresh = rows[0], rows[1:]
    kept = {}
    if path.exists():
        for line in path.read_text().splitlines()[1:]:
            kept[tuple(line.split("\t")[:2])] = line
    for line in fresh:
        kept[tuple(line.split("\t")[:2])] = line
    return [header] + sorted(kept.values())


def wanted(spec, only, only_flag):
    """Return the command-and-flag pairs this run will actually measure."""
    pairs = []
    for command, entry in sorted(spec.items()):
        if (only and only not in command) or entry.get("skip"):
            continue
        for flag in entry["flags_to_test"]:
            if not only_flag or only_flag == flag:
                pairs.append((command, flag))
    return pairs


def audit(binary, spec, only, only_flag, as_ci, home):
    """Run every command-and-flag combination in spec, returning result rows."""
    rows = ["command\tflag\tempty_exit\tomitted_exit\tset_exit\trefused_by"
            "\tvs_omitted\tvs_set\tmessage"]
    total, done = len(wanted(spec, only, only_flag)), 0
    for command in in_order(spec):
        entry = spec[command]
        if only and only not in command:
            continue
        # A command needing a service we cannot reach is still run. Its own
        # checks happen before it gets that far, so an empty value can be seen
        # being refused, or seen surviving every check the CLI has and failing
        # at the service instead. Only what the service would have done stays
        # unknown.
        for flag in entry["flags_to_test"]:
            if only_flag and only_flag != flag:
                continue
            # One set of fixtures per run, all created before any run happens.
            # Separate fixtures keep a mutating command from meeting its own
            # leftovers - `create flow` would otherwise create on the first run
            # and update on the next. Creating them all up front leaves nothing
            # changing between the runs, which is what a command that lists
            # everything in the org needs.
            how = {"omitted": flag + "-omitted", "set": flag + "-set",
                   "empty": flag}
            prepared, failure = {}, None
            for kind, key in how.items():
                fail, captured = prepare(binary, entry, command, key, as_ci, home)
                prepared[kind] = captured
                failure = failure or fail
            if failure:
                rows.append(f"{command}\t--{flag}\t\t\t\t\tsetup failed"
                            f"\tsetup failed\t{failure}")
                done += 1
                print(f"[{done}/{total}] {'setup failed':12} {command} --{flag}",
                      flush=True)
                continue

            runs = [(key, prepared[kind]) for kind, key in how.items()]
            seen = {}
            for kind, key in how.items():
                code, first, text = observe(
                    binary, entry, command, key, prepared[kind],
                    invocation_for(command, entry, key, flag, kind, prepared[kind]),
                    as_ci, home)
                seen[kind] = (code, normalise(text, command, runs), first)

            code, empty_state, msg = seen["empty"]
            who = ""
            if code != 0:
                who = refused_by(
                    binary,
                    invocation_for(command, entry, how["empty"], flag, "empty",
                                   prepared["empty"]),
                    as_ci, home)
            rows.append(
                f"{command}\t--{flag}\t{code}\t{seen['omitted'][0]}"
                f"\t{seen['set'][0]}\t{who}"
                f"\t{same(seen['empty'][:2], seen['omitted'][:2])}"
                f"\t{same(seen['empty'][:2], seen['set'][:2])}\t{msg[:160]}")
            # Flushed, because stdout is a pipe or a file whenever this is run
            # in the background, and a buffered run looks identical to a hung
            # one for minutes at a time.
            done += 1
            print(f"[{done}/{total}] {'refused' if code else 'accepted':12}"
                  f" {command} --{flag}", flush=True)
    return rows


def unaudited(spec):
    """Return the command-and-flag pairs the CLI has and spec.json does not.

    A missing pair is not a small omission. The whole claim this audit makes is
    that it measured every flag of every command, and a pair it never ran is
    counted nowhere: it is absent from the totals rather than reported as
    unknown, so the figures read as complete while covering less than they say.
    """
    covered = json.loads(COVERAGE.read_text())
    missing = []
    for command, flags in sorted(covered.items()):
        entry = spec.get(command)
        if entry is None:
            missing.append(f"{command} (no entry at all)")
            continue
        tested = set(entry.get("flags_to_test", []))
        missing += [f"{command} --{flag}" for flag in flags if flag not in tested]
    return missing


def require_full_coverage(spec):
    """Stop the run when the CLI has a command or flag spec.json does not.

    Stopping rather than warning is the point. A warning at the top of a run of
    several hundred combinations is gone by the time the results are read, and
    the results file it wrote looks exactly like a complete one.
    """
    missing = unaudited(spec)
    if not missing:
        return
    raise SystemExit(
        f"{len(missing)} command-and-flag combinations exist in the CLI but are "
        f"not in {SPEC.name}:\n\n  "
        + "\n  ".join(missing)
        + f"\n\nAdd them to {SPEC.name}, by hand or with bootstrap.py, and run "
        f"again. If one of them is genuinely not worth auditing, say so in "
        f"{COVERAGE.name} by way of the test that writes it, so the reason is "
        f"recorded rather than remembered."
    )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="./kosli", help="CLI to audit")
    parser.add_argument("--only", help="limit to commands containing this text")
    parser.add_argument("--flag", help="limit to this one flag")
    parser.add_argument("--ci", action="store_true",
                        help="run as if inside GitHub Actions, so flag defaults"
                             " come from CI variables")
    args = parser.parse_args()

    # Checked before the server is reset, so a run that cannot be complete
    # stops without having thrown away the results of the last one that was.
    spec = json.loads(SPEC.read_text())
    require_full_coverage(spec)

    reset_server()
    home = tempfile.mkdtemp(prefix="kosli-audit-home-")
    rows = audit(args.binary, spec, args.only, args.flag, args.ci, home)
    out = RESULTS_CI if args.ci else RESULTS
    out.write_text("\n".join(merged(out, rows)) + "\n")
    print(f"\nwrote {out}")


if __name__ == "__main__":
    main()
