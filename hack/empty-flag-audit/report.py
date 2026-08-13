#!/usr/bin/env python3
"""Turn the two audit passes into the figures the decision document quotes.

Reads results.tsv and results-ci.tsv, prints every number the document states,
and with --write-appendices rewrites the two generated tables in it. Run it
after any audit run, or the document and the results drift apart.
"""

import argparse
import csv
import json
import pathlib
import re
from collections import Counter, defaultdict

HERE = pathlib.Path(__file__).parent
REPO = HERE.parent.parent
DOC = REPO / "docs/handover/2026-08-13-empty-value-decision.md"
CATEGORIES = HERE / "categories.json"

# The seven flags the root command declares. They behave the same wherever they
# appear, so the audit measures them once and they are counted apart here.
GLOBALS = {"api-token", "org", "host", "http-proxy", "max-api-retries",
           "debug", "quiet"}


def load(path):
    """Read a results file.

    Quoting is switched off: the message column contains double quotes, and
    csv's default handling would swallow every row after one of them.
    """
    with open(path) as f:
        reader = csv.reader(f, delimiter="\t", quoting=csv.QUOTE_NONE)
        header = next(reader)
        return [dict(zip(header, row)) for row in reader]


def measured(row):
    """Say whether this combination produced a result at all."""
    return row["vs_omitted"] not in ("not tested", "setup failed")


def refused(row):
    """Say whether the empty value was refused.

    It counts as refused when it failed and another run of the same command
    worked. Without that second condition a command that was broken anyway -
    because the audit could not give it a value it liked - would look as though
    it were refusing something.
    """
    works = row["omitted_exit"] == "0" or row["set_exit"] == "0"
    return row["empty_exit"] != "0" and works


def verdict(counts):
    """Summarise what one flag name does across all its commands."""
    if counts["measured"] == 0:
        return "not measured"
    if counts["refused"] == 0:
        answer = "never"
    elif counts["refused"] == counts["measured"]:
        answer = "always" if counts["downstream"] == 0 \
            else "always, some only by the server"
    else:
        answer = "some commands"
    if counts["ci_refused"] < counts["refused"]:
        answer += ", and not at all in CI"
    elif counts["ci_cli_gave_up"]:
        answer += ", but in CI only the server does"
    return answer


def per_flag(laptop, ci):
    """Gather, for each flag name, what happened across its commands."""
    by_key = {(r["command"], r["flag"]): r for r in ci}
    counts = defaultdict(lambda: defaultdict(int))
    for row in laptop:
        c = counts[row["flag"].lstrip("-")]
        if not measured(row):
            c["skipped"] += 1
            continue
        c["measured"] += 1
        if refused(row):
            c["refused"] += 1
            c["cli" if row["refused_by"] == "cli" else "downstream"] += 1
        other = by_key.get((row["command"], row["flag"]))
        if other and measured(other):
            if refused(other):
                c["ci_refused"] += 1
            if refused(row) and row["refused_by"] == "cli" \
                    and other["refused_by"] != "cli":
                c["ci_cli_gave_up"] += 1
    return counts


def figures(laptop, ci):
    """Print every number the document states."""
    print(f"combinations         {len(laptop)}")
    print(f"commands with flags  {len({r['command'] for r in laptop})}")
    print(f"distinct flag names  {len({r['flag'] for r in laptop})}")
    skipped = [r for r in laptop if r["vs_omitted"] == "not tested"]
    print(f"skipped              {len(skipped)} combinations on "
          f"{len({r['command'] for r in skipped})} commands")

    for name, rows in (("laptop", laptop), ("CI", ci)):
        seen = [r for r in rows if measured(r)]
        ref = [r for r in seen if refused(r)]
        through = [r for r in seen if r["empty_exit"] == "0"]
        stuck = [r for r in seen if r not in ref and r not in through]
        by = Counter(r["refused_by"] for r in ref)
        print(f"\n{name}: measured {len(seen)}")
        print(f"   refused by the CLI              {by['cli']}")
        print(f"   accepted, refused by the server {by['downstream']}")
        print(f"   nothing refuses it              {len(through)}")
        print(f"   nothing can be said             {len(stuck)}")
        print(f"   of those let through: same as omitting "
              f"{sum(1 for r in through if r['vs_omitted'] == 'same')}")

    through = [r for r in laptop if measured(r) and r["empty_exit"] == "0"]
    print("\nlet through, by command:")
    for command, n in Counter(r["command"] for r in through).most_common():
        print(f"  {n:4}  {command}")
    print("\nlet through but NOT the same as omitting the flag:")
    for r in through:
        if r["vs_omitted"] != "same":
            print(f"  {r['command']} {r['flag']}")


def appendices(laptop, ci):
    """Return the markdown for the two generated tables."""
    cats = json.loads(CATEGORIES.read_text())
    counts = per_flag(laptop, ci)

    rows = []
    for name in sorted(counts):
        c = counts[name]
        kind = "global" if name in GLOBALS else cats.get(name, "uncategorised")
        rows.append(f"| `--{name}` | {kind} | {c['measured']} of "
                    f"{c['measured'] + c['skipped']} | {verdict(c)} |")

    column = {"always": "always", "always, some only by the server": "server",
              "some commands": "some", "never": "never",
              "not measured": "unmeasured"}
    grouped = defaultdict(Counter)
    for name, c in counts.items():
        base = verdict(c).split(", and not")[0].split(", but in CI")[0]
        kind = "global" if name in GLOBALS else cats.get(name, "uncategorised")
        grouped[kind][column[base]] += 1
    return rows, grouped


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--write-appendices", action="store_true",
                        help="rewrite the flag-by-flag table in the document")
    args = parser.parse_args()

    laptop, ci = load(HERE / "results.tsv"), load(HERE / "results-ci.tsv")
    figures(laptop, ci)

    rows, grouped = appendices(laptop, ci)
    order = ["identity", "location", "filter", "credentials", "output",
             "switch", "metadata", "global"]
    cols = ["always", "server", "some", "never", "unmeasured"]
    print("\nappendix 2, by kind of flag:")
    print(f"  {'category':14}" + "".join(f"{c:>12}" for c in cols))
    for kind in order:
        c = grouped[kind]
        print(f"  {kind:14}" + "".join(f"{c[k]:>12}" for k in cols)
              + f"{sum(c.values()):>8}")

    if args.write_appendices:
        text = DOC.read_text()
        head = text[:text.index("| `--")]
        DOC.write_text(head + "\n".join(rows) + "\n")
        print(f"\nrewrote the flag-by-flag table in {DOC.name}: {len(rows)} rows")
    else:
        print("\n(--write-appendices rewrites the flag-by-flag table in the document)")


if __name__ == "__main__":
    main()
