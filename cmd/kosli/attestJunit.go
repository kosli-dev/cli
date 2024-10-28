package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type JunitAttestationPayload struct {
	*CommonAttestationPayload
	JUnitResults []*JUnitResults `json:"junit_results"`
}

type attestJunitOptions struct {
	*CommonAttestationOptions
	testResultsDir   string
	uploadResultsDir bool
	payload          JunitAttestationPayload
}

const attestJunitShortDesc = `Report a junit attestation to an artifact or a trail in a Kosli flow.
JUnit xml files are read from the ^--results-dir^ directory which defaults to the current directory.
The xml files are automatically uploaded as ^--attachments^ via the ^--upload-results^ flag which defaults to ^true^.  `

const attestJunitLongDesc = attestJunitShortDesc + attestationBindingDesc + `

` + commitDescription

const attestJunitExample = `
# report a junit attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest junit yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--results-dir yourFolderWithJUnitResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a junit attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest junit \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--results-dir yourFolderWithJUnitResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a junit attestation about a trail:
kosli attest junit \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--results-dir yourFolderWithJUnitResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a junit attestation about an artifact which has not been reported yet in a trail:
kosli attest junit \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--results-dir yourFolderWithJUnitResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a junit attestation about a trail with an attachment:
kosli attest junit \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--results-dir yourFolderWithJUnitResults \
	--attachments yourAttachmentPathName \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestJunitCmd(out io.Writer) *cobra.Command {
	o := &attestJunitOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: JunitAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:     "junit [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestJunitShortDesc,
		Long:    attestJunitLongDesc,
		Example: attestJunitExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {

			err := CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			err = RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, false)
			if err != nil {
				return err
			}

			err = ValidateSliceValues(o.redactedCommitInfo, allowedCommitRedactionValues)
			if err != nil {
				return fmt.Errorf("%s for --redact-commit-info", err.Error())
			}

			err = ValidateAttestationArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return ValidateRegistryFlags(cmd, o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", ".", resultsDirFlag)
	cmd.Flags().BoolVar(&o.uploadResultsDir, "upload-results", true, uploadJunitResultsFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestJunitOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/junit", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	o.payload.JUnitResults, err = ingestJunitDir(o.testResultsDir)
	if err != nil {
		return err
	}

	if o.uploadResultsDir {
		// prepare the files to upload as attachments. We are only interested in the actual Junit XMl files
		junitFilenames, err := getJunitFilenames(o.testResultsDir)
		if err != nil {
			return err
		}
		o.attachments = append(o.attachments, junitFilenames...)
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("junit attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
