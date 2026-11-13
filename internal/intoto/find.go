package intoto

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Source supplies the attestations held about an artifact. A source with its
// own rules for choosing among them is also a Selector.
type Source interface {
	// Attestations returns every attestation the source holds that may be
	// about the artifact with SHA-256 digest. Find checks each one.
	Attestations(digest string) ([]*Attestation, error)
	// String names the source in messages.
	String() string
}

// BundleFile is a file holding a Sigstore bundle, a DSSE envelope or an
// in-toto statement, or several one per line.
type BundleFile struct {
	Path string
}

// Attestations reads every attestation in the file; the digest selects
// nothing here.
func (f BundleFile) Attestations(string) ([]*Attestation, error) {
	content, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}
	return Read(content)
}

func (f BundleFile) String() string {
	return fmt.Sprintf("bundle file %s", f.Path)
}

// Selection is a source's own rules for choosing among the attestations it
// holds. The zero value suits a source the caller handed over whole, such as a
// file, where every attestation is expected to verify.
type Selection struct {
	// SkipFailed passes over an attestation that fails verification, whose
	// signature cannot be checked, or that the source marked Unreadable, when
	// another one verifies. It suits a store that several signers write to,
	// where an attestation by another workflow says nothing against the one
	// the policy pins.
	SkipFailed bool
	// PreferredRun is the run-invocation URI of the CI run the caller is part
	// of. An attestation whose certificate names it is chosen ahead of a newer
	// one from another run: first one naming the same attempt of the run, then
	// one naming any attempt of it.
	PreferredRun string
}

// Selector is a Source with selection rules of its own. Find applies the zero
// Selection to any other Source.
type Selector interface {
	Source
	Selection() Selection
}

// Found is the attestation chosen to record, with what verifying it
// established.
type Found struct {
	*Attestation
	Evidence *Evidence
	// Rejected holds why other attestations of the same kind about the same
	// artifact were not chosen, such as being unsigned or, from a source that
	// skips them, failing verification, and why any the source holds could not
	// be read, so a caller can report them.
	Rejected []error
}

// Find returns the one attestation of kind about the artifact with SHA-256
// digest that source holds, verified by verifier. When several qualify, a
// verified one is preferred to an unsigned one, then one from the same attempt
// of the source's preferred run, then one from any attempt of that run, then
// the newest by its trusted time, then the one read last.
//
// If any signed one fails verification, or its signature cannot be checked,
// Find returns that and chooses none: a signature that is bad, or that nobody
// could check, is a stronger signal than a missing one, so an unsigned or
// verified alternative is not recorded in its place. An attestation of any
// kind that the source marked Unreadable counts the same way. A source whose
// Selection skips failures relaxes this only when another attestation
// verifies; an unsigned one still does not stand in for one that failed.
func Find(source Source, kind PredicateKind, digest string, verifier *Verifier) (*Found, error) {
	var selection Selection
	if selector, ok := source.(Selector); ok {
		selection = selector.Selection()
	}

	attestations, err := source.Attestations(digest)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	var readable []*Attestation
	var unreadable []error
	for _, attestation := range attestations {
		if attestation.Unreadable != nil {
			unreadable = append(unreadable, attestation.Unreadable)
		} else {
			readable = append(readable, attestation)
		}
	}
	candidates, err := Matching(readable, kind, digest)
	if err != nil {
		if len(unreadable) == 0 {
			return nil, fmt.Errorf("%s: %w", source, err)
		}
		if len(readable) > 0 {
			unreadable = append([]error{err}, unreadable...)
		}
		return nil, fmt.Errorf("%s: %w", source, joinedErrors(unreadable))
	}

	var accepted []*Found
	var rejected []error
	unverifiable := unreadable
	anyVerified := false
	for _, candidate := range candidates {
		evidence, err := verifier.Verify(candidate, digest)
		switch {
		case errors.Is(err, ErrFailedVerification), errors.Is(err, ErrCannotVerify):
			unverifiable = append(unverifiable, err)
		case err != nil:
			rejected = append(rejected, err)
		default:
			accepted = append(accepted, &Found{Attestation: candidate, Evidence: evidence})
			anyVerified = anyVerified || evidence.Verified
		}
	}
	if selection.SkipFailed && anyVerified {
		rejected = append(rejected, unverifiable...)
		unverifiable = nil
	}
	if len(unverifiable) == 1 {
		return nil, fmt.Errorf("%s: %w", source, unverifiable[0])
	}
	if len(unverifiable) > 1 {
		return nil, fmt.Errorf("%s: %d attestations could not be accepted: %w", source, len(unverifiable), joinedErrors(unverifiable))
	}
	if len(accepted) == 0 {
		if len(rejected) == 1 {
			return nil, fmt.Errorf("%s: %w", source, rejected[0])
		}
		return nil, fmt.Errorf("%s: none of the %d %s attestations could be accepted: %w", source, len(rejected), kind, joinedErrors(rejected))
	}

	runPreference := func(e *Evidence) int {
		switch {
		case selection.PreferredRun == "":
			return 0
		case e.RunInvocationURI == selection.PreferredRun:
			return 2
		case withoutAttempt(e.RunInvocationURI) == withoutAttempt(selection.PreferredRun):
			return 1
		}
		return 0
	}
	sort.SliceStable(accepted, func(i, j int) bool {
		a, b := accepted[i].Evidence, accepted[j].Evidence
		if a.Verified != b.Verified {
			return !a.Verified
		}
		if runPreference(a) != runPreference(b) {
			return runPreference(a) < runPreference(b)
		}
		return a.Time().Before(b.Time())
	})
	chosen := accepted[len(accepted)-1]
	chosen.Rejected = rejected
	return chosen, nil
}

// withoutAttempt names the run a GitHub Actions run-invocation URI belongs to,
// whichever attempt of it the URI is for. Re-running only some of a run's jobs
// starts a new attempt, while the others keep what an earlier attempt made.
func withoutAttempt(runURI string) string {
	if i := strings.LastIndex(runURI, "/attempts/"); i >= 0 {
		return runURI[:i]
	}
	return runURI
}

// joinedErrors reports several errors on one line, which reads best on a
// terminal, while errors.Is still finds each of them.
type joinedErrors []error

func (e joinedErrors) Error() string {
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

func (e joinedErrors) Unwrap() []error {
	return e
}
