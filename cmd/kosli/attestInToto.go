package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	ghUtils "github.com/kosli-dev/cli/internal/github"
	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/intoto"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/spf13/cobra"
)

// Annotations attest in-toto derives from the attestation. A caller cannot
// set them with --annotate.
const (
	builderIDAnnotation       = "builder_id"
	buildTypeAnnotation       = "build_type"
	workflowAnnotation        = "workflow"
	verifiedAnnotation        = "verified"
	signerIdentityAnnotation  = "signer_identity"
	signerIssuerAnnotation    = "signer_issuer"
	logIndexAnnotation        = "log_index"
	integratedTimeAnnotation  = "integrated_time"
	signedTimestampAnnotation = "signed_timestamp"
)

// verificationAnnotationKeys are the annotations recording what verifying an
// attestation established, which --annotate cannot set.
var verificationAnnotationKeys = []string{verifiedAnnotation, signerIdentityAnnotation, signerIssuerAnnotation,
	logIndexAnnotation, integratedTimeAnnotation, signedTimestampAnnotation}

// sigstoreTrustRootDesc says what verifying a signature needs from the runner,
// for the help of each command that verifies one.
const sigstoreTrustRootDesc = `Verifying a certificate-signed attestation needs the trust root of the Sigstore instance
that signed it: Sigstore's public-good instance, or GitHub's own, which signs attestations
for private repositories. It is fetched over TUF on every run that verifies one, from
^https://tuf-repo-cdn.sigstore.dev^ or ^https://tuf-repo.github.com^ respectively, and
cached under ^~/.sigstore^, so the runner needs outbound HTTPS to that host and a writable
^~/.sigstore^ directory. A key-signed attestation needs neither, unless it carries a
transparency log entry or a signed timestamp, which are checked against Sigstore's
public-good trust root. An unsigned attestation needs neither.
`

// attestationExternalURL labels the link to the attestation: see
// attestationLink.
const attestationExternalURL = "attestation"

// templateArtifactNamePattern is what the API accepts as an artifact's name
// in a flow template.
var templateArtifactNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// intotoTrustRoots supplies the Sigstore trust roots signatures are checked
// against. Tests replace it so that no test fetches one over the network.
var intotoTrustRoots = intoto.SigstoreTrustRoots

const (
	intotoSourceBundle = "bundle"
	intotoSourceGitHub = "github"
)

var intotoSources = []string{intotoSourceBundle, intotoSourceGitHub}

// githubAttachmentName names the bundle GitHub returned for an attestation,
// which has no file name of its own.
const githubAttachmentName = "attestation.sigstore.json"

