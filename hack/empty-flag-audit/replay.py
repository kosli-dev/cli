#!/usr/bin/env python3
"""Ask the server what it does with an emptied field, by replaying a request.

audit.py answers "does the CLI refuse an empty value". This answers the other
half: when the CLI does not refuse it, does the server? That question cannot be
put through the CLI, because an empty value mostly produces the same request as
omitting the flag, so nothing empty ever reaches the server to be rejected.

So the request is captured and replayed. --debug logs the method, the URL and
the JSON body of every request the CLI sends, which gives both the request to
replay and, by comparing the run with the flag set against the run with it
omitted, the payload field that flag controls. The field is then emptied here
rather than by the CLI, which is why this keeps working after the CLI starts
refusing empty values.

Each combination is replayed twice: once unmodified, as a control that says the
replay itself works, and once with the field emptied. The difference between the
two answers is the server's, and it is an answer about every client rather than
about the CLI.

A third replay asks whether an accepted empty value reaches the record. The
resource the `set` run created still holds a real value, so replaying the emptied
payload without freshening it names that same resource, and the spec's own verify
steps say whether reading it back afterwards shows the empty value.

What the server did underneath is not measured, and it differs by endpoint: some
commands write a field in place, while the attest and report commands append a
document that a later read answers with, destroying nothing. Both appear here as
the empty value reaching the record, which is the question the API is being asked
- whether accepting an empty value for this field is right - rather than a claim
about anything being lost.

This is the one question freshening destroys: a fresh name every time leaves no
record holding a value for the empty one to reach.

Everything here talks to the local test server. It never points at app.kosli.com.
"""

import argparse
import datetime
import hashlib
import json
import pathlib
import re
import subprocess
import tempfile
import urllib.error
import urllib.parse
import urllib.request
import uuid

from audit import (GLOBALS, HOST, SERVER_CONTAINER, SPEC, TOKEN, expand,
                   invocation_for, normalise, prepare, reset_server, run)

RESULTS = "results-api.tsv"

# Everything volatile or bulky goes here rather than into RESULTS, whose worth is
# that two runs of it diff line by line. A timestamp or a whole response body in a
# column would end that.
EVIDENCE_FILE = "results-api-evidence.jsonl"
EVIDENCE = []

# The line --debug prints before a request body, naming what it is sending to
# where. requests.go logs the method beside the URL for this.
SENT = re.compile(r"payload sent to: (\w+) (\S+)")

# The line --debug prints after every request, whatever it carried. A read sends
# no body, so this is the only trace of it, and it holds the query string the
# flags became.
MADE = re.compile(r"request made to (\S+) and got status (\d+)")

# The commands whose requests are reads. That line does not name the method, and
# a command with no body could be a GET or a DELETE, so the ones probed this way
# are the ones whose verb settles it. Replaying a read is also the only kind of
# replay that is safe to repeat: nothing is created, so the control and the
# emptied request cannot interfere with each other.
READ_VERBS = ("get ", "list ", "log ", "diff ", "search")

# Payload fields that must differ between replays. A replay of a create is a
# create: sending the captured payload twice would update the first resource
# rather than make a second, and the control would then be measuring an update.
# Asking whether an empty value reaches the record needs the captured name kept,
# so that replay is the one that goes unfreshened.
FRESHEN = ["name"]


def captured_request(text):
    """Return (method, url, payload) read from a --debug log, or None.

    The payload is pretty-printed JSON following the line naming the request,
    so it is read by matching braces from the first one after that line rather
    than to the end of the block, which would swallow the log lines after it.
    """
    match = SENT.search(text)
    if not match:
        return None
    start = text.find("{", match.end())
    if start < 0:
        return None
    depth = 0
    for index in range(start, len(text)):
        if text[index] == "{":
            depth += 1
        elif text[index] == "}":
            depth -= 1
            if depth == 0:
                return match.group(1), match.group(2), json.loads(text[start:index + 1])
    return None


