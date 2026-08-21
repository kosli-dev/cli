package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

Matching is case-insensitive: ^proj-42^ and ^PROJ-42^ in a commit message are both
recognised and returned as ^PROJ-42^. Any token that matches the Jira key format
(a word boundary, two or more letters/digits starting with a letter, a dash, and one
or more digits) is treated as a candidate, regardless of whether it is an intentional
Jira reference. For example, a commit message ^see note-1 for context, fixes PROJ-42^
will look up both ^NOTE-1^ and ^PROJ-42^ in Jira. If ^NOTE-1^ does not exist, the
attestation will be non-compliant even though ^PROJ-42^ is valid.
Use ^--jira-project-key^ to restrict matching to one or more known project keys and
avoid unintended candidates.

Any candidate match is automatically excluded if every occurrence in the parsed text is
immediately followed by a hyphen and a digit — for example, ^CVE-2026-41284^ is excluded
because ^CVE-2026^ would be followed by ^-4^. This applies across all parsed sources
(commit message, branch name, and secondary source).
Note: if your Jira project key collides with this pattern (e.g. a project key of ^CVE^), an
issue reference that happens to be the prefix of a longer hyphenated number (such as a CVE
identifier) will be filtered out. Use ^--jira-secondary-source^ with a different identifier
format as a workaround.

If you want to restrict the Jira issue matching to a specific project, use the
^--jira-project-key^ flag to specify your own project key. You can specify multiple project keys if needed.

If the ^--ignore-branch-match^ is set, the branch name is not parsed for a match.

The found issue references will be checked against Jira to confirm their existence.
The attestation is reported in all cases, and its compliance status depends on referencing
existing Jira issues.  
If you have wrong Jira credentials or wrong Jira-base-url it will be reported as non existing Jira issue.
This is because Jira returns same 404 error code in all cases. When Jira's response shows that it did not
accept the credentials, a warning naming them is printed and the issue is reported as not confirmed rather
than silently as missing; run with ^--debug^ to see the status Jira returned for each issue.

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
			o.repoURLExplicit = cmd.Flags().Changed("repo-url")
			o.repoNameExplicit = cmd.Flags().Changed("repository")
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
	url, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org, o.flowName, "trail", o.trailName, "jira")
	if err != nil {
		return err
	}

	err = o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
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
	commitInfo, err := gv.GetCommitInfoFromCommitSHA(o.payload.Commit.Sha1, true, []string{})
	if err != nil {
		return err
	}

	// Search commit message, branch name, and secondary source for Jira issue keys,
	// filtering out false positives from multi-segment identifiers like CVE-2026-41284.
	issueIDs := jira.FindJiraIssueKeys(jiraSearchText(commitInfo, o.secondarySource, o.ignoreBranchMatch), o.projectKeys)
	logger.Debug("Checked for Jira issue references in Git commit %s on branch %s commit message:\n%s", commitInfo.Sha1, commitInfo.Branch, commitInfo.Message)
	logger.Debug("the following Jira references are found in commit message or branch name: %v", issueIDs)

	issueLog := ""
	issueFoundCount := 0
	for _, issueID := range issueIDs {
		result, err := jc.GetJiraIssueInfo(issueID, o.issueFields, logger)
		if err != nil {
			return err
		}
		o.payload.JiraResults = append(o.payload.JiraResults, result)
		switch result.LookupStatus {
		case jira.IssueFound:
			issueFoundCount++
		case jira.LookupUnverified:
			logger.Warn("could not confirm Jira issue %s: %s. The attestation is reported with the issue counted as not found.",
				issueID, result.LookupReason)
		}
		issueLog += fmt.Sprintf("\n\t%s: %s", result.IssueID, jiraIssueLogLine(result))
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer func() {
			if err := os.Remove(evidencePath); err != nil {
				logger.Warn("failed to remove evidence file: %v", err)
			}
		}()
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

	if len(issueIDs) == 0 && o.assert && !global.DryRun {
		errString := ""
		if err != nil {
			errString = fmt.Sprintf("%s\nError: ", err.Error())
		}
		err = fmt.Errorf("%sno Jira references are found in commit message or branch name", errString)
	}

	if issueFoundCount != len(issueIDs) && o.assert && !global.DryRun {
		errString := ""
		if err != nil {
			errString = fmt.Sprintf("%s\nError: ", err.Error())
		}
		err = fmt.Errorf("%smissing Jira issues from references found in commit message or branch name%s", errString, issueLog)
	}
	return wrapAttestationError(err)
}

