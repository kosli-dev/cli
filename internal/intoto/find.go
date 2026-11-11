package intoto

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Source supplies the attestations held about an artifact. Further sources,
// such as GitHub's attestations API or an OCI registry, implement it too.
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

// Found is the attestation chosen to record, with what verifying it
// established.
type Found struct {
	*Attestation
	Evidence *Evidence
	// Rejected holds why other attestations of the same kind about the same
	// artifact were not chosen, such as being unsigned, so a caller can report
	// them.
	Rejected []error
}

// Find returns the one attestation of kind about the artifact with SHA-256
// digest that source holds, verified by verifier. When several qualify, a
// verified one is preferred to an unsigned one, then the newest by its
// trusted time, then the one read last. If any fails verification, Find
// returns that and chooses none: a bad signature is a stronger signal than a
// missing one, so an unsigned or verified alternative is not recorded in its
// place.
func Find(source Source, kind PredicateKind, digest string, verifier *Verifier) (*Found, error) {
	attestations, err := source.Attestations(digest)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	candidates, err := Matching(attestations, kind, digest)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}

	var accepted []*Found
	var rejected, failed []error
	for _, candidate := range candidates {
		evidence, err := verifier.Verify(candidate, digest)
		switch {
		case errors.Is(err, ErrFailedVerification):
			failed = append(failed, err)
		case err != nil:
			rejected = append(rejected, err)
		default:
			accepted = append(accepted, &Found{Attestation: candidate, Evidence: evidence})
		}
	}
	if len(failed) == 1 {
		return nil, fmt.Errorf("%s: %w", source, failed[0])
	}
	if len(failed) > 1 {
		return nil, fmt.Errorf("%s: %d %s attestations %w: %w", source, len(failed), kind, ErrFailedVerification, joinErrors(failed))
	}
	if len(accepted) == 0 {
		if len(rejected) == 1 {
			return nil, fmt.Errorf("%s: %w", source, rejected[0])
		}
		return nil, fmt.Errorf("%s: none of the %d %s attestations could be accepted: %w", source, len(rejected), kind, joinErrors(rejected))
	}

	sort.SliceStable(accepted, func(i, j int) bool {
		a, b := accepted[i].Evidence, accepted[j].Evidence
		if a.Verified != b.Verified {
			return !a.Verified
		}
		return a.Time().Before(b.Time())
	})
	chosen := accepted[len(accepted)-1]
	chosen.Rejected = rejected
	return chosen, nil
}

func joinErrors(errs []error) error {
	messages := make([]string, len(errs))
	for i, err := range errs {
		messages[i] = err.Error()
	}
	return errors.New(strings.Join(messages, "; "))
}
