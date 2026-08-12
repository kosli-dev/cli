package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNormalizeBoolFlagArgsLeavesTrailingBareBoolFlagAlone covers the boundary
// where a boolean flag is the final token, so there is no following token to
// consider joining. The bare form keeps its NoOptDefVal of true.
func TestNormalizeBoolFlagArgsLeavesTrailingBareBoolFlagAlone(t *testing.T) {
	args := []string{
		"fingerprint", "testdata/person-schema.json",
		"--artifact-type", "file",
		"--debug",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, args, normalized)
}

// TestNormalizeBoolFlagArgsLeavesGroupedShorthandsAlone pins the deliberate
// decision not to rewrite a grouped shorthand token, so that "-qC false" keeps
// failing with the arg-count error and its link to the boolean flags FAQ. See
// boolFlagTokens for why. This test exists so that widening the rewrite to
// cover groups is a visible choice rather than a silent one.
func TestNormalizeBoolFlagArgsLeavesGroupedShorthandsAlone(t *testing.T) {
	args := []string{
		"attest", "generic", "testdata/file1",
		"--artifact-type", "file",
		"-qC", "false",
		"--name", "foo",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, args, normalized)
}

// TestNormalizeBoolFlagArgsStopsAtTerminator checks that nothing after the "--"
// terminator is rewritten. pflag stops parsing flags there, so every later
// token is a positional argument, however much it looks like a flag. Rewriting
// past the terminator would alter arguments the flag parser never inspects.
func TestNormalizeBoolFlagArgsStopsAtTerminator(t *testing.T) {
	args := []string{
		"fingerprint", "--artifact-type", "file",
		"--", "--debug", "false",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, args, normalized)
}

// TestNormalizeBoolFlagArgsLeavesUndocumentedBoolLiteralsAlone pins the
// deliberate restriction to the two literals Kosli documents. pflag's bool
// values go through strconv.ParseBool, so the "=" form also accepts 1, 0, t, f,
// TRUE, False and friends, but the space form is rewritten only for "true" and
// "false". Widening this would capture positionals named "0" or "1", which are
// far more plausible artifact names than "true", and would mean guessing intent
// on input no Kosli documentation describes. The undocumented forms therefore
// keep failing loudly instead of being interpreted.
func TestNormalizeBoolFlagArgsLeavesUndocumentedBoolLiteralsAlone(t *testing.T) {
	for _, literal := range []string{"TRUE", "True", "FALSE", "False", "1", "0", "t", "f"} {
		t.Run(literal, func(t *testing.T) {
			args := []string{
				"attest", "generic", "testdata/file1",
				"--artifact-type", "file",
				"--compliant", literal,
				"--name", "foo",
			}
			root, err := newRootCmd(io.Discard, io.Discard, args)
			require.NoError(t, err)

			normalized := normalizeBoolFlagArgs(root, args)

			require.Equal(t, args, normalized)
		})
	}
}

// TestNormalizeBoolFlagArgsCapturesPositionalNamedTrue pins the one accepted
// ambiguity of the rewrite: a positional argument whose literal value is "true"
// or "false" immediately after a bare boolean flag is taken as that flag's
// value. Here the artifact being fingerprinted is a file named "true", and it
// is swallowed by --debug. This is deliberate: Kosli positionals are artifact
// names, fingerprints and file paths, for which such a name is pathological,
// and distinguishing the two cases would mean predicting each command's
// expected argument count.
func TestNormalizeBoolFlagArgsCapturesPositionalNamedTrue(t *testing.T) {
	args := []string{"fingerprint", "--debug", "true", "--artifact-type", "file"}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, []string{"fingerprint", "--debug=true", "--artifact-type", "file"}, normalized)
}

// TestNormalizeBoolFlagArgsCapturesBoolFlagTokenUsedAsFlagValue pins the second
// accepted ambiguity of the rewrite: a boolean flag token given as the value of
// a preceding value-expecting flag is still rewritten. Here pflag would give
// --name the value "--compliant" and leave "false" as a positional argument,
// whereas the rewrite reads the same tokens as a space-form boolean. This is
// deliberate: telling the two apart would mean tracking which tokens pflag
// consumes as values, and a flag value that is itself a flag token is
// pathological.
func TestNormalizeBoolFlagArgsCapturesBoolFlagTokenUsedAsFlagValue(t *testing.T) {
	args := []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--name", "--compliant", "false",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--name", "--compliant=false",
	}, normalized)
}

// TestNormalizeBoolFlagArgsJoinsInheritedBoolFlag checks that a boolean flag
// inherited from a parent command is rewritten too, not just one declared on
// the subcommand itself.
func TestNormalizeBoolFlagArgsJoinsInheritedBoolFlag(t *testing.T) {
	args := []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--debug", "false",
		"--flow", "my-flow",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--debug=false",
		"--flow", "my-flow",
	}, normalized)
}

// TestNormalizeBoolFlagArgsJoinsGlobalBoolFlagBeforeSubcommand checks that a
// global boolean flag written in the space form before the subcommand is
// rewritten too. Left alone, the stray "false" stops cobra resolving `list
// flows`, and the CLI prints the root help text and exits 0, so the command
// silently never runs.
func TestNormalizeBoolFlagArgsJoinsGlobalBoolFlagBeforeSubcommand(t *testing.T) {
	args := []string{"--debug", "false", "list", "flows"}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, []string{"--debug=false", "list", "flows"}, normalized)
}

