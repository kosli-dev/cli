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
