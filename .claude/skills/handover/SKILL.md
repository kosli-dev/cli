---
name: 'handover'
description: 'Create and maintain a handover document scoped to a GitHub issue, shared across all branches and PRs that address it. Use when starting work on a new issue, resuming an existing one, making decisions, completing changes, or wrapping up a session. Stores context, plan, and next steps in docs/handover/<issue-number>-<slug>.md. Triggers on "handover", "update handover", "session notes", "handoff document", or when an agent makes a decision or completes a change.'
---

# Handover

Create and maintain a living handover document scoped to a **GitHub issue**, not a branch. The document lives at `docs/handover/<issue-number>-<slug>.md` (filename never changes) and is shared across every branch and PR that addresses the issue. Update it whenever a decision is made, a PR merges, or a session ends.

## When to Use This

- Starting work on a new issue (creates the document)
- Starting a new branch for an issue that already has a handover (reads and updates)
- After making an architectural or implementation decision
- After completing a significant change or set of changes
- Before ending a working session
- When another human or agent is about to take over

---

## Instructions

### 1. Determine the Issue Number

Extract leading digits from the current branch name:
```bash
git branch --show-current
```

For example, branch `5638-control-creation-ui` → issue `5638`. If the branch name has no leading digits, ask the user: "What GitHub issue number does this work address?"

If the user confirms there is no associated issue, proceed with the old branch-based format (`docs/handover/<YYYY-MM-DD>-<branch-name>.md`) and skip the Plan and PRs & Branches sections.

---

### 2. Check for an Existing Handover Document

Use Glob to search for `docs/handover/<issue-number>-*.md`. If multiple files match, pick the one where the number prefix matches exactly (e.g., `5638-` not `56380-`).

- **If it exists**: Read it to understand current state, then proceed to step 4 (Update).
  - After reading, check whether the header has a `Project decisions:` line. If so, read that file into context before proceeding — it contains cross-cutting decisions and constraints.
  - Also check whether the project file's `## Handovers` section already references this handover file. If not, add the missing entry (same back-reference rule as step 3a).
- **If it does not exist**: Proceed to step 3 (Create).

---

### 3. Create a New Handover Document

#### 3a. Gather Metadata

**Ticket URL** — Construct as `https://github.com/kosli-dev/server/issues/<number>`. Confirm with the user if unsure.

**Figma designs** — Ask: "Are there any Figma design links?" If yes, collect them for the `Figma:` header line (one URL per line, same `> **Figma:**` format). If no or skipped, omit the line.

**Project decisions** — List existing project decision files:
```bash
ls docs/projects/ 2>/dev/null || echo "(none)"
```
Ask: "Is this part of a larger project? Should it link to an existing project decisions file, or create a new one?"
- If they name an existing file: add `> **Project decisions:** [docs/projects/<name>.md](../projects/<name>.md)` to the header and read that file into context.
- If they want a new file: create `docs/projects/<name>.md` from the template below, seed it with the user's input, then add the header link and read it into context.
- If no: omit the line.

**Back-reference** — Whenever a project decisions file is linked (new or existing), also add a reference to the new handover file in that project file's `## Handovers` section. If the section doesn't exist, append it before the end of the file:
```markdown
## Handovers

<!-- Links to handover documents for issues within this project. -->

- [#<number>: <issue-title>](../handover/<issue-number>-<slug>.md)
```
If the section already exists, append a new list item. Do not add duplicate entries.

New project decisions file template:
```markdown
# <Project Name> — Decisions and Constraints

> Tracks cross-cutting decisions and constraints for the <Project Name> feature.
> Load this document at the start of any Claude session working on a <project-name>-related branch.
>
> Last updated: <YYYY-MM-DD>

---

## Overview

<!-- Brief description of the project and its goal. -->

---

## Domain Constraints

<!-- Immutable rules, format constraints, uniqueness requirements, permission requirements, etc. -->

---

## Design Decisions

<!-- Cross-cutting decisions that apply to all tickets in this project. -->

---

## Open Questions

<!-- Unresolved questions that affect multiple tickets. -->

---

## Related Tickets

<!-- Links to the root issue and major sub-issues. -->

---

## Handovers

<!-- Links to handover documents for issues within this project. -->
```

**Collaborators** — Collect names of contributors so far:
```bash
git log --format="%an" $(git merge-base HEAD master)..HEAD | sort -u
```
Also include the current AI agent identity (e.g., "Claude (claude-sonnet-4-6)").

#### 3b. Fetch the GitHub Issue

Fetch the issue to populate the Problem Definition section:
```bash
gh issue view <number> --json title,body,labels
```

Use the issue title as the file slug: lowercase it, replace spaces and special characters with hyphens, strip non-alphanumeric/hyphen characters, trim to ≤ 50 chars. Example: "Add controls list API" → `add-controls-list-api`.