// currentGitHubRun returns the run-invocation URI of the GitHub Actions run
// this command is part of, as Sigstore writes it in the certificate of
// anything this attempt of the run signs, and whether there is such a run. Tests replace it,
// so that they behave the same inside GitHub Actions as outside.
var currentGitHubRun = func() (runURI string, inActions bool) {
	if WhichCI() != github {
		return "", false
	}
	return fmt.Sprintf("%s/%s/actions/runs/%s/attempts/%s", os.Getenv("GITHUB_SERVER_URL"),
		os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID"), os.Getenv("GITHUB_RUN_ATTEMPT")), true
}

// intotoGitHubAPI is the base URL github.com's API is reached at when
// --github-base-url is not given; empty leaves it to the GitHub client. Tests
// replace it so that none reaches github.com.
var intotoGitHubAPI = ""

type attestInTotoOptions struct {
	fingerprintOptions   *fingerprintOptions
	fingerprint          string
	flowName             string
	trailName            string
	name                 string
	source               string
	bundlePath           string
	predicateType        string
	kind                 intoto.PredicateKind
	policy               intoto.Policy
	buildURL             string
	commitURL            string
	repoProvider         string
	repoRoot             string
	githubToken          string
	githubBaseURL        string
	githubOrg            string
	repository           string
	annotations          map[string]string
	externalURLs         map[string]string
	externalFingerprints map[string]string
}

const attestInTotoShortDesc = `Record an in-toto attestation, such as SLSA provenance or an SBOM, that a build already produced.  `

const attestInTotoLongDesc = attestInTotoShortDesc + `
Build systems produce in-toto attestations as they build: ^actions/attest^, ^cosign attest^
and ^docker buildx --provenance --sbom^ among them. This command records one in Kosli,
verifying its signature first, rather than regenerating the same evidence with other
^kosli attest^ commands.

^--predicate-type^ chooses what to record, one kind per command:

- ^provenance^ (SLSA provenance v1 or v0.2) reports the artifact named by ^--name^, as
  ^kosli attest artifact^ does. The commit, branch, repository and build URL come from the
  statement, never from the CI environment, so the build's own commit and run are recorded
  even when this command runs in a different job or workflow.
- ^sbom^ (SPDX or CycloneDX) reports an SBOM attestation named by ^--name^, as
  ^kosli attest sbom^ does, with the attestation attached in place of the SBOM file.

Recording provenance and an SBOM takes two commands.

^--source^ chooses where the attestation is read from:

- ^github^ reads the attestations GitHub holds for the artifact, where ^actions/attest^
  stores them, in the repository given with ^--github-org^ and ^--repository^. It needs
  ^--github-token^, with the ^attestations: read^ permission. It is the default when
  running in GitHub Actions. GitHub holds them by the digest that was attested, so name
  an image by the digest the build pushed, as ^yourImageName@sha256:...^ with
  ^--artifact-type oci^, or give that digest with ^--fingerprint^: a tag may since have
  moved to another image.
- ^bundle^ reads the file given with ^--bundle^: a Sigstore bundle (^.sigstore.json^), a
  DSSE envelope, an in-toto statement, or an ^.intoto.jsonl^ file with one of these per
  line. It is the default elsewhere, and whenever ^--bundle^ is given.

Attestations of other kinds are ignored. When several qualify, a verified one is
preferred, then one made by the same attempt of the GitHub Actions run this command is
part of, then one made by any attempt of that run, then the newest.

The statement must list the artifact as its subject. A signed attestation is verified with
Sigstore against a signer you pin: ^--certificate-identity^ (or
^--certificate-identity-regexp^) with ^--certificate-oidc-issuer^, or ^--key^. There is no
way to accept any valid certificate. With ^--source github^ in GitHub Actions, a signer you
do not pin defaults to a workflow in the repository itself: the issuer
^https://token.actions.githubusercontent.com^ and an identity starting
^https://github.com/<github-org>/<repository>/.github/workflows/^. Pin the signer yourself
when a reusable workflow in another repository did the signing. There is no default
outside GitHub Actions, when ^--github-base-url^ is given, or when the run is on a server
other than github.com, and the command stops before reading anything unless a signer is
pinned. An unsigned attestation is refused unless
^--allow-unsigned^ is given, and is then recorded as unverified.

In a file, an attestation that fails verification, because another signer made it or it
was changed after signing, fails the command and nothing is recorded, even when another
attestation in the file would qualify. So does a signed attestation whose signature cannot
be checked, such as when no signer is pinned, a trust root cannot be fetched or the signer
flags do not fit how it was signed; the error then says it could not be verified, not that
it failed. So does an entry of the file that cannot be read as an attestation. GitHub
holds the attestations of every workflow that attested the artifact, so there such an
attestation, or one that cannot be downloaded or read, is passed over with a warning as
long as another one verifies. When none does, the command fails, naming each one, and
nothing is recorded.

` + sigstoreTrustRootDesc + `
The record is annotated with ^verified^ and, for a signed attestation, ^signer_identity^,
^signer_issuer^, ^log_index^, ^integrated_time^ and ^signed_timestamp^ where they apply.
Provenance adds ^builder_id^, ^build_type^ and ^workflow^, and an SBOM adds ^sbom_format^
and ^sbom_sha256^. The ^attestation^ external URL links the attestation's transparency
log entry or, for one read from GitHub that has none, GitHub's listing of the artifact's
attestations. It is pinned by the SHA-256 of the attestation's bytes.

For provenance, the commit's message, author and time are looked up on GitHub with
^--github-token^ when the statement names a GitHub repository, or else read from the git
checkout at ^--repo-root^ if it holds the commit. Failing both, only the commit's SHA is
recorded, without its details or branch. A repository on GitHub Enterprise Server needs
^--repo-provider github^ and ^--github-base-url^ to be looked up there: the token is only
sent to github.com or the server you name, never to a host taken from the statement or to
the storage GitHub serves a bundle from. A build URL or commit URL the statement does not
carry can be given with ^--build-url^ and ^--commit-url^.
` + fingerprintDesc + kosliIgnoreDesc

const attestInTotoExample = `
# in GitHub Actions, record the SLSA provenance actions/attest stored for an image, as the artifact named app.
# The image is named by the digest the build pushed, such as ghcr.io/yourOrg/yourImage@sha256:0123...:
kosli attest in-toto yourImageName@yourImageDigest \
	--artifact-type oci \
	--predicate-type provenance \
	--name app \
	--github-token yourGithubToken \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# record the SBOM attestation stored for the same image, as app.sbom:
kosli attest in-toto yourImageName@yourImageDigest \
	--artifact-type oci \
	--predicate-type sbom \
	--name app.sbom \
	--github-token yourGithubToken \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# outside GitHub Actions, read a repository's attestations and pin the workflow that signed them.
# Giving the digest as the fingerprint skips reading it from the registry:
kosli attest in-toto \
	--fingerprint yourArtifactFingerprint \
	--source github \
	--github-org yourGithubOrg \
	--repository yourGithubRepo \
	--github-token yourGithubToken \
	--predicate-type provenance \
	--name app \
	--certificate-identity-regexp '^https://github.com/yourGithubOrg/yourGithubRepo/.github/workflows/' \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# record verified SLSA provenance from a Sigstore bundle, as the artifact named app:
kosli attest in-toto app.tar.gz \
	--artifact-type file \
	--bundle app.sigstore.json \
	--predicate-type provenance \
	--name app \
	--certificate-identity https://github.com/yourOrg/yourRepo/.github/workflows/build.yml@refs/heads/main \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com \
	--github-token yourGithubToken \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# record the SBOM attestation from the same build, as app.sbom:
kosli attest in-toto app.tar.gz \
	--artifact-type file \
	--bundle app-sbom.sigstore.json \
	--predicate-type sbom \
	--name app.sbom \
	--certificate-identity-regexp '^https://github.com/yourOrg/yourRepo/' \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# record unsigned provenance written by docker buildx, marked as unverified:
kosli attest in-toto yourImageName \
	--artifact-type docker \
	--bundle provenance.json \
	--predicate-type provenance \
	--name app \
	--allow-unsigned \
	--build-url https://ci.example.com/builds/123 \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestInTotoCmd(out io.Writer) *cobra.Command {
	o := &attestInTotoOptions{fingerprintOptions: &fingerprintOptions{}}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:         "in-toto [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:       attestInTotoShortDesc,
		Long:        attestInTotoLongDesc,
		Example:     attestInTotoExample,
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			err = RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, true)
			if err != nil {
				return err
			}

			err = ValidateAttestationArtifactArg(args, o.fingerprintOptions.artifactType, o.fingerprint)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			// Cobra checks required flags only after PreRunE, and validate
			// would otherwise report an empty --predicate-type as an unknown
			// one rather than a missing one.
			if err := cmd.ValidateRequiredFlags(); err != nil {
				return err
			}
			if err := o.validate(); err != nil {
				return err
			}

			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.fingerprint, "fingerprint", "F", "", attestationFingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.trailName, "trail", "T", "", trailNameFlag)
	cmd.Flags().StringVarP(&o.name, "name", "n", "", intotoNameFlag)
	cmd.Flags().StringVar(&o.source, "source", "", intotoSourceFlag)
	cmd.Flags().StringVar(&o.bundlePath, "bundle", "", intotoBundleFlag)
	cmd.Flags().StringVar(&o.predicateType, "predicate-type", "", intotoPredicateTypeFlag)
	addSignerPolicyFlags(cmd, &o.policy)
	cmd.Flags().BoolVar(&o.policy.AllowUnsigned, "allow-unsigned", false, allowUnsignedFlag)
	cmd.Flags().StringVar(&o.buildURL, "build-url", "", intotoBuildURLFlag)
	cmd.Flags().StringVar(&o.commitURL, "commit-url", "", intotoCommitURLFlag)
	cmd.Flags().StringVar(&o.repoProvider, "repo-provider", "", intotoRepoProviderFlag)
	cmd.Flags().StringVar(&o.repoRoot, "repo-root", ".", intotoRepoRootFlag)
	cmd.Flags().StringVar(&o.githubToken, "github-token", "", intotoGithubTokenFlag)
	cmd.Flags().StringVar(&o.githubBaseURL, "github-base-url", "", intotoGithubBaseURLFlag)
	ci := WhichCI()
	cmd.Flags().StringVar(&o.githubOrg, "github-org", DefaultValue(ci, "org"), intotoGithubOrgFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), intotoRepositoryFlag)
	cmd.Flags().StringToStringVar(&o.annotations, "annotate", map[string]string{}, annotationFlag)
	cmd.Flags().StringToStringVar(&o.externalURLs, "external-url", map[string]string{}, externalURLFlag)
	cmd.Flags().StringToStringVar(&o.externalFingerprints, "external-fingerprint", map[string]string{}, externalFingerprintFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "predicate-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

// addSignerPolicyFlags adds the flags that pin who must have signed an
// attestation. They mirror cosign verify-attestation and gh attestation
// verify, so a policy written for either carries across.
func addSignerPolicyFlags(cmd *cobra.Command, policy *intoto.Policy) {
	cmd.Flags().StringVar(&policy.CertificateIdentity, "certificate-identity", "", certificateIdentityFlag)
	cmd.Flags().StringVar(&policy.CertificateIdentityRegexp, "certificate-identity-regexp", "", certificateIdentityRegexpFlag)
	cmd.Flags().StringVar(&policy.CertificateOIDCIssuer, "certificate-oidc-issuer", "", certificateOIDCIssuerFlag)
	cmd.Flags().StringVar(&policy.KeyPath, "key", "", signerKeyFlag)
}

// validate checks everything that needs nothing from the attestation, so a
// mistake costs neither a fingerprint calculation nor a file read.
func (o *attestInTotoOptions) validate() error {
	if err := o.validateSource(); err != nil {
		return err
	}

	kind, err := intoto.ParsePredicateKind(o.predicateType)
	if err != nil {
		return fmt.Errorf("%s for --predicate-type", err)
	}
	o.kind = kind

	if err := o.policy.Validate(); err != nil {
		return err
	}
	if err := o.pinDefaultSigner(); err != nil {
		return err
	}

	if kind == intoto.KindProvenance && !templateArtifactNamePattern.MatchString(o.name) {
		return fmt.Errorf("--name '%s' is not a valid artifact name. With --predicate-type provenance it names the artifact in the flow template, which can only contain letters, digits, '-' and '_'", o.name)
	}
	if kind == intoto.KindSBOM {
		if _, _, err := parseAttestationNameTemplate(o.name); err != nil {
			return fmt.Errorf("--name '%s' is not a valid attestation name: %s", o.name, err)
		}
	}
	if err := validateRepoFlags("", o.repoProvider, false); err != nil {
		return err
	}

	for _, reserved := range o.reservedAnnotations() {
		if _, taken := o.annotations[reserved]; taken {
			return fmt.Errorf("annotation key '%s' is set by this command from the attestation and cannot be provided with --annotate", reserved)
		}
	}
	if _, taken := o.externalURLs[attestationExternalURL]; taken {
		return fmt.Errorf("external URL label '%s' is set by this command from the attestation and cannot be provided with --external-url", attestationExternalURL)
	}
	return nil
}

// validateSource settles which source the attestation is read from and
// checks the flags that source needs. With no --source, a --bundle means the
// file; otherwise GitHub Actions means GitHub's attestations.
func (o *attestInTotoOptions) validateSource() error {
	if o.source == "" {
		o.source = intotoSourceBundle
		if _, inActions := currentGitHubRun(); inActions && o.bundlePath == "" {
			o.source = intotoSourceGitHub
		}
	}

	switch o.source {
	case intotoSourceBundle:
		if o.bundlePath == "" {
			return fmt.Errorf("--bundle is required with --source %s", intotoSourceBundle)
		}
	case intotoSourceGitHub:
		if o.bundlePath != "" {
			return fmt.Errorf("--bundle cannot be used with --source %s", intotoSourceGitHub)
		}
		if o.githubOrg == "" || o.repository == "" {
			return fmt.Errorf("--github-org and --repository are required with --source %s", intotoSourceGitHub)
		}
		if o.githubToken == "" {
			return fmt.Errorf("--github-token is required with --source %s", intotoSourceGitHub)
		}
	default:
		return fmt.Errorf("--source %q is not supported. Must be one of: %s", o.source, strings.Join(intotoSources, ", "))
	}
	return nil
}

// pinDefaultSigner gives --source github a signer when the caller pinned
// none: in GitHub Actions, a workflow in the repository itself. Everything
// GitHub holds is signed, so with no signer at all nothing could be recorded,
// and saying so here saves fetching attestations to find that out.
func (o *attestInTotoOptions) pinDefaultSigner() error {
	if o.source != intotoSourceGitHub || o.policy.PinsSigner() {
		return nil
	}
	runURI, inActions := currentGitHubRun()
	if !inActions {
		return fmt.Errorf("--source %s needs a signer to verify attestations against. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key. In GitHub Actions the signer defaults to a workflow in the repository itself", intotoSourceGitHub)
	}
	// The default names github.com, which another server's workflows do not
	// sign as, so every attestation would fail on a mismatched identity.
	if o.githubBaseURL != "" {
		return fmt.Errorf("--source %s has no default signer when --github-base-url is given: the default is a workflow on github.com. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key", intotoSourceGitHub)
	}
	if !strings.HasPrefix(runURI, "https://github.com/") {
		return fmt.Errorf("--source %s has no default signer in a GitHub Actions run on a server other than github.com: the default is a workflow on github.com. Name the server with --github-base-url, and pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key", intotoSourceGitHub)
	}
	signer := o.githubRepository().WorkflowSigner()
	o.policy.CertificateOIDCIssuer = signer.CertificateOIDCIssuer
	o.policy.CertificateIdentityRegexp = signer.CertificateIdentityRegexp
	logger.Debug("no signer is pinned, so attestations must be signed by a workflow in the repository: issuer %s, identity matching %s",
		signer.CertificateOIDCIssuer, signer.CertificateIdentityRegexp)
	return nil
}

// attestationSource is where --source says to read the attestation from.
func (o *attestInTotoOptions) attestationSource() intoto.Source {
	if o.source == intotoSourceGitHub {
		return o.githubRepository()
	}
	return intoto.BundleFile{Path: o.bundlePath}
}

func (o *attestInTotoOptions) githubRepository() intoto.GitHubRepository {
	// NewGithubConfig strips the owner from --repository, which GitHub Actions
	// defaults to owner/name.
	config := ghUtils.NewGithubConfig(o.githubToken, o.githubAPI(), o.githubOrg, o.repository, global.Debug)
	runURI, _ := currentGitHubRun()
	return intoto.GitHubRepository{
		Owner:   config.Org,
		Name:    config.Repository,
		Token:   config.Token,
		BaseURL: config.BaseURL,
		Debug:   config.Debug,
		RunURI:  runURI,
	}
}

// githubAPI is the base URL of the GitHub server the token is sent to.
func (o *attestInTotoOptions) githubAPI() string {
	return firstNonEmpty(o.githubBaseURL, intotoGitHubAPI)
}

func (o *attestInTotoOptions) reservedAnnotations() []string {
	reserved := append([]string{}, verificationAnnotationKeys...)
	if o.kind == intoto.KindSBOM {
		return append(reserved, sbomFormatAnnotation, sbomSha256Annotation)
	}
	return append(reserved, builderIDAnnotation, buildTypeAnnotation, workflowAnnotation)
}

func (o *attestInTotoOptions) run(args []string) error {
	if o.fingerprint == "" {
		var err error
		o.fingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return fmt.Errorf("failed to calculate artifact fingerprint: %s", err)
		}
	}

	source := o.attestationSource()
	found, err := intoto.Find(source, o.kind, o.fingerprint, intoto.NewVerifier(o.policy, intotoTrustRoots()))
	if err != nil {
		return err
	}
	for _, rejected := range found.Rejected {
		logger.Warn("another attestation was not recorded: %s", rejected)
	}

	externalURLs, err := processExternalURLs(o.externalURLs, o.externalFingerprints)
	if err != nil {
		return err
	}
	if link := attestationLink(found); link != nil {
		externalURLs[attestationExternalURL] = link
	}
	annotations, err := processAnnotations(o.annotations)
	if err != nil {
		return err
	}
	annotations = withAnnotations(annotations, verificationAnnotations(found.Evidence))

	if o.kind == intoto.KindSBOM {
		return o.recordSBOM(source, found, externalURLs, annotations)
	}
	return o.recordProvenance(args, found, externalURLs, annotations)
}

// attestationLink points a reader at the attestation: its transparency log
// entry, which anyone can read, or failing that where its source holds it.
// The link is pinned by the SHA-256 of the attestation's bytes.
func attestationLink(found *intoto.Found) *URLInfo {
	href := firstNonEmpty(found.Evidence.LogEntryURL, found.URL)
	if href == "" {
		return nil
	}
	return &URLInfo{Href: href, Fingerprint: found.Evidence.RawSHA256}
}

func (o *attestInTotoOptions) recordProvenance(args []string, found *intoto.Found, externalURLs map[string]*URLInfo, annotations map[string]string) error {
	provenance, err := intoto.ProvenanceFrom(found.Attestation, o.fingerprint, o.repoProvider)
	if err != nil {
		return err
	}
	buildURL, commitURL := provenance.BuildURL, provenance.CommitURL
	var missingFields, missingFlags []string
	noRepository := false
	for _, field := range provenance.Missing {
		switch field {
		case intoto.MissingCommit:
			return fmt.Errorf("the provenance names no git commit, and Kosli records an artifact by the commit it was built from")
		case intoto.MissingBuildURL:
			buildURL = o.buildURL
			if buildURL == "" {
				missingFields, missingFlags = append(missingFields, field), append(missingFlags, "--build-url")
			}
		case intoto.MissingCommitURL:
			commitURL = o.commitURL
			if commitURL == "" {
				missingFields, missingFlags = append(missingFields, field), append(missingFlags, "--commit-url")
			}
		case intoto.MissingRepository:
			noRepository = true
		}
	}
	if len(missingFlags) > 0 {
		return fmt.Errorf("the provenance does not say the %s, which Kosli records an artifact with. Pass %s",
			strings.Join(missingFields, " or the "), strings.Join(missingFlags, " and "))
	}
	if noRepository {
		logger.Warn("the provenance does not name the repository it was built from, so no repository is recorded")
	}

	payload := AttestArtifactPayload{
		Fingerprint:   o.fingerprint,
		Filename:      firstNonEmpty(provenance.Filename, argOrEmpty(args), o.fingerprint),
		GitCommit:     provenance.Commit,
		GitCommitInfo: o.commitInfo(provenance),
		BuildUrl:      buildURL,
		CommitUrl:     commitURL,
		RepoUrl:       provenance.Repository.URL,
		Name:          o.name,
		TrailName:     o.trailName,
		ExternalURLs:  externalURLs,
		Annotations: withAnnotations(annotations, nonEmpty(map[string]string{
			builderIDAnnotation: provenance.BuilderID,
			buildTypeAnnotation: provenance.BuildType,
			workflowAnnotation:  provenance.Workflow,
		})),
	}
	if provenance.Repository.URL != "" {
		payload.GitRepoInfo = mergeGitRepoInfo(nil, provenance.Repository.ID, provenance.Repository.Name,
			provenance.Repository.URL, provenance.Repository.Provider, true)
	}

	endpoint, err := url.JoinPath(global.Host, "api/v2/artifacts", global.Org, o.flowName)
	if err != nil {
		return err
	}
	_, err = kosliClient.Do(&requests.RequestParams{
		Method:  http.MethodPost,
		URL:     endpoint,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	})
	if err == nil && !global.DryRun {
		logger.Info("artifact %s was attested with fingerprint: %s", payload.Filename, payload.Fingerprint)
	}
	return err
}

// commitInfo returns the details of the commit the provenance names. They are
// looked up in the repository the statement names, not in the one this
// command runs in, then read from a local checkout. Failing both it returns
// nil, leaving the commit recorded by its SHA alone: POST /artifacts refuses
// commit info that lacks the message, author, timestamp or branch, although
// the API's schema marks them optional.
func (o *attestInTotoOptions) commitInfo(provenance *intoto.Provenance) *gitview.BasicCommitInfo {
	if finder := o.commitFinder(provenance.Repository); finder != nil {
		commit, err := finder.CommitBySHA(provenance.Repository.Name, provenance.Commit)
		if err == nil {
			return &gitview.BasicCommitInfo{
				Sha1:      commit.SHA,
				Message:   commit.Message,
				Author:    commit.Author,
				Timestamp: commit.Timestamp,
				Branch:    provenance.Branch,
				URL:       commit.URL,
			}
		}
		logger.Warn("could not look up commit %s: %s", provenance.Commit, err)
	}

	if gitView, err := gitview.New(o.repoRoot); err == nil {
		if local, err := gitView.GetCommitInfoFromCommitSHA(provenance.Commit, false, nil); err == nil {
			info := local.BasicCommitInfo
			// The checkout's branch is this job's, not necessarily the build's.
			info.Branch = provenance.Branch
			return &info
		}
	}

	logger.Warn("the details and branch of commit %s are not recorded, only its SHA: they could not be looked up, and the git repository at --repo-root %s does not hold the commit", provenance.Commit, o.repoRoot)
	return nil
}

// commitFinder chooses how to look up a commit by the repository's provider.
// GitHub is the only provider implemented; a lookup also needs a token.
func (o *attestInTotoOptions) commitFinder(repo intoto.Repository) types.CommitFinder {
	if repo.Provider == "github" && o.githubToken != "" {
		return ghUtils.CommitFinder{Token: o.githubToken, BaseURL: o.githubAPI(), Debug: global.Debug}
	}
	return nil
}

func (o *attestInTotoOptions) recordSBOM(source intoto.Source, found *intoto.Found, externalURLs map[string]*URLInfo, annotations map[string]string) error {
	document, err := intoto.SBOMDocument(found.Attestation)
	if err != nil {
		return err
	}
	data, err := sbomAttestationData(found.Raw, document)
	if err != nil {
		return fmt.Errorf("failed to parse the SBOM in %s: %s", source, err)
	}

	common := &CommonAttestationPayload{
		ArtifactFingerprint: o.fingerprint,
		ExternalURLs:        externalURLs,
		Annotations:         withAnnotations(annotations, sbomAnnotations(data)),
	}
	artifactName, attestationName, err := parseAttestationNameTemplate(o.name)
	if err != nil {
		return err
	}
	if attestationName != "" {
		common.TargetArtifacts = []string{artifactName}
		common.AttestationName = attestationName
	} else {
		common.AttestationName = artifactName
	}

	return newAttestationSubmitter().submit(attestationSubmission{
		flow:  o.flowName,
		trail: o.trailName,
		// The slug is the attestation family, not the type. The server tells
		// them apart by type_name in the body.
		slug:  "system",
		label: "sbom",
		payload: SbomAttestationPayload{
			CommonAttestationPayload: common,
			TypeName:                 "sbom",
			AttestationData:          data,
		},
		evidence: &requests.FileBytes{Name: o.attachmentName(found.Attestation), Data: found.Raw},
	})
}

// attachmentName names the attached bytes after the file they came from, and
// after their line when the file holds several attestations. Bytes read from
// GitHub get a fixed name.
func (o *attestInTotoOptions) attachmentName(a *intoto.Attestation) string {
	if o.source == intotoSourceGitHub {
		return githubAttachmentName
	}
	name := filepath.Base(o.bundlePath)
	if a.Position == 0 {
		return name
	}
	return fmt.Sprintf("%s.%d.json", strings.TrimSuffix(name, filepath.Ext(name)), a.Position)
}

// verificationAnnotations records what verifying the attestation established
// where a reader sees it on the trail page.
func verificationAnnotations(evidence *intoto.Evidence) map[string]string {
	annotations := map[string]string{
		verifiedAnnotation:       strconv.FormatBool(evidence.Verified),
		signerIdentityAnnotation: evidence.SignerIdentity,
		signerIssuerAnnotation:   evidence.SignerIssuer,
	}
	if evidence.LogIndex != nil {
		annotations[logIndexAnnotation] = strconv.FormatInt(*evidence.LogIndex, 10)
	}
	if !evidence.IntegratedTime.IsZero() {
		annotations[integratedTimeAnnotation] = evidence.IntegratedTime.Format(time.RFC3339)
	}
	if !evidence.SignedTime.IsZero() {
		annotations[signedTimestampAnnotation] = evidence.SignedTime.Format(time.RFC3339)
	}
	return nonEmpty(annotations)
}

func nonEmpty(m map[string]string) map[string]string {
	kept := make(map[string]string, len(m))
	for key, value := range m {
		if value != "" {
			kept[key] = value
		}
	}
	return kept
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func argOrEmpty(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