// TestNormalizeBoolFlagArgsJoinsSpaceSeparatedBoolValue checks that a boolean
// flag written in the space form is rewritten into the "=" form, so that
// `--compliant false` stops leaving "false" behind as a positional argument.
func TestNormalizeBoolFlagArgsJoinsSpaceSeparatedBoolValue(t *testing.T) {
	args := []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--compliant", "false",
		"--flow", "my-flow",
		"--trail", "my-trail",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	normalized := normalizeBoolFlagArgs(root, args)

	require.Equal(t, []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--compliant=false",
		"--flow", "my-flow",
		"--trail", "my-trail",
	}, normalized)
}

// TestRejectEmptyBoolFlagValueNamesTheFlag pins that an empty token following a
// boolean flag is an error naming that flag.
//
// A boolean flag takes a value only in the "--flag=value" form; it never
// consumes the following token. "--compliant false" works in this CLI solely
// because normalizeBoolFlagArgs rewrites it to "--compliant=false" before pflag
// sees it, and that rewrite requires a boolean literal. An empty token is not
// one, so it is left alone: pflag then sets the flag to true from its presence
// and treats the "" as a positional argument. Without this rejection
// `--compliant "$UNSET"` would record a compliant attestation, and
// `--new-compliance-status "$UNSET"` would override an artifact to compliant
// even though that flag is declared false.
func TestRejectEmptyBoolFlagValueNamesTheFlag(t *testing.T) {
	args := []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--compliant", "",
		"--flow", "my-flow",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	err = rejectEmptyBoolFlagValues(root, args)

	require.Error(t, err)
	require.ErrorContains(t, err, "--compliant")
}

// TestRejectEmptyBoolFlagValueInEqualsForm pins that "--flag=" is refused the
// same way as "--flag \"\"". pflag refuses it either way, but with a message
// naming strconv.ParseBool, which describes the parser rather than what the
// user got wrong. Both spellings of an empty boolean value are the same
// mistake, so both get the same message.
func TestRejectEmptyBoolFlagValueInEqualsForm(t *testing.T) {
	args := []string{
		"attest", "generic", "Dockerfile",
		"--artifact-type", "file",
		"--compliant=",
		"--flow", "my-flow",
	}
	root, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	err = rejectEmptyBoolFlagValues(root, args)

	require.Error(t, err)
	require.ErrorContains(t, err, "--compliant")
	require.ErrorContains(t, err, "empty value")
}

// TestRejectEmptyBoolFlagValueCoversShorthands pins that a shorthand is refused
// the same way as its long form. boolFlagTokens emits both, so this holds
// already; the cases are here so that a change narrowing the token set to long
// forms cannot pass unnoticed. -C is the shorthand for --compliant, which
// carries a compliance verdict.
func TestRejectEmptyBoolFlagValueCoversShorthands(t *testing.T) {
	for _, shorthand := range []string{"-C", "-C="} {
		t.Run(shorthand, func(t *testing.T) {
			args := []string{
				"attest", "generic", "Dockerfile",
				"--artifact-type", "file",
				shorthand,
			}
			if shorthand == "-C" {
				args = append(args, "")
			}
			args = append(args, "--flow", "my-flow")

			root, err := newRootCmd(io.Discard, io.Discard, args)
			require.NoError(t, err)

			err = rejectEmptyBoolFlagValues(root, args)

			require.Error(t, err)
			require.ErrorContains(t, err, "-C")
			require.ErrorContains(t, err, "empty value")
		})
	}
}

// TestRejectEmptyBoolFlagValuesAcceptsEveryLegitimateForm draws the boundary of
// the rejection: only an empty token is refused, and every way of writing a
// boolean flag that pflag or normalizeBoolFlagArgs accepts still passes. It
// covers the gate flags by name, since rejecting one of those wrongly would
// break a pipeline rather than a payload.
//
// These cases pass as soon as the rejection exists, so they are here to pin the
// boundary rather than to drive it: without them, tightening the rule later
// could start refusing valid input with nothing to catch it.
func TestRejectEmptyBoolFlagValuesAcceptsEveryLegitimateForm(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "bare bool flag", args: []string{"list", "flows", "--debug"}},
		{name: "equals form true", args: []string{"list", "flows", "--debug=true"}},
		{name: "equals form false", args: []string{"list", "flows", "--debug=false"}},
		{name: "space form", args: []string{"list", "flows", "--debug", "false"}},
		{name: "bare dry-run", args: []string{"list", "flows", "--dry-run"}},
		{name: "dry-run with a value", args: []string{"list", "flows", "--dry-run=false"}},
		{name: "empty value for a non-bool flag", args: []string{"list", "flows", "--org", ""}},
		{name: "empty token after the terminator", args: []string{"fingerprint", "--", "--debug", ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := newRootCmd(io.Discard, io.Discard, tc.args)
			require.NoError(t, err)

			err = rejectEmptyBoolFlagValues(root, normalizeBoolFlagArgs(root, tc.args))

			require.NoError(t, err)
		})
	}
}
