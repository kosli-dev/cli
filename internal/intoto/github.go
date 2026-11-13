package intoto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	gh "github.com/google/go-github/v93/github"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/klauspost/compress/snappy"
	ghUtils "github.com/kosli-dev/cli/internal/github"
)

// GitHubActionsIssuer is the OIDC issuer Sigstore certificates name for
// workflows running in GitHub Actions on github.com.
const GitHubActionsIssuer = "https://token.actions.githubusercontent.com"

const githubAttestationsPerPage = 100

// bundleRetryWait is a variable so that tests can shorten it.
var bundleRetryWait = 200 * time.Millisecond

const (
	bundleRetries         = 3
	bundleDownloadTimeout = time.Minute
	// maxBundleMiB bounds a downloaded bundle, both as it arrives and once
	// decompressed. Bundles run to kilobytes; the limit only stops a response
	// that is not one from exhausting memory.
	maxBundleMiB = 32
)

// GitHubRepository is the attestations GitHub holds about the artifacts of
// one repository, read through its attestations API. Any workflow in the
// repository can add to them, so one attestation failing verification, or
// being unreadable, does not rule out another.
type GitHubRepository struct {
	Owner string
	Name  string
	// Token reads the repository's attestations. It is sent to github.com, or
	// to BaseURL when that is set, and to no other host.
	Token string
	// BaseURL is a GitHub Enterprise server; empty means github.com.
	BaseURL string
	Debug   bool
	// RunURI is the run-invocation URI of the GitHub Actions run the caller is
	// part of, or empty outside one. An attestation that run made is
	// preferred: one from the same attempt first, then one from any attempt.
	RunURI string
}

func (g GitHubRepository) String() string {
	return fmt.Sprintf("GitHub repository %s/%s", g.Owner, g.Name)
}

// Selection passes over attestations made by signers other than the pinned
// one, and ones that cannot be read, and prefers the attestation made by the
// caller's own run.
func (g GitHubRepository) Selection() Selection {
	return Selection{SkipFailed: true, PreferredRun: g.RunURI}
}

// DefaultSigner is the signer to trust for a caller with no policy of its own:
// a GitHub Actions workflow in the repository itself, on github.com, in the
// run the caller is part of. The identity matches a workflow file on any
// branch, tag or pull request, which anyone who can push a branch controls, so
// it is tied to that run; there is no default outside one. A reusable workflow
// in another repository signs as that repository and needs pinning explicitly.
func (g GitHubRepository) DefaultSigner() (Policy, error) {
	if g.RunURI == "" {
		return Policy{}, fmt.Errorf("%s has no default signer outside a GitHub Actions run to tie it to", g)
	}
	// GitHub treats owner and repository names as case-insensitive.
	return Policy{
		CertificateOIDCIssuer:     GitHubActionsIssuer,
		CertificateIdentityRegexp: `(?i)^https://github\.com/` + regexp.QuoteMeta(g.Owner+"/"+g.Name) + `/\.github/workflows/`,
		RunInvocationURI:          g.RunURI,
	}, nil
}

// githubAttestation is one entry of the attestations API's listing. GitHub
// gives the bundle inline, at a URL to download it from, or both; go-github's
// own type has no field for the URL.
type githubAttestation struct {
	Bundle    json.RawMessage `json:"bundle"`
	BundleURL string          `json:"bundle_url"`
}

// Attestations returns the attestations of kind that GitHub holds in the
// repository for the artifact with SHA-256 digest, without fetching the ones
// of other kinds. When it holds none of kind it returns them all, so that a
// caller can say what GitHub holds instead. One whose bundle cannot be
// downloaded or does not parse is returned Unreadable.
func (g GitHubRepository) Attestations(kind PredicateKind, digest string) ([]*Attestation, error) {
	digest = normaliseDigest(digest)
	ctx := context.Background()
	client, err := ghUtils.NewGithubClientFromToken(ctx, g.Token, g.BaseURL, g.Debug)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("repos/%s/%s/attestations/sha256:%s", url.PathEscape(g.Owner), url.PathEscape(g.Name), digest)
	var listed []githubAttestation
	for _, predicateType := range kind.predicateTypes() {
		ofType, err := listGitHubAttestations(ctx, client, path, predicateType)
		if err != nil {
			return nil, describeListError(err)
		}
		listed = append(listed, ofType...)
	}
	if len(listed) == 0 {
		// GitHub answers 404 both when nothing matches the filter and when the
		// token cannot read the repository, and does not document how the
		// filter matches. The whole listing tells the two apart.
		//
		// Every bundle in it is then downloaded, whatever its kind, so that the
		// error can name the predicate types GitHub holds instead. That cost is
		// a deliberate trade-off: it is paid only here, where no filter found
		// anything, which ends in an error unless a filter missed an attestation.
		listed, err = listGitHubAttestations(ctx, client, path, "")
		if err != nil {
			return nil, describeListError(err)
		}
	}
	if len(listed) == 0 {
		return nil, fmt.Errorf("it holds no attestation for the artifact (sha256:%s)", digest)
	}

	blobStore := newBundleClient()
	bundles := make([]json.RawMessage, len(listed))
	missing := make([]error, len(listed))
	for i, entry := range listed {
		switch {
		case len(entry.Bundle) > 0 && string(entry.Bundle) != "null":
			bundles[i] = entry.Bundle
		case entry.BundleURL != "":
			bundles[i], missing[i] = downloadBundle(ctx, blobStore, entry.BundleURL)
		default:
			missing[i] = errors.New("GitHub lists neither a bundle nor a URL to download one from")
		}
	}

	attestations := parseEach(bundles)
	location := client.BaseURL() + path
	for i, attestation := range attestations {
		if missing[i] != nil {
			// More use than parseEach's own reason, that there is nothing to parse.
			attestation.Unreadable = fmt.Errorf("%s: %w", attestation.Describe(), missing[i])
		}
		attestation.URL = location
	}
	return attestations, nil
}

