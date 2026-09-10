---
name: comment-style
description: >
  Guidance for writing precise, useful code comments in any language — inline comments,
  docstrings, and block comments. Decides when a comment earns its place and what it
  should say: explain the non-obvious why or what, never restate the code, narrate
  history, or point at sibling code. Use when writing or editing comments, adding a
  docstring, deciding whether a comment is needed, reviewing comments in a diff, or
  cleaning up redundant, stale, or over-long comments.
---

# Comment Style

A comment must add information that is not already in the code. If deleting the comment
loses no information, delete it. Default to no comment.

Exception: doc comments a language's conventions require — Go godoc on exported
identifiers, for example — stay even when they restate the signature.

## When to Use This
- Writing or editing inline comments, docstrings, or block comments
- Deciding whether a line or block needs a comment at all
- Reviewing a diff and judging whether its comments earn their place
- Cleaning up comments that restate the code, narrate history, or sprawl too long

## The test (in order)
1. **Information, not narration.** Does the comment state a fact the code can't show on its
   own — an invariant, constraint, non-obvious consequence, unit, or reason? If not, cut it.
2. **The fact, not the story.** State what is true *now*, in as few words as it takes.
   No "previously…", "used to…", "now we…" — history lives in git. A counterfactual is
   fine when it *is* the reason ("without this, retries share one deadline"); not when
   it recounts the edit that introduced the line.
3. **This code, not other code.** Don't describe sibling code, UI, or behavior enforced
   elsewhere ("Mirrors the …") — that goes stale and says nothing about this line.
4. **Why over what.** Prefer explaining *why*; a genuinely non-obvious *what* (surprising
   return value, silent edge case) also qualifies.

## Precise means
One sentence carrying the load-bearing fact. When a comment sprawls, find the single thing
a future reader actually needs and keep only that.

A PR or issue ref as a pointer is fine (`# … (#5765)`) — but it supplements the fact, it
does not replace stating it.

## Examples
## Examples
Bad — resemblance, no information about this code:
`# Mirrors the deletability re-check pending box`

Bad — history and justification burying one fact:
`# Without this the analytics preflight only ran from the client-triggered background`
`# check, so an operator who navigated away before the initiate response landed got a`
`# plan with no analytics check recorded...`

Good — same fact, stated precisely (ref kept as a pointer):
`# Record the analytics preflight here too — the post-initiate auto-run misses it`
`# when the client navigates away before the initiate response lands (#5765).`

Good — non-obvious what:
`// CommitsCount returns -1 when the trail has no commits yet.`
