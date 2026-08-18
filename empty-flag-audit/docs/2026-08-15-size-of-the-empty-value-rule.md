# How much code does refusing an empty flag value need?

The decision document proposes that an empty value is always an error. This is
about what that costs to build. The short answer is that the rule itself is
small, because three partial versions of it are already in the CLI and the work
is joining them up rather than inventing anything.

## TL;DR

- The check already exists, in `root.go`, applied to required flags only.
- Removing that restriction is about four lines, and it also closes the CI blind
  spot in Appendix 1 for free.
- Three things stop it being a four-line change: slice-valued flags, the
  ordering against config and environment values, and the two flags that clear a
  description on purpose.
- There are three places the rule could live. The one to prefer is the one
  #1092 already chose for two flags, generalised to all of them.
- Test fallout is small: five test files mention the current message.
- The clean-up sweep is the large part, and it is deletion, and it can wait.

## Three partial versions already exist

None of this starts from nothing.

**The post-parse check**, `cmd/kosli/root.go:408-418`. The root command's
`PersistentPreRunE` already walks every flag and refuses an empty value:

```go
if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
    if f.Changed && f.Value.String() == "" {
        flagError = fmt.Errorf("flag '--%s' is required, but empty string was provided", f.Name)
    }
}
```

`f.Changed` is exactly "provided, not defaulted". The predicate wanted is
already written. It is the outer `if` that limits it to flags marked required.

That gate is also the mechanism behind Appendix 1: `RequireFlags` marks a flag
required only when its default is empty, and CI environment variables fill in
some defaults, so inside CI those flags are not marked and the check stops
running. Removing the gate removes that blind spot as a side effect.

**The value-level check**, `cmd/kosli/nonEmptyStringSlice.go`, added by #1092.
A pflag value type that refuses an empty element in `Set`, `Append` and
`Replace`, used on `--attachments` (`cmd/kosli/flags.go:113`) and `--template`
(`cmd/kosli/createFlow.go:90`). Its comments record why `Set` is the place: pflag
splits a slice flag with a CSV reader, which yields nothing for an empty string,
so after parsing `--attachments ""` cannot be told from the flag being absent.

**The args-level check**, `rejectEmptyBoolFlagValues` in
`cmd/kosli/normalizeBoolFlagArgs.go:24`, called from `cmd/kosli/main.go:144`. It
reads the raw argv before parsing and refuses an empty token after a boolean
flag, with the message the decision document quotes.

So the CLI already refuses empty values at three different layers, each covering
a different subset. The work is picking one layer and covering everything.

## Why it is not simply four lines

**Slice-valued flags cannot be tested with `== ""`.** pflag's
`stringSliceValue.String()` returns `"[" + csv + "]"`, so `--exclude ""`
stringifies to `[]` and never compares equal to `""`. The post-parse check as
written is structurally unable to catch a slice flag. That matters because slice
flags are where most of the gap is. 34 of the 165 flag names are slice-valued,
and 25 of them are refused on none of their commands. Six are refused everywhere,
and two of those six are `--attachments` and `--template`, the two
`nonEmptyStringSlice` already covers.

The fix is not a new type. `nonEmptyStringSlice` implements `pflag.SliceValue`,
and its `GetSlice` docstring already says it exists so that "a post-parse check
[can] see the elements rather than the bracketed String form". A central check
asserts to `pflag.SliceValue` and tests the elements.

**Config and environment values arrive through the same door.** `bindFlags`
applies a value from `~/.kosli.yml` or a `KOSLI_*` variable with
`cmd.Flags().Set(...)` (`cmd/kosli/root.go:641`), and pflag's `Set` sets
`Changed = true` (`pflag/flag.go:509`). `initialize()` runs before the check in
`PersistentPreRunE`, so by the time the check runs, a value from a config file
is indistinguishable from one typed on the command line.

That is a decision, not a bug: either `org: ""` in a config file is an error too,
or the check has to run before `initialize()`. #1092 already decided it one way
for its two flags, and wrote the reasoning down: config and environment values
"reach the refusal by the same path as the command line". Following that keeps
one rule instead of two.

**Two flags lose a capability.** `update control --description ""` and
`update service-account --description ""` clear a description on purpose,
which the decision document identifies as the whole cost of the proposal.
That cost is accepted: the two commands stop being able to empty a
description, and nothing is added in its place.

## Where the rule could live

| Layer | How | Strength | Weakness |
|---|---|---|---|
| Post-parse, in `PersistentPreRunE` | extend the existing loop at `root.go:408` | the loop and the predicate exist; one place to read | needs type-aware emptiness for slices; runs after `initialize()`, so config and environment values are already indistinguishable |
| Value wrapping, at registration | generalise `nonEmptyStringSlice` into a decorator wrapping every flag's `Value` | type-agnostic, so slices need no special case; catches every `Set`, whoever calls it; follows a pattern already reviewed and shipped | a walk over the command tree to wrap |
| Pre-parse, on raw argv | extend `rejectEmptyBoolFlagValues` to all flags | sees the raw token, so slice splitting never happens; command-line only, by construction | duplicates pflag's parsing rules, which the comments in that file show is fiddly: `=` and space forms, shorthands, grouping, the `--` terminator |

Value wrapping is the one to prefer. It is the pattern the CLI already uses for
this exact problem, it needs no special case for slices, and it puts the refusal
where the value is set rather than where it is later inspected. The post-parse
loop is worth keeping as a backstop for anything the wrapping cannot reach.

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
