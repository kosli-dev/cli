# Wiring reference: registration, flags, lifecycle, tests

This file covers the four cross-cutting steps that apply to every new command, regardless of archetype. Archetype-specific logic lives in the `archetype-*.md` files.

---

## 1. Registration

Add `new<VerbNoun>Cmd(out)` to the parent verb's `AddCommand(...)` block.

Canonical examples to read: `cmd/kosli/create.go` (verb with multiple subcommands), `cmd/kosli/attest.go` (same pattern).

```go
cmd.AddCommand(
    newCreateFlowCmd(out),
    newCreateFooCmd(out),   // add the new command here
)
```

**New verb:** If the verb does not yet exist, create `cmd/kosli/<verb>.go` with a factory following the same shape as `create.go`:

```go
func newFooCmd(out io.Writer) *cobra.Command {
    cmd := &cobra.Command{Use: "foo", Short: fooDesc, Long: fooDesc}
    cmd.AddCommand(newFooBarCmd(out))
    return cmd
}
```

Then add `newFooCmd(out)` to the `AddCommand` block in `cmd/kosli/root.go` (the block that lists `newGetCmd`, `newCreateCmd`, etc.).

---

## 2. Flag constants

Append new flag description constants to `cmd/kosli/root.go` in the large `const (...)` block that holds the flag-description strings (e.g. `apiTokenFlag`, `flowNameFlag`). Use the `[optional]`/`[conditional]`/`[defaulted]` prefix convention:

```go
fooNameFlag    = "[optional] The name of the foo."
fooTimeoutFlag = "[defaulted] Timeout in seconds for the foo operation."
```

- `[optional]` - flag is never required
- `[conditional]` - required only under certain conditions (document in the description)
- `[defaulted]` - always has a meaningful default; mention it in the description
- No prefix - unconditionally required

Reference the existing flag-description `const` block in `root.go` for naming and formatting conventions.

---

## 3. Lifecycle annotations

**Beta** - add to the `cobra.Command` literal:

```go
Annotations: map[string]string{betaCLIAnnotation: ""},
```

`betaCLIAnnotation` is defined in `root.go` as `"betaCLI"`. No import needed.

**Hidden/incubating** - combine `Hidden: true` with `docgen.DocHiddenAnnotation` to suppress from both Cobra help listing and generated docs:

```go
Hidden:      true,
Annotations: map[string]string{docgen.DocHiddenAnnotation: "", betaCLIAnnotation: ""},
```

Import: `"github.com/kosli-dev/cli/internal/docgen"`

Canonical examples: `cmd/kosli/evaluate.go` (beta only, parent verb), `cmd/kosli/attestDecision.go` (hidden + beta, leaf command).

---

## 4. Test skeleton

The dominant test pattern in `cmd/kosli/` is a testify suite. Read `cmd/kosli/getFlow_test.go` for the full shape; the key elements are:

**Suite struct and setup:**

```go
type FooBarCommandTestSuite struct {
    suite.Suite
    defaultKosliArguments string
}

func (suite *FooBarCommandTestSuite) SetupTest() {
    global = &GlobalOpts{
        ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
        Org:      "docs-cmd-test-user",
        Host:     "http://localhost:8001",
    }
    suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
    CreateFlow(suite.flowName, suite.T())   // attest archetype also calls BeginTrail
}
```

The test API token and org are the standard test values used across all suites - copy verbatim from `getFlow_test.go`.

**Two minimal test cases:**

```go
func (suite *FooBarCommandTestSuite) TestFooBarCmd() {
    tests := []cmdTestCase{
        {
            wantError: true,
            name:      "missing required flag fails",
            cmd:       fmt.Sprintf(`foo bar %s`, suite.defaultKosliArguments),
            golden:    "Error: ...\n",
        },
        {
            name: "happy path works",
            cmd:  fmt.Sprintf(`foo bar --some-flag value %s`, suite.defaultKosliArguments),
        },
    }
    runTestCmd(suite.T(), tests)
}
```

**Suite entrypoint:**

```go
func TestFooBarCommandTestSuite(t *testing.T) {
    suite.Run(t, new(FooBarCommandTestSuite))
}
```

Leave `golden: ""` on the happy-path case; capture actual output after the first run against a real server and paste it in. See `cmd/kosli/testHelpers.go` for `runTestCmd`, `CreateFlow`, `BeginTrail`, and other setup helpers.
