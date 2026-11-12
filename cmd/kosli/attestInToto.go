package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kosli-dev/cli/internal/digest"
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

// attestationExternalURL labels the link to the attestation's transparency
// log entry.
const attestationExternalURL = "attestation"

// templateArtifactNamePattern is what the API accepts as an artifact's name
// in a flow template.
var templateArtifactNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// intotoTrustRoots supplies the Sigstore trust roots signatures are checked
// against. Tests replace it so that no test fetches one over the network.
var intotoTrustRoots = intoto.SigstoreTrustRoots

const (
	intotoSourceBundle = "bundle"
	intotoSourceOCI    = "oci"
)

var intotoSources = []string{intotoSourceBundle, intotoSourceOCI}

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

The attestation is read from the file given with ^--bundle^: a Sigstore bundle
(^.sigstore.json^), a DSSE envelope, an in-toto statement, or an ^.intoto.jsonl^ file with
one of these per line. Attestations of other kinds are ignored. When several qualify, a
verified one is preferred, then the newest.

With ^--source oci^ the attestation is read instead from the registry holding the image
named by the argument, wherever the tools that build and sign an image attach one to it:

- among the image's OCI referrers, where ^cosign attest^ puts a Sigstore bundle, as does
  ^actions/attest^ with ^push-to-registry^;
- under the tag ^sha256-<digest>.att^, where cosign put attestations before its bundle format;
- in the attestation manifests that ^docker buildx build --provenance --sbom^ adds to the
  image index.

The registry is read with the credentials that fingerprint an image: ^--registry-username^
and ^--registry-password^, or else those Docker or Podman hold on the host. The attestations
buildx makes are unsigned, so recording one needs ^--allow-unsigned^. Each is about one
platform's image, not the image index a tag names, so record it against that image, giving
its digest with ^--fingerprint^.

The statement must list the artifact as its subject. A signed attestation is verified with
Sigstore against a signer you pin: ^--certificate-identity^ (or
^--certificate-identity-regexp^) with ^--certificate-oidc-issuer^, or ^--key^. There is no
way to accept any valid certificate. An unsigned attestation is refused unless
^--allow-unsigned^ is given, and is then recorded as unverified. An attestation that fails
verification, because another signer made it or it was changed after signing, fails the
command and nothing is recorded, even when another attestation would qualify. So does a
signed attestation whose signature cannot be checked, such as when no signer is pinned, a
trust root cannot be fetched or the signer flags do not fit how it was signed; the error
then says it could not be verified, not that it failed.

` + sigstoreTrustRootDesc + `
The record is annotated with ^verified^ and, for a signed attestation, ^signer_identity^,
^signer_issuer^, ^log_index^, ^integrated_time^ and ^signed_timestamp^ where they apply.
Provenance adds ^builder_id^, ^build_type^ and ^workflow^, and an SBOM adds ^sbom_format^
and ^sbom_sha256^. A transparency log entry is linked as the ^attestation^ external URL,
pinned by the SHA-256 of the attestation's bytes.

For provenance, the commit's message, author and time are looked up on GitHub with
^--github-token^ when the statement names a GitHub repository, or else read from the git
checkout at ^--repo-root^ if it holds the commit. Failing both, only the commit's SHA is
recorded, without its details or branch. A repository on GitHub Enterprise Server needs
^--repo-provider github^ and ^--github-base-url^ to be looked up there: the token is only
sent to the server you name, never to a host taken from the statement. A build URL or
commit URL the statement does not carry can be given with ^--build-url^ and ^--commit-url^.
` + fingerprintDesc + kosliIgnoreDesc

const attestInTotoExample = `
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

# record the SLSA provenance cosign attached to an image in its registry, with
# cosign attest --type slsaprovenance1 --predicate provenance.json yourRegistry/yourImage:1.2
kosli attest in-toto yourRegistry/yourImage:1.2 \
	--artifact-type oci \
	--source oci \
	--predicate-type provenance \
	--name app \
	--certificate-identity https://github.com/yourOrg/yourRepo/.github/workflows/build.yml@refs/heads/main \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com \
	--github-token yourGithubToken \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# record the unsigned SBOM docker buildx attached to an image it pushed, with
