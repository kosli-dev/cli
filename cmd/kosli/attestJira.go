package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/jira"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type JiraAttestationPayload struct {
	*CommonAttestationPayload
	JiraResults []*jira.JiraIssueInfo `json:"jira_results"`
}

type attestJiraOptions struct {
	*CommonAttestationOptions
	baseURL  string
	username string
	apiToken string
	pat      string
	assert   bool
	payload  JiraAttestationPayload
}

const attestJiraShortDesc = `Report a jira attestation to an artifact or a trail in a Kosli flow.  `

const attestJiraLongDesc = attestJiraShortDesc + `
Parses the given commit's message or current branch name for Jira issue references of the 
form:  
'at least 2 characters long, starting with an uppercase letter project key followed by
dash and one or more digits'. 

The found issue references will be checked against Jira to confirm their existence.
The evidence is reported in all cases, and its compliance status depends on referencing
existing Jira issues.  
If you have wrong Jira credentials or wrong Jira-base-url it will be reported as non existing Jira issue.
This is because Jira returns same 404 error code in all cases.
` + fingerprintDesc

const attestJiraExample = `
# report a jira attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest jira yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest jira \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about a trail:
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about an artifact which has not been reported yet in a trail:
kosli attest jira \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about a trail with an evidence file:
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--evidence-paths=yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if no issue reference is found, or the issue is not found in your jira instance
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--evidence-paths=yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert
`

func newAttestJiraCmd(out io.Writer) *cobra.Command {
	o := &attestJiraOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: JiraAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		Use:     "jira [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestJiraShortDesc,
		Long:    attestJiraLongDesc,
		Example: attestJiraExample,
		Args:    cobra.MaximumNArgs(1),
		Hidden:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, false)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"jira-pat", "jira-api-token"}, true)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"jira-pat", "jira-username"}, true)
			if err != nil {
				return err
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
	cmd.Flags().StringVar(&o.baseURL, "jira-base-url", "", jiraBaseUrlFlag)
	cmd.Flags().StringVar(&o.username, "jira-username", "", jiraUsernameFlag)
	cmd.Flags().StringVar(&o.apiToken, "jira-api-token", "", jiraAPITokenFlag)
	cmd.Flags().StringVar(&o.pat, "jira-pat", "", jiraPATFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, attestationAssertFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "commit", "jira-base-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestJiraOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/jira", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	o.baseURL = strings.TrimSuffix(o.baseURL, "/")
	jc := jira.NewJiraConfig(o.baseURL, o.username, o.apiToken, o.pat)

	o.payload.JiraResults = []*jira.JiraIssueInfo{}

	gv, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}
	// Jira issue keys consist of [project-key]-[sequential-number]
	// project key must be at least 2 characters long and start with an uppercase letter
	// more info: https://support.atlassian.com/jira-software-cloud/docs/what-is-an-issue/#Workingwithissues-Projectandissuekeys
	jiraIssueKeyPattern := `[A-Z][A-Z0-9]{1,9}-[0-9]+`

	issueIDs, commitInfo, err := gv.MatchPatternInCommitMessageORBranchName(jiraIssueKeyPattern, o.payload.Commit.Sha1)
	if err != nil {
		return err
	}
	logger.Debug("Checked for Jira issue references in Git commit %s on branch %s commit message:\n%s", commitInfo.Sha1, commitInfo.Branch, commitInfo.Message)
	logger.Debug("the following Jira references are found in commit message or branch name: %v", issueIDs)

	if len(issueIDs) == 0 && o.assert {
		return fmt.Errorf("no Jira references are found in commit message or branch name")
	}

	issueLog := ""
	issueFoundCount := 0
	for _, issueID := range issueIDs {
		result, err := jc.GetJiraIssueInfo(issueID)
		if err != nil {
			return err
		}
		o.payload.JiraResults = append(o.payload.JiraResults, result)
		issueExistLog := "issue not found"
		if result.IssueExists {
			issueExistLog = "issue found"
			issueFoundCount++
		}
		issueLog += fmt.Sprintf("\n\t%s: %s", result.IssueID, issueExistLog)
	}
	if issueFoundCount != len(issueIDs) && o.assert {
		return fmt.Errorf("missing Jira issues from references found in commit message or branch name%s", issueLog)
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.evidencePaths)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("jira attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return err
}
