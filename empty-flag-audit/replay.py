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
import urllib.request
import uuid

from audit import (HOST, SPEC, TOKEN, invocation_for, normalise, prepare,
                   reset_server, run)

RESULTS = "results-api.tsv"

# The line --debug prints before a request body, naming what it is sending to
# where. requests.go logs the method beside the URL for this.
SENT = re.compile(r"payload sent to: (\w+) (\S+)")

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
    request = urllib.request.Request(
        url, method=method, data=json.dumps(payload).encode(),
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
    request = captured_request(text)
    return request and request + ((key, captured),)


def probe(binary, entry, command, flag, home):
    """Return the rows for one command-and-flag pair."""
    omitted = capture(binary, entry, command, flag, "omitted", home)
    given = capture(binary, entry, command, flag, "set", home)
    if not given:
        return [f"{command}\t--{flag}\t\t\t\t\t\tnothing to read\tthe run with"
                f" the flag set sent no request"]

    method, url, payload, owned = given
    runs = [owned] + ([omitted[3]] if omitted else [])
    fields = controlled_fields(
        without_fixtures(omitted[2], command, runs) if omitted else None,
        without_fixtures(payload, command, runs))
    if not fields:
        return [f"{command}\t--{flag}\t{method}\t{url}\t\t\t\tno field\tthe flag"
                f" changes no field of the payload"]

    rows = []
    for field in fields:
        control = replay(method, url, freshened(payload))
        answer = replay(method, url, emptied(freshened(payload), field))
        rows.append(f"{command}\t--{flag}\t{method}\t{url}\t{field}"
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
