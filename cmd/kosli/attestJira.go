package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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
	baseURL           string
	username          string
	apiToken          string
	pat               string
	projectKeys       []string
	issueFields       string
	secondarySource   string
	ignoreBranchMatch bool
	assert            bool
	payload           JiraAttestationPayload
}

const attestJiraShortDesc = `Report a jira attestation to an artifact or a trail in a Kosli flow.  `

const attestJiraLongDesc = attestJiraShortDesc + `
Parses the given commit's message, current branch name or the content of the ^--jira-secondary-source^
argument for Jira issue references of the form:  
'at least 2 characters long, starting with an uppercase letter project key followed by
dash and one or more digits'.

If you want to restrict the Jira issue matching to a specific project, use the
^--jira-project-key^ flag to specify your own project key. You can specify multiple project keys if needed.

If the ^--ignore-branch-match^ is set, the branch name is not parsed for a match.

The found issue references will be checked against Jira to confirm their existence.
The attestation is reported in all cases, and its compliance status depends on referencing
existing Jira issues.  
If you have wrong Jira credentials or wrong Jira-base-url it will be reported as non existing Jira issue.
This is because Jira returns same 404 error code in all cases.

The ^--jira-issue-fields^ can be used to include fields from the jira issue. By default no fields
are included. ^*all^ will give all fields. Using ^--jira-issue-fields "*all" --dry-run^ will give you
the complete list so you can select the once you need. The issue fields uses the jira API that is documented here:
https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issues/#api-rest-api-2-issue-issueidorkey-get-request
` + attestationBindingDesc + `

` + commitDescription

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

# report a jira attestation matching a specific jira project key:
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--jira-project-key ABC \
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about a trail and include jira issue summary, description and creator:
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--jira-issue-fields "summary,description,creator"
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about an artifact which has not been reported yet in a trail:
kosli attest jira \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName

# report a jira attestation about a trail with an attachment:
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--attachments yourAttachmentPathName \
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
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert

# get jira reference from original branch name in a GitHub Pull Request merge job
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-secondary-source ${{ github.head_ref }} \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName
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
		// Args:    cobra.MaximumNArgs(1), // See CustomMaximumNArgs() below
		Use:     "jira [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestJiraShortDesc,
		Long:    attestJiraLongDesc,
		Example: attestJiraExample,
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

			err = MuXRequiredFlags(cmd, []string{"jira-pat", "jira-api-token"}, true)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"jira-pat", "jira-username"}, true)
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
	cmd.Flags().StringVar(&o.baseURL, "jira-base-url", "", jiraBaseUrlFlag)
	cmd.Flags().StringVar(&o.username, "jira-username", "", jiraUsernameFlag)
	cmd.Flags().StringVar(&o.apiToken, "jira-api-token", "", jiraAPITokenFlag)
	cmd.Flags().StringVar(&o.pat, "jira-pat", "", jiraPATFlag)
	cmd.Flags().StringSliceVar(&o.projectKeys, "jira-project-key", []string{}, jiraProjectKeyFlag)
	cmd.Flags().StringVar(&o.issueFields, "jira-issue-fields", "", jiraIssueFieldFlag)
	cmd.Flags().StringVar(&o.secondarySource, "jira-secondary-source", "", jiraSecondarySourceFlag)
	cmd.Flags().BoolVar(&o.ignoreBranchMatch, "ignore-branch-match", false, ignoreBranchMatchFlag)
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

	err = o.validateJiraProjectKeys()
	if err != nil {
		return err
	}

	gv, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}
	jiraIssueKeyPattern := jira.MakeJiraIssueKeyPattern(o.projectKeys)

	issueIDs, commitInfo, err := gv.MatchPatternInCommitMessageORBranchName(jiraIssueKeyPattern, o.payload.Commit.Sha1,
		o.secondarySource, o.ignoreBranchMatch)
	if err != nil {
		return err
	}
	logger.Debug("Checked for Jira issue references in Git commit %s on branch %s commit message:\n%s", commitInfo.Sha1, commitInfo.Branch, commitInfo.Message)
	logger.Debug("the following Jira references are found in commit message or branch name: %v", issueIDs)

	issueLog := ""
	issueFoundCount := 0
	for _, issueID := range issueIDs {
		result, err := jc.GetJiraIssueInfo(issueID, o.issueFields)
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
		logger.Info("jira attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}

	if len(issueIDs) == 0 && o.assert {
		return fmt.Errorf("no Jira references are found in commit message or branch name")
	}

	if issueFoundCount != len(issueIDs) && o.assert {
		return fmt.Errorf("missing Jira issues from references found in commit message or branch name%s", issueLog)
	}
	return wrapAttestationError(err)
}

func (o *attestJiraOptions) validateJiraProjectKeys() error {
	// According to Jira documentation https://confluence.atlassian.com/adminjiraserver/changing-the-project-key-format-938847081.html
	// the Jira project key has to start with a capital letter and can then have capital letters numbers and underscore.
	// But Jira itself will accept lower case letters when searching a repository for matching branches and commits
	matchesJiraProjectKeys, err := regexp.Compile("^[A-Za-z][A-Za-z0-9_]{1,9}$")
	if err != nil {
		return err
	}

	invalidKeys := []string{}
	for _, projectKey := range o.projectKeys {
		isValid := matchesJiraProjectKeys.MatchString(projectKey)
		if !isValid {
			invalidKeys = append(invalidKeys, projectKey)
		}
	}
	if len(invalidKeys) > 0 {
		return fmt.Errorf("Invalid Jira project keys: %s", strings.Join(invalidKeys, ", "))
	}
	return nil
}