# docker buildx build --provenance=true --sbom=true --push -t yourRegistry/yourImage:1.2 .
# It is about the platform's image, so give that image's digest, not the tag's:
kosli attest in-toto yourRegistry/yourImage:1.2 \
	--fingerprint yourPlatformImageDigest \
	--source oci \
	--predicate-type sbom \
	--name app.sbom \
	--allow-unsigned \
	--registry-username yourRegistryUsername \
	--registry-password yourRegistryPassword \
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
			if err := o.validate(args); err != nil {
				return err
			}

			if o.source == intotoSourceOCI {
				// The credentials read the image's attestations, whatever
				// supplied its fingerprint.
				return validateRegistryCredentials(cmd, o.fingerprintOptions)
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
	cmd.Flags().StringVar(&o.source, "source", intotoSourceBundle, intotoSourceFlag)
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
func (o *attestInTotoOptions) validate(args []string) error {
	switch o.source {
	case intotoSourceBundle:
		if o.bundlePath == "" {
			return fmt.Errorf("--bundle is required with --source %s", intotoSourceBundle)
		}
	case intotoSourceOCI:
		if err := o.validateOCISource(args); err != nil {
			return err
		}
	default:
		return fmt.Errorf("--source %q is not supported. Must be one of: %s", o.source, strings.Join(intotoSources, ", "))
	}

	kind, err := intoto.ParsePredicateKind(o.predicateType)
	if err != nil {
		return fmt.Errorf("%s for --predicate-type", err)
	}
	o.kind = kind

	if err := o.policy.Validate(); err != nil {
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

func (o *attestInTotoOptions) validateOCISource(args []string) error {
	if o.bundlePath != "" {
		return fmt.Errorf("--bundle is only read with --source %s", intotoSourceBundle)
	}
	if argOrEmpty(args) == "" {
		return fmt.Errorf("--source %s needs the image's name as the argument, to know which repository to read its attestations from", intotoSourceOCI)
	}
	if artifactType := o.fingerprintOptions.artifactType; artifactType != "" && artifactType != "oci" && artifactType != "docker" {
		return fmt.Errorf("--source %s reads the attestations attached to an image, so --artifact-type must be oci or docker, not %s", intotoSourceOCI, artifactType)
	}
	return nil
}

// attestationSource is where --source says to read the attestation from.
func (o *attestInTotoOptions) attestationSource(args []string) (intoto.Source, error) {
	if o.source == intotoSourceOCI {
		repository, err := digest.OpenRepository(args[0], o.fingerprintOptions.registryUsername, o.fingerprintOptions.registryPassword)
		if err != nil {
			return nil, err
		}
		return intoto.OCIImage{Name: args[0], Repository: repository}, nil
	}
	return intoto.BundleFile{Path: o.bundlePath}, nil
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

	source, err := o.attestationSource(args)
	if err != nil {
		return err
	}
	found, err := intoto.Find(source, o.kind, o.fingerprint, intoto.NewVerifier(o.policy, intotoTrustRoots()))
	if err != nil {
		return err
	}
	for _, rejected := range found.Rejected {
		logger.Warn("another %s attestation was not recorded: %s", o.kind, rejected)
	}

	externalURLs, err := processExternalURLs(o.externalURLs, o.externalFingerprints)
	if err != nil {
		return err
	}
	if found.Evidence.LogEntryURL != "" {
		externalURLs[attestationExternalURL] = &URLInfo{Href: found.Evidence.LogEntryURL, Fingerprint: found.Evidence.RawSHA256}
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
		return ghUtils.CommitFinder{Token: o.githubToken, BaseURL: o.githubBaseURL, Debug: global.Debug}
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
		evidence: &requests.FileBytes{Name: o.attachmentName(found), Data: found.Raw},
	})
}

// attachmentName names the attached bytes after the file they came from, and
// after their line when the file holds several attestations. Bytes read from
// a registry are named by their digest, which is the registry's name for them.
func (o *attestInTotoOptions) attachmentName(found *intoto.Found) string {
	if o.source == intotoSourceOCI {
		return fmt.Sprintf("sha256-%s.json", found.Evidence.RawSHA256)
	}
	name := filepath.Base(o.bundlePath)
	if found.Position == 0 {
		return name
	}
	return fmt.Sprintf("%s.%d.json", strings.TrimSuffix(name, filepath.Ext(name)), found.Position)
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