def captured_read(text, command):
    """Return (method, url, query) for a read, or None.

    A read carries its flags in the query string rather than a body, so what a
    flag controls is found by comparing two urls rather than two payloads. The
    method is not logged, which is why only commands whose verb settles it are
    read this way.
    """
    if not command.startswith(READ_VERBS):
        return None
    match = MADE.search(text)
    if not match:
        return None
    url = match.group(1)
    # keep_blank_values, or a parameter sent as blank reads as one that was never
    # sent, and the diff against the other run names the wrong parameter.
    return "GET", url, dict(urllib.parse.parse_qsl(
        urllib.parse.urlsplit(url).query, keep_blank_values=True))


def controlled_parameters(omitted, given):
    """Return the query parameters that differ between two captured urls."""
    if omitted is None or given is None:
        return []
    return sorted(k for k in set(omitted) | set(given)
                  if omitted.get(k) != given.get(k))


def with_parameter(url, parameter, value):
    """Return the url with one query parameter set to value.

    keep_blank_values, or a parameter the url already carries as blank is dropped
    rather than kept, so the replayed request differs from the captured one by more
    than the parameter under test and its verdict is about something else.
    """
    parts = urllib.parse.urlsplit(url)
    query = dict(urllib.parse.parse_qsl(parts.query, keep_blank_values=True))
    query[parameter] = value
    return urllib.parse.urlunsplit(parts._replace(
        query=urllib.parse.urlencode(query)))


def server_log(since):
    """Return the server's own log from when a request was sent, or why not.

    Fetched only for a 5xx, where the answer body says nothing and the traceback
    says everything: a 4xx is the server explaining itself and needs no log. A
    second is taken off the start because the request and the log entry are
    timestamped by different clocks.
    """
    window = (since - datetime.timedelta(seconds=1)).isoformat()
    try:
        done = subprocess.run(
            ["docker", "logs", SERVER_CONTAINER, "--since", window],
            capture_output=True, text=True, timeout=30)
    except (OSError, subprocess.SubprocessError) as exc:
        return f"unavailable: {exc}"
    return (done.stdout + done.stderr).strip().splitlines()[-200:]


def only_this_parameter_changed(control_url, emptied_url, parameter):
    """Return why the emptied url differs from the control beyond one parameter.

    A row's verdict is only about the parameter it names if that is the one thing
    the two requests do not share. Without this, a verdict can be recorded against
    a request that could not have produced it, and the row reads as a finding.
    """
    control, emptied_parts = (urllib.parse.urlsplit(control_url),
                              urllib.parse.urlsplit(emptied_url))
    if (control.scheme, control.netloc, control.path) != (
            emptied_parts.scheme, emptied_parts.netloc, emptied_parts.path):
        return (f"the emptied request went to {emptied_parts.path} rather than"
                f" {control.path}")
    # keep_blank_values, or a parameter that is already blank looks like one that
    # is absent, and dropping it counts as a second change.
    before = dict(urllib.parse.parse_qsl(control.query, keep_blank_values=True))
    after = dict(urllib.parse.parse_qsl(emptied_parts.query, keep_blank_values=True))
    changed = {key for key in set(before) | set(after)
               if before.get(key) != after.get(key)}
    if changed != {parameter}:
        return (f"the emptied request changed {sorted(changed)} rather than only"
                f" {parameter}")
    if after.get(parameter) != "":
        return f"the emptied request did not empty {parameter}"
    return None


def only_this_field_changed(control, emptied_payload, field):
    """Return why the emptied payload differs from the control beyond one field.

    FRESHEN's fields are expected to differ, because each replay of a create needs
    a name of its own.
    """
    if control is None or emptied_payload is None:
        return None
    changed = {key for key in set(control) | set(emptied_payload)
               if control.get(key) != emptied_payload.get(key)}
    unexpected = changed - {field} - set(FRESHEN)
    if unexpected:
        return f"the emptied request also changed {sorted(unexpected)}"
    if emptied_payload.get(field) not in ("", [""]):
        return f"the emptied request did not empty {field}"
    return None


def controlled_fields(omitted, given):
    """Return the payload fields that differ between two captured payloads.

    Comparing the run with the flag omitted against the run with it set says
    what that flag controls, without anyone writing the mapping down. The two
    runs with an empty value are no use for this: the point of the audit is
    that an empty value often produces the payload omitting the flag produces,
    and after the CLI starts refusing empty values it produces none at all.
    """
    if not omitted or not given:
        return []
    return sorted(k for k in set(omitted) | set(given)
                  if omitted.get(k) != given.get(k))


def without_fixtures(payload, command, runs):
    """Return the payload with each run's fixture names put back as placeholders.

    The audit's own rule, applied to a payload rather than to output: two runs
    of one command own different flows and trails, so their payloads differ by
    those names even when the flag changed nothing. Comparing them without this
    reports the fixture as a field the flag controls.

    Applied value by value rather than to the whole payload as text, because
    normalise ends by replacing volatile values with placeholders, and
    `"timestamp": {epoch}` is no longer JSON to read back.
    """
    def cleaned(value):
        """Return one payload value with any run's fixture names put back."""
        if isinstance(value, str):
            return normalise(value, command, runs)
        if isinstance(value, list):
            return [cleaned(each) for each in value]
        if isinstance(value, dict):
            return {key: cleaned(each) for key, each in value.items()}
        return value

    return cleaned(payload)


def emptied(payload, field):
    """Return the payload with one field emptied, as that field's own type."""
    copy = dict(payload)
    copy[field] = [""] if isinstance(payload.get(field), list) else ""
    return copy


def freshened(payload):
    """Return the payload with the fields that must be unique made unique."""
    copy = dict(payload)
    for field in FRESHEN:
        if field in copy:
            copy[field] = f"replay-{uuid.uuid4().hex[:8]}"
    return copy


def replay(method, url, payload, label="replay"):
    """Send one payload to the local server and return (status, answer).

    A 4xx arrives as an exception rather than a response, and it is the answer
    being looked for, so it is read from the exception instead of raised.

    The answer is returned whole, and the request and answer are both appended to
    EVIDENCE. The results file keeps only the first 120 characters, because a
    field holding a whole listing stops the file being read or diffed, and a
    truncated answer is the one thing that cannot be recovered afterwards: a row
    whose verdict surprises you is answered by its body, and by the request that
    earned it.
    """
    if not url.startswith(HOST):
        raise SystemExit(f"refusing to send anywhere but {HOST}: {url}")
    # A read carries nothing. Sending it an empty body rather than no body is a
    # different request from the one the CLI made.
    body = None if payload is None else json.dumps(payload).encode()
    headers = {"Content-Type": "application/json; charset=utf-8",
               "Authorization": f"Bearer {TOKEN}"}
    request = urllib.request.Request(url, method=method, data=body, headers=headers)
    started = datetime.datetime.now(datetime.timezone.utc)
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            status, answer = response.status, response.read().decode()
    except urllib.error.HTTPError as error:
        status, answer = error.code, error.read().decode()
    item = {
        "label": label,
        "at": started.isoformat(),
        "method": method,
        "url": url,
        # The headers are here because a request that cannot be seen has to be
        # guessed at: this GET carries a Content-Type, which is not obvious from
        # the row and took a reading of this function to discover.
        "headers": dict(headers, Authorization="Bearer <token>"),
        "sent": payload,
        "status": status,
        "answer": answer,
    }
    if status >= 500:
        item["server_log"] = server_log(started)
    EVIDENCE.append(item)
    return status, answer


def read_back(binary, entry, command, key, captured, home):
    """Return what the spec's verify steps read back, or None if it has none.

    The steps the audit uses to see what a command did, run here to see whether
    an emptied field reached the store. A command with no verify step cannot
    answer the question, which is reported rather than guessed at.
    """
    steps = entry.get("verify", [])
    if not steps:
        return None
    return "\n".join(
        run(binary, [expand(a, command, key, captured) for a in step] + GLOBALS,
            False, home)[2]
        for step in steps)


def stored_answer(binary, entry, command, method, url, carried, field, owned,
                  runs, home):
    """Say whether an accepted empty value reaches the record.

    Asked of the resource the `set` run already created, by replaying the
    emptied payload without freshening it, so the request names that resource
    rather than a new one. The comparison is of everything the verify steps
    print, before against after, because what a read calls a field is not always
    what the payload calls it.

    Says nothing about how the server got there, deliberately. A command that
    writes a field in place and one that appends a document a later read answers
    with both come out as the empty value reaching the record; which of the two
    happened is a property of the endpoint, and reading it off this measurement
    would be inventing it.
    """
    key, captured = owned
    before = read_back(binary, entry, command, key, captured, home)
    if before is None:
        return "not asked, the command has no verify step"
    status, _ = replay(method, url, emptied(carried, field),
                       label=f"{command} {field} emptied, unfreshened")
    if not 200 <= status < 300:
        return f"not an answer, the replay was itself refused with {status}"
    after = read_back(binary, entry, command, key, captured, home)
    # A changed read-back is evidence and outranks the status, which does not
    # say which record was reached: reporting an artifact that already exists
    # answers 201 and the empty value is what a read returns afterwards. The
    # status is consulted only when nothing changed, where a 201 leaves it open
    # whether the record was reached and kept its value or a separate record was
    # named instead.
    if normalise(before, command, runs) != normalise(after, command, runs):
        return "reaches the record"
    if status == 201:
        return ("not an answer, the reply was 201 and nothing changed, so the"
                " replay may have named a separate record")
    return "does not reach the record"


def capture(binary, entry, command, flag, how, home):
    """Run one of the audit's own invocations with --debug and read the request.

    Returns what was sent together with the fixtures this run owned, because
    naming them is what lets two runs' payloads be compared: each run creates
    its own flow and trail, so their payloads differ by those names whatever
    the flag did.
    """
    key = flag if how == "empty" else f"{flag}-{how}"
    failure, captured = prepare(binary, entry, command, key, False, home)
    if failure:
        return None
    argv = invocation_for(command, entry, key, flag, how, captured)
    _, _, text = run(binary, argv + ["--debug"], False, home)
    request = captured_request(text) or captured_read(text, command)
    return request and request + ((key, captured),)


def probe(binary, entry, command, flag, home):
    """Return the rows for one command-and-flag pair."""
    omitted = capture(binary, entry, command, flag, "omitted", home)
    given = capture(binary, entry, command, flag, "set", home)
    if not given:
        return [f"{command}\t--{flag}\t\t\t\t\t\tnothing to read"
                f"\tnot asked, nothing was replayed"
                f"\tthe run with the flag set sent no request"]

    method, url, carried, owned = given
    runs = [owned] + ([omitted[3]] if omitted else [])
    if method == "GET":
        return read_rows(command, flag, method, url, carried,
                         omitted[2] if omitted else None)

    fields = controlled_fields(
        without_fixtures(omitted[2], command, runs) if omitted else None,
        without_fixtures(carried, command, runs))
    if not fields:
        return [f"{command}\t--{flag}\t{method}\t{url}\t\t\t\tno field"
                f"\tnot asked, nothing was replayed"
                f"\tthe flag changes no field of the payload"]

    rows = []
    for index, field in enumerate(fields):
        control_payload = freshened(carried)
        emptied_payload = emptied(freshened(carried), field)
        control = replay(method, url, control_payload,
                         label=f"{command} --{flag} {field} control")
        answer = replay(method, url, emptied_payload,
                        label=f"{command} --{flag} {field} emptied")
        suspect = only_this_field_changed(control_payload, emptied_payload, field)
        outcome = f"suspect, {suspect}" if suspect else verdict(control[0], answer[0])
        if suspect:
            stored = "not asked, the emptied request was not the one intended"
        elif outcome != "the server accepts it":
            stored = f"not asked, {outcome}"
        elif index:
            stored = ("not asked, an earlier field of this flag already changed"
                      " the resource")
        else:
            stored = stored_answer(binary, entry, command, method, url, carried,
                                   field, owned, runs, home)
        rows.append(f"{command}\t--{flag}\t{method}\t{url}\t{field}"
                    f"\t{control[0]}\t{answer[0]}\t{outcome}\t{stored}"
                    f"\t{answer[1].strip()[:120]}")
    return rows