// listGitHubAttestations lists what GitHub holds at path, through every page.
// A predicateType narrows the listing to the attestations whose predicate
// type starts with it, and is empty to list them all. GitHub answers a filter
// that matches nothing with 404, which is returned here as an empty listing.
func listGitHubAttestations(ctx context.Context, client *gh.Client, path, predicateType string) ([]githubAttestation, error) {
	var listed []githubAttestation
	query := url.Values{"per_page": {fmt.Sprint(githubAttestationsPerPage)}}
	if predicateType != "" {
		query.Set("predicate_type", predicateType)
	}
	for {
		request, err := client.NewRequest(ctx, http.MethodGet, path+"?"+query.Encode(), nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			Attestations []githubAttestation `json:"attestations"`
		}
		response, err := client.Do(request, &page)
		if err != nil {
			var refused *gh.ErrorResponse
			if predicateType != "" && errors.As(err, &refused) && refused.Response != nil && refused.Response.StatusCode == http.StatusNotFound {
				return listed, nil
			}
			return nil, err
		}
		listed = append(listed, page.Attestations...)

		// The listing is paged by cursor. An unchanged cursor would loop forever.
		if response.After == "" || response.After == query.Get("after") {
			return listed, nil
		}
		query.Set("after", response.After)
	}
}

func describeListError(err error) error {
	var refused *gh.ErrorResponse
	if errors.As(err, &refused) && refused.Response != nil {
		switch refused.Response.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("failed to list its attestations: %w. Either GitHub holds no attestation for the artifact, or the token cannot read the repository and its attestations (the attestations: read permission)", err)
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("failed to list its attestations: %w. The token must be able to read the repository and its attestations (the attestations: read permission)", err)
		}
	}
	return fmt.Errorf("failed to list its attestations: %w", err)
}

// newBundleClient returns the client that downloads the bundles GitHub stores
// apart from its API. One is shared by every download of a listing, so that
// they reuse its connections.
func newBundleClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = nil
	client.HTTPClient.Timeout = bundleDownloadTimeout
	client.RetryMax = bundleRetries
	client.RetryWaitMin = bundleRetryWait
	client.RetryWaitMax = bundleRetryWait
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return client
}

// downloadBundle fetches a bundle GitHub stores apart from its API. The URL
// is signed and points at blob storage, so the request carries no token. The
// body is the bundle's JSON, snappy-compressed.
func downloadBundle(ctx context.Context, client *retryablehttp.Client, bundleURL string) ([]byte, error) {
	// The URL's query is its signature, which does not belong in a message.
	shown := bundleURL
	if parsed, err := url.Parse(bundleURL); err == nil {
		parsed.RawQuery = ""
		shown = parsed.String()
	}
	fail := func(format string, args ...any) ([]byte, error) {
		return nil, fmt.Errorf("its bundle could not be downloaded from %s: %s", shown, fmt.Sprintf(format, args...))
	}

	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, bundleURL, nil)
	if err != nil {
		return fail("the URL is not valid")
	}
	response, err := client.Do(request)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fail("%s", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fail("%s", response.Status)
	}
	const maxBytes = maxBundleMiB << 20
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return fail("%s", err)
	}
	if len(body) > maxBytes {
		return fail("it is larger than the %d MiB a bundle may be", maxBundleMiB)
	}

	// snappy.Decode allocates whatever length the body declares.
	size, err := snappy.DecodedLen(body)
	if err != nil {
		return fail("it is not snappy-compressed")
	}
	if size > maxBytes {
		return fail("it would decompress to more than the %d MiB a bundle may be", maxBundleMiB)
	}
	bundle, err := snappy.Decode(nil, body)
	if err != nil {
		return fail("it is not snappy-compressed")
	}
	return bundle, nil
}
