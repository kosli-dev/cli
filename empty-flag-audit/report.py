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
DOC = HERE / "docs/2026-08-13-empty-value-decision.md"
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

    It counts as refused when it failed and either another run of the same
    command worked, or the empty run failed differently from both of them. The
    first covers commands that work here. The second covers commands that cannot
    work here at all - a check in the CLI still caught the empty value, and
    without this they would look as though they refused nothing.
    """
    works = row["omitted_exit"] == "0" or row["set_exit"] == "0"
    differs = row["vs_omitted"] == "differs" or row["vs_set"] == "differs"
    return row["empty_exit"] != "0" and (works or differs)


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


def reacted(row):
    """Say whether anything in the CLI responded to the empty value.

    For a command that cannot succeed here this is all that can be seen. If the
    empty run failed differently from the other two, a check in the CLI caught
    it. If all three failed the same way, the empty value passed every check the
    CLI has and died at the service instead.
    """
    return row["empty_exit"] != "0" and (row["vs_omitted"] == "differs"
                                         or row["vs_set"] == "differs")


def figures(laptop, ci, needs):
    """Print every number the document states."""
    print(f"combinations         {len(laptop)}")
    print(f"commands with flags  {len({r['command'] for r in laptop})}")
    print(f"distinct flag names  {len({r['flag'] for r in laptop})}")
    print(f"of those, on commands needing a service we cannot reach: "
          f"{sum(1 for r in laptop if r['command'] in needs)}")

    for name, rows in (("laptop", laptop), ("CI", ci)):
        here = [r for r in rows if r["command"] not in needs]
        away = [r for r in rows if r["command"] in needs]
        by = Counter(r["refused_by"] for r in here if refused(r))
        through = [r for r in here if r["empty_exit"] == "0"]
        print(f"\n{name}, the {len(here)} commands that work here:")
        print(f"   refused by the CLI                {by['cli']}")
        print(f"   accepted, refused by the server   {by['downstream']}")
        print(f"   nothing refuses it                {len(through)}")
        print(f"   of those, same as omitting        "
              f"{sum(1 for r in through if r['vs_omitted'] == 'same')}")
        print(f"{name}, the {len(away)} needing a service:")
        print(f"   refused by the CLI                "
              f"{sum(1 for r in away if reacted(r))}")
        print(f"   no CLI check reacted, so it reached the service   "
              f"{sum(1 for r in away if r['empty_exit'] != '0' and not reacted(r))}")
        print(f"   let through, exit 0               "
              f"{sum(1 for r in away if r['empty_exit'] == '0')}")


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
    spec = json.loads((HERE / "spec.json").read_text())
    needs = {c for c, e in spec.items() if e.get("needs")}
    figures(laptop, ci, needs)

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
        # The table is the last thing in the document, and everything from this
        # heading down is regenerated. Anchoring on the heading keeps prose
        # above it safe: a row of any other table can begin with a flag name.
        text = DOC.read_text()
        anchor = "## Appendix 3: every flag in the CLI"
        if anchor not in text:
            raise SystemExit(f"{DOC.name} has no '{anchor}' heading to write under")
        head = text[:text.index(anchor) + len(anchor)]
        preamble = text[len(head):text.index("| `--", len(head))]
        DOC.write_text(head + preamble + "\n".join(rows) + "\n")
        print(f"\nrewrote the flag-by-flag table in {DOC.name}: {len(rows)} rows")
    else:
        print("\n(--write-appendices rewrites the flag-by-flag table in the document)")


if __name__ == "__main__":
    main()
