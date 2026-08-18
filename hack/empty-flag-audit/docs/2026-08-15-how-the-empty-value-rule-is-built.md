# How the empty-value rule is built

The decision document proposes that an empty value is always an error. This is
about what that costs to build. The short answer is that the rule itself is
small, because it is one wrapper around a flag's value and one walk to apply it.

## TL;DR

- The check is one wrapper, applied to every flag by a walk in `root.go`.
- It refuses what a command is handed, so a default filled in from a CI
  environment variable is untouched and Appendix 1's blind spot is closed.
- Two things shape where it lives: slice-valued flags, which cannot be caught
  after parsing, and config and environment values, which arrive through the
  same `Set`.
- A boolean flag in the space form needs the args-level check as well, because
  pflag never hands it a value to refuse.
- Nine test files name the message, so its wording is worth settling once.
- The clean-up sweep is the large part, and it is deletion, and it can wait.

## Where the rule lives

Two mechanisms, at two layers, because an empty value reaches a flag in two
shapes and only one of them reaches pflag.

**The value wrapper**, `cmd/kosli/nonEmptyValue.go`. The walk in
`cmd/kosli/root.go` wraps every flag's value, and the wrapper refuses an empty
value, or an empty element of a comma-separated one, in `Set`. That is the only
moment an empty element exists: pflag splits a stringSlice with a CSV reader,
which yields nothing for an empty string, so after parsing `--exclude ""` cannot
be told from the flag being absent. Values from `~/.kosli.yml` and `KOSLI_*`
variables arrive through the same `Set`, so one wrapper covers every source.

**The args-level check**, `rejectEmptyBoolFlagValues` in
`cmd/kosli/normalizeBoolFlagArgs.go`, called from `cmd/kosli/main.go`. A boolean
flag in the space form never reaches `Set` with an empty value: pflag reads the
flag's presence as true and leaves the `""` as a positional argument. Nothing at
the value layer can see that, so it is caught on the raw args instead.

Both report in the same words, so which layer spoke is invisible to the user.

## Why it is not simply four lines

**Slice-valued flags cannot be tested with `== ""`.** pflag's
`stringSliceValue.String()` returns `"[" + csv + "]"`, so `--exclude ""`
stringifies to `[]` and never compares equal to `""`. A check reading a flag's
rendered value is structurally unable to catch a slice flag, which is where most
of the gap is: 34 of the 165 flag names are slice-valued, and 25 of them are
refused on none of their commands until this rule. Wrapping the value sidesteps
the rendering, because `Set` is handed the raw text.

**Config and environment values arrive through the same door.** `bindFlags`
applies a value from `~/.kosli.yml` or a `KOSLI_*` variable with
`cmd.Flags().Set(...)` (`cmd/kosli/root.go:641`), and pflag's `Set` sets
`Changed = true` (`pflag/flag.go:509`). The wrapper sits under that `Set`, so
`org: ""` in a config file is refused exactly as it is on the command line: one
rule instead of two.

**Two flags lose a capability.** `update control --description ""` and
`update service-account --description ""` clear a description on purpose,
which the decision document identifies as the whole cost of the proposal.
That cost is accepted: the two commands stop being able to empty a
description, and nothing is added in its place.

## Size

| Piece | Rough size |
|---|---|
| the rule itself | one small type plus a registration walk, or four lines if the existing loop is extended |
| removing the required-only gate | 4 lines, and Appendix 1's CI blind spot goes with it |
| the two deliberate-clear flags | nothing, the capability goes |
| warning instead of erroring, for step 2 | the same call site, returning a warning rather than an error |
| tests for the rule | new, and worth writing per flag category rather than per flag |
| existing test fallout | small: five test files mention the current message, so keeping its wording costs almost nothing |
| step 4, deleting what the rule replaces | the large part, and deletion |

On that last row: there are 42 `RequireFlags` call sites and one
`MarkFlagRequired`, plus 58 places comparing a string against `""`. Those 58 are
not all rejections. Some default a value, some derive one, some format output.
So the sweep is a reading job rather than a delete-by-grep job, which is exactly
why the decision document puts it in the plan as its own step rather than
leaving it to someone's memory.

## What stays out of reach

Nothing here helps an unquoted variable. `--compliant ${NOPE}` is deleted by the
shell before the CLI starts, so there is no token left to refuse.
`rejectEmptyBoolFlagValues` already catches the quoted form. The unquoted one
needs the bare flag spelling to stop being accepted on the flags that carry a
verdict, which is a separate decision, as the decision document says.
