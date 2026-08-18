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

Everything here talks to the local test server. It never points at app.kosli.com.
"""

import argparse
import json
import re
import tempfile
import urllib.error
import urllib.parse
import urllib.request
import uuid

from audit import (HOST, SPEC, TOKEN, invocation_for, normalise, prepare,
                   reset_server, run)

RESULTS = "results-api.tsv"

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
    return "GET", url, dict(urllib.parse.parse_qsl(urllib.parse.urlsplit(url).query))


def controlled_parameters(omitted, given):
    """Return the query parameters that differ between two captured urls."""
    if omitted is None or given is None:
        return []
    return sorted(k for k in set(omitted) | set(given)
                  if omitted.get(k) != given.get(k))


def with_parameter(url, parameter, value):
    """Return the url with one query parameter set to value."""
    parts = urllib.parse.urlsplit(url)
    query = dict(urllib.parse.parse_qsl(parts.query))
    query[parameter] = value
    return urllib.parse.urlunsplit(parts._replace(
        query=urllib.parse.urlencode(query)))


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


def replay(method, url, payload):
    """Send one payload to the local server and return (status, first line).

    A 4xx arrives as an exception rather than a response, and it is the answer
    being looked for, so it is read from the exception instead of raised.
    """
    if not url.startswith(HOST):
        raise SystemExit(f"refusing to send anywhere but {HOST}: {url}")
    # A read carries nothing. Sending it an empty body rather than no body is a
    # different request from the one the CLI made.
    body = None if payload is None else json.dumps(payload).encode()
    request = urllib.request.Request(
        url, method=method, data=body,
        headers={"Content-Type": "application/json; charset=utf-8",
                 "Authorization": f"Bearer {TOKEN}"})
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            return response.status, response.read().decode()[:120]
    except urllib.error.HTTPError as error:
        return error.code, error.read().decode()[:120]


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
        return [f"{command}\t--{flag}\t\t\t\t\t\tnothing to read\tthe run with"
                f" the flag set sent no request"]

    method, url, carried, owned = given
    runs = [owned] + ([omitted[3]] if omitted else [])
    if method == "GET":
        return read_rows(command, flag, method, url, carried,
                         omitted[2] if omitted else None)

    fields = controlled_fields(
        without_fixtures(omitted[2], command, runs) if omitted else None,
        without_fixtures(carried, command, runs))
    if not fields:
        return [f"{command}\t--{flag}\t{method}\t{url}\t\t\t\tno field\tthe flag"
                f" changes no field of the payload"]

    rows = []
    for field in fields:
        control = replay(method, url, freshened(carried))
        answer = replay(method, url, emptied(freshened(carried), field))
        rows.append(f"{command}\t--{flag}\t{method}\t{url}\t{field}"
                    f"\t{control[0]}\t{answer[0]}\t{verdict(control[0], answer[0])}"
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
        return [f"{command}\t--{flag}\t{method}\t{url}\t\t\t\tno field\tthe flag"
                f" changes no query parameter"]
    rows = []
    for parameter in parameters:
        control = replay(method, url, None)
        answer = replay(method, with_parameter(url, parameter, ""), None)
        rows.append(f"{command}\t--{flag}\t{method}\t{url}\t{parameter}"
                    f"\t{control[0]}\t{answer[0]}\t{verdict(control[0], answer[0])}"
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


def main():
    """Replay every named combination and write what the server answered."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="./kosli", help="CLI to capture from")
    parser.add_argument("--only", help="limit to commands containing this text")
    parser.add_argument("--flag", help="limit to this one flag")
    args = parser.parse_args()

    spec = json.loads(SPEC.read_text())
    reset_server()
    home = tempfile.mkdtemp(prefix="kosli-replay-home-")

    rows = ["command\tflag\tmethod\turl\tfield\tcontrol\temptied\tverdict\tanswer"]
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
    print(f"\nwrote {out}")


if __name__ == "__main__":
    main()