// jiraIssueLogLine describes one lookup for the per-issue list that goes into the log and
// into the --assert error.
//
// A lookup that could not be completed is listed as not confirmed, with the reason, rather
// than as a missing issue: reporting an expired Jira token as "issue not found" is what
// sent users looking for a missing issue. It still counts as not found everywhere else, so
// the attestation, its compliance status and the --assert exit code are unchanged.
func jiraIssueLogLine(result *jira.JiraIssueInfo) string {
	switch result.LookupStatus {
	case jira.IssueFound:
		return "issue found"
	case jira.LookupUnverified:
		return fmt.Sprintf("issue not confirmed: %s", result.LookupReason)
	default:
		return "issue not found"
	}
}

// jiraSearchText joins the texts that are searched for Jira issue keys: the commit
// message, the branch name unless ignoreBranchMatch is set, and the secondary source
// when one is given.
func jiraSearchText(commitInfo *gitview.CommitInfo, secondarySource string, ignoreBranchMatch bool) string {
	searchTexts := []string{commitInfo.Message}
	if !ignoreBranchMatch {
		searchTexts = append(searchTexts, commitInfo.Branch)
	}
	if secondarySource != "" {
		searchTexts = append(searchTexts, secondarySource)
	}
	return strings.Join(searchTexts, "\n")
}

// jiraProjectKeyRegexp is compiled once, so validateJiraProjectKeys does not re-compile a
// constant pattern on every invocation.
//
// According to Jira documentation https://confluence.atlassian.com/adminjiraserver/changing-the-project-key-format-938847081.html
// the Jira project key has to start with a capital letter and can then have capital letters numbers and underscore.
// But Jira itself will accept lower case letters when searching a repository for matching branches and commits.
var jiraProjectKeyRegexp = regexp.MustCompile("^[A-Za-z][A-Za-z0-9_]{1,9}$")

// normaliseJiraProjectKeys trims each project key in place. cobra splits a comma-separated
// list with encoding/csv, which does not trim, so --jira-project-key "ABC, DEF" arrives as
// {"ABC", " DEF"} and the untrimmed fragment fails validation. Normalising means a single
// canonical key reaches both the validation below and jira.FindJiraIssueKeys, rather than
// each trimming separately and having to agree.
func (o *attestJiraOptions) normaliseJiraProjectKeys() {
	for i, projectKey := range o.projectKeys {
		o.projectKeys[i] = strings.TrimSpace(projectKey)
	}
}

// validateJiraProjectKeys normalises the keys first, so that a caller cannot reach the
// validation without it and get the untrimmed keys rejected.
//
// Keys are reported with %q, because a key that is only whitespace normalises to "" and %v
// renders that as nothing at all, leaving an error naming no key.
func (o *attestJiraOptions) validateJiraProjectKeys() error {
	o.normaliseJiraProjectKeys()
	invalidKeys := []string{}
	for _, projectKey := range o.projectKeys {
		isValid := jiraProjectKeyRegexp.MatchString(projectKey)
		if !isValid {
			invalidKeys = append(invalidKeys, projectKey)
		}
	}
	if len(invalidKeys) > 0 {
		return fmt.Errorf("invalid Jira project keys: %q", invalidKeys)
	}
	return nil
}