def read_rows(command, flag, method, url, query, omitted_query):
    """Return the rows for a read, whose flags are query parameters.

    The fixture names a read sends are in its path rather than its parameters,
    so two runs' parameters can be compared as they are. Emptying one is the
    same question the body path asks: the server is sent `page=` where it was
    sent `page=1`, and its answer is about every client rather than the CLI.
    """
    parameters = controlled_parameters(omitted_query, query)
    if not parameters:
        return [f"{command}\t--{flag}\t{method}\t{url}\t\t\t\tno field"
                f"\tnot asked, nothing was replayed"
                f"\tthe flag changes no query parameter"]
    rows = []
    for parameter in parameters:
        emptied_url = with_parameter(url, parameter, "")
        control = replay(method, url, None,
                         label=f"{command} --{flag} {parameter} control")
        answer = replay(method, emptied_url, None,
                        label=f"{command} --{flag} {parameter} emptied")
        suspect = only_this_parameter_changed(url, emptied_url, parameter)
        outcome = f"suspect, {suspect}" if suspect else verdict(control[0], answer[0])
        stored = ("not asked, the emptied request was not the one intended"
                  if suspect else "not asked, a read stores nothing")
        rows.append(f"{command}\t--{flag}\t{method}\t{url}\t{parameter}"
                    f"\t{control[0]}\t{answer[0]}\t{outcome}"
                    f"\t{stored}"
                    f"\t{answer[1].strip()[:120]}")
    return rows


def verdict(control, emptied_status):
    """Say what the two statuses mean, or that they mean nothing.

    A control that did not succeed says the replay was never a working request,
    so whatever the emptied one answered is about the replay rather than about
    the empty value. Saying so in the file matters more than it sounds: a
    control of 400 beside an emptied 400 reads as a refusal to anyone skimming,
    and it is not one.
    """
    if not 200 <= control < 300:
        return "unusable, the control failed"
    if 200 <= emptied_status < 300:
        return "the server accepts it"
    return "the server refuses it"


def provenance(binary):
    """Name what this run was against, so two runs can be told apart.

    A row that answers differently between runs raises one question first: was it
    the same server? Answering it from memory costs a rebuild and a re-run, so it
    is recorded here instead. The CLI is identified by the hash of the binary
    rather than by `kosli version`, which reports what was built, not which build
    is on disk.
    """
    def said(argv):
        """Return a command's whole output, or why there is none."""
        try:
            done = subprocess.run(argv, capture_output=True, text=True, timeout=30)
        except (OSError, subprocess.SubprocessError) as exc:
            return f"unavailable: {exc}"
        return (done.stdout or done.stderr).strip()

    binary_path = pathlib.Path(binary)
    changed = said(["git", "status", "--porcelain", "."])
    return {
        "label": "provenance",
        "at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "host": HOST,
        "server_image": said(
            ["docker", "inspect", SERVER_CONTAINER, "--format", "{{.Image}}"]),
        "server_started": said(
            ["docker", "inspect", SERVER_CONTAINER, "--format", "{{.State.StartedAt}}"]),
        "cli_binary": str(binary_path),
        "cli_sha256": hashlib.sha256(binary_path.read_bytes()).hexdigest()
        if binary_path.is_file() else "unavailable: not a file",
        "audit_commit": said(["git", "rev-parse", "HEAD"]),
        "audit_uncommitted": changed.splitlines(),
    }


def main():
    """Replay every named combination and write what the server answered."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="./kosli", help="CLI to capture from")
    parser.add_argument("--only", help="limit to commands containing this text")
    parser.add_argument("--flag", help="limit to this one flag")
    args = parser.parse_args()

    spec = json.loads(SPEC.read_text())
    reset_server()
    EVIDENCE.append(provenance(args.binary))
    home = tempfile.mkdtemp(prefix="kosli-replay-home-")

    rows = ["command\tflag\tmethod\turl\tfield\tcontrol\temptied\tverdict"
            "\tstored\tanswer"]
    for command, entry in sorted(spec.items()):
        if args.only and args.only not in command:
            continue
        # A shortcut, not the rule. These commands die at the service they need
        # before Kosli is contacted, so they have no request to replay - but
        # what decides that is whether one was captured, which is reported as
        # "nothing to read" and catches the cases nobody predicted.
        if entry.get("needs"):
            continue
        for flag in entry["flags_to_test"]:
            if args.flag and args.flag != flag:
                continue
            for row in probe(args.binary, entry, command, flag, home):
                print(row.replace("\t", "  ")[:140], flush=True)
                rows.append(row)

    out = SPEC.parent / RESULTS
    out.write_text("\n".join(rows) + "\n")
    evidence = SPEC.parent / EVIDENCE_FILE
    evidence.write_text(
        "".join(json.dumps(item) + "\n" for item in EVIDENCE))
    print(f"\nwrote {out}")
    print(f"wrote {evidence}")


if __name__ == "__main__":
    main()
