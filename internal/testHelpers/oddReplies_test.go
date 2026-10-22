package testHelpers

import (
	"errors"
	"testing"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

func TestNullVariants_NullsEachValueInTurn(t *testing.T) {
	variants, err := NullVariants([]byte(`{"a":{"b":1,"c":"x"},"d":[{"e":true},{"e":false}],"f":[]}`))
	require.NoError(t, err)

	got := map[string]string{}
	for _, v := range variants {
		got[v.Path] = string(v.Body)
	}
	require.Equal(t, map[string]string{
		".":      `null`,
		"a":      `{"a":null,"d":[{"e":true},{"e":false}],"f":[]}`,
		"a.b":    `{"a":{"b":null,"c":"x"},"d":[{"e":true},{"e":false}],"f":[]}`,
		"a.c":    `{"a":{"b":1,"c":null},"d":[{"e":true},{"e":false}],"f":[]}`,
		"d":      `{"a":{"b":1,"c":"x"},"d":null,"f":[]}`,
		"d[]":    `{"a":{"b":1,"c":"x"},"d":[],"f":[]}`,
		"d[0]":   `{"a":{"b":1,"c":"x"},"d":[null,{"e":false}],"f":[]}`,
		"d[0].e": `{"a":{"b":1,"c":"x"},"d":[{"e":null},{"e":false}],"f":[]}`,
		"d[1]":   `{"a":{"b":1,"c":"x"},"d":[{"e":true},null],"f":[]}`,
		"d[1].e": `{"a":{"b":1,"c":"x"},"d":[{"e":true},{"e":null}],"f":[]}`,
		"f":      `{"a":{"b":1,"c":"x"},"d":[{"e":true},{"e":false}],"f":null}`,
	}, got)
}

func TestNullVariants_LeavesAnAlreadyNullValueAlone(t *testing.T) {
	variants, err := NullVariants([]byte(`{"a":null,"b":1}`))
	require.NoError(t, err)
	paths := []string{}
	for _, v := range variants {
		paths = append(paths, v.Path)
	}
	require.Equal(t, []string{".", "b"}, paths)
}

func TestNullVariants_RejectsInvalidJSON(t *testing.T) {
	_, err := NullVariants([]byte(`{`))
	require.Error(t, err)
}

func TestOutcome(t *testing.T) {
	require.Equal(t, "ok", Outcome(func() error { return nil }))
	require.Equal(t, "error", Outcome(func() error { return errors.New("boom") }))
	require.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference",
		Outcome(func() error {
			var p *struct{ x int }
			return errors.New(string(rune(p.x)))
		}))
}

func TestEvidenceOutcome(t *testing.T) {
	baseline := []*types.PREvidence{{URL: "u", Title: "t", Commits: []types.Commit{{SHA: "a"}, {SHA: "b"}}}}
	fetch := func(e ...*types.PREvidence) func() ([]*types.PREvidence, error) {
		return func() ([]*types.PREvidence, error) { return e, nil }
	}

	require.Equal(t, "ok", EvidenceOutcome(baseline, fetch(baseline[0])))
	require.Equal(t, "ok lost:commits lost:title", EvidenceOutcome(baseline,
		fetch(&types.PREvidence{URL: "u", Commits: []types.Commit{{SHA: "a"}}})))
	require.Equal(t, "ok changed:commits changed:title", EvidenceOutcome(baseline,
		fetch(&types.PREvidence{URL: "u", Title: "x", Commits: []types.Commit{{SHA: "a"}, {SHA: ""}}})))
	require.Equal(t, "ok lost:PullRequests(1->0)", EvidenceOutcome(baseline, fetch()))
	require.Equal(t, "error", EvidenceOutcome(baseline, func() ([]*types.PREvidence, error) {
		return nil, errors.New("boom")
	}))
}

func TestNullVariants_NullsAndEmptiesARootList(t *testing.T) {
	variants, err := NullVariants([]byte(`[{"a":1}]`))
	require.NoError(t, err)
	got := map[string]string{}
	for _, v := range variants {
		got[v.Path] = string(v.Body)
	}
	require.Equal(t, map[string]string{
		".":     `null`,
		".[]":   `[]`,
		"[0]":   `[null]`,
		"[0].a": `[{"a":null}]`,
	}, got)
}

func TestNullVariants_RejectsTrailingData(t *testing.T) {
	_, err := NullVariants([]byte(`{"a":1} garbage`))
	require.Error(t, err)
}

func TestEvidenceOutcome_ReportsANilEntryWithoutPanicking(t *testing.T) {
	baseline := []*types.PREvidence{{URL: "u"}}
	require.Equal(t, "ok lost:PullRequest0", EvidenceOutcome(baseline, func() ([]*types.PREvidence, error) {
		return []*types.PREvidence{nil}, nil
	}))
}

func TestEvidenceOutcome_ComparesWhatTheAPIReceives(t *testing.T) {
	baseline := []*types.PREvidence{{URL: "u", Reviews: []types.PRApprovals{}}}

	// Commits are always sent as a list, so nil and empty are the same payload.
	require.Equal(t, "ok", EvidenceOutcome([]*types.PREvidence{{URL: "u", Commits: []types.Commit{}}},
		func() ([]*types.PREvidence, error) { return []*types.PREvidence{{URL: "u"}}, nil }))
	// No reviews key means reviews were not collected, unlike an empty list.
	require.Equal(t, "ok lost:reviews", EvidenceOutcome(baseline,
		func() ([]*types.PREvidence, error) { return []*types.PREvidence{{URL: "u"}}, nil }))
	// A key only the result has is still compared.
	require.Equal(t, "ok changed:reviews changed:title", EvidenceOutcome([]*types.PREvidence{{URL: "u"}},
		func() ([]*types.PREvidence, error) {
			return []*types.PREvidence{{URL: "u", Title: "t", Reviews: []types.PRApprovals{}}}, nil
		}))
	// Empty in both is not a loss.
	require.Equal(t, "ok changed:approvers", EvidenceOutcome([]*types.PREvidence{{URL: "u"}},
		func() ([]*types.PREvidence, error) { return []*types.PREvidence{{URL: "u", Approvers: []any{}}}, nil }))
}

func TestEvidenceOutcome_NilInBothIsNotALoss(t *testing.T) {
	require.Equal(t, "ok", EvidenceOutcome([]*types.PREvidence{nil}, func() ([]*types.PREvidence, error) {
		return []*types.PREvidence{nil}, nil
	}))
}