File path: `docs/handover/<issue-number>-<slug>.md`

Use the issue title and body to write the **Problem Definition** section. If the issue body is minimal or absent, ask the user to describe the problem.

#### 3c. Write the File

Create the `docs/handover/` directory if needed, then write the file from the template at `assets/handover-template.md` with all sections populated.

---

### 4. Update an Existing Handover Document

Update only the sections that have changed. Preserve all existing content unless superseded.

#### Problem Definition
Describe the problem or goal the issue addresses. Written once; revise only if scope changes. Include:
- What problem is being solved
- Why it matters
- Constraints and acceptance criteria

#### Plan
A prioritised checklist of the work items — PRs, phases, or tasks — needed to fully resolve the issue. This section is maintained by the team and evolves as work progresses:
```markdown
- [ ] Work item 1
- [x] Work item 2 (merged in #NNNN)
- [ ] Work item 3
```
- Check items off as PRs merge, noting the PR number in parentheses.
- Add new items as scope becomes clearer.
- Do not remove items; mark dropped scope as "(deferred)" or "(out of scope)".

When creating a new handover, leave this section as a skeleton for the team to fill in. Do not auto-generate a plan.

#### PRs & Branches
A table tracking every branch/PR that has contributed to this issue. Add the **current branch** to this table on every session start (if not already listed). Add the PR number once it is opened.

```markdown
| Branch | PR | Status |
|--------|----|--------|
| `5638-control-creation-ui` | #5721 | merged |
| `5638-control-detail-page` | — | in progress |
```

Status values: `in progress` → `open` (once PR is opened) → `merged` / `closed`.

#### Decisions Made
Append a bullet for each notable decision. Each bullet: one concise sentence covering what was decided and why. Do not delete old entries. Mark superseded decisions as "(superseded by: ...)".

Write decisions at the **problem-domain level**, not the implementation level. Apply this filter before recording:

> **Would a future collaborator need to know this to understand *why* the codebase is shaped the way it is — or can they just read the diff?**
> If the diff explains it, skip it. Only record decisions where the *reasoning* or *trade-off* would be invisible from the code alone.

Rules:
- **No technical identifiers.** If the bullet contains a CSS class name, function name, file name, or line of code — rewrite or drop it.
- **No "we renamed/moved/extracted" bullets.** Refactors are self-evident from the diff.
- **No bug-fix implementation details.** If the fix reflects a deliberate architectural choice, record the choice, not the diagnosis.
- **Describe the choice, not the change.** Capture the trade-off: why this approach and not another?

When in doubt, leave it out. Five high-signal decisions are more useful than twenty granular ones.

**If a proposed update would modify the wording of an existing decision entry (rather than appending a new one), ask the user to confirm before making the change.**

#### Next Steps
A short, prioritised checklist of substantive work for the **current branch or session**. Replace completed items and add new ones. Focus on implementation tasks, decisions to make, or things to investigate — do not include routine workflow steps such as committing, pushing, or opening PRs.
```markdown
- [ ] Item to do
- [x] Completed item
```

#### Collaborators
Maintained in the document header. Append new names; do not remove existing ones.

---

### 5. Write the File

Before writing, update the `Last updated:` field in the header to today's date (`YYYY-MM-DD`). Write the updated content back to the existing file (preserving its original filename).

---

### 6. Confirm

After writing, briefly state:
- The path of the handover file
- What was added or changed

---

## Closing a Session

When the user says "close handover", "end handover", "stop tracking", or similar: do a final update (steps 4–5), update the PRs & Branches table with the current branch's status, and confirm to the user that the session is closed.

---

## Update Triggers

Agents should proactively update the handover document (without being asked) when:
- A design or implementation decision is made
- A significant code change is completed
- A blocker or open question is identified
- The session is ending (update "Next Steps" to reflect where work left off)

---

## Mid-Task Updates (Background Agent Pattern)

When a significant decision is made mid-task, spawn a background agent to record it immediately without interrupting the current task:

```
Agent tool:
  description: "Update handover with decision"
  run_in_background: true
  prompt: |
    Update the handover document at '<path>'.
    Read the current file. Append this decision to the "Decisions Made" section:
    "<what was decided and why, in one concise sentence>"
    Update the "Last updated" date to today. Preserve all other content exactly.
    Allowed tools: Read, Write, Edit, Bash
```

Use Glob (`docs/handover/<issue-number>-*.md`) to find the file path if needed.

Use this pattern for:
- Architectural or approach decisions made during conversation
- Decisions to defer, skip, or not implement something
- Constraints or trade-offs identified mid-task
- Any decision the git diff alone won't explain

---

## File Location

Handover files live at `docs/handover/<issue-number>-<slug>.md`. The filename is set at creation and never renamed. Multiple branches and PRs for the same issue all update the same file.

## Supporting Files

- `assets/handover-template.md` — blank template used when creating a new handover document