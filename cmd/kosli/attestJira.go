package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
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
	trailerKey        string
	ignoreBranchMatch bool
	assert            bool
	payload           JiraAttestationPayload
}

const attestJiraShortDesc = `Report a jira attestation to an artifact or a trail in a Kosli flow.  `

const attestJiraLongDesc = attestJiraShortDesc + `
By default, parses the given commit's message, current branch name, or the content of the
^--jira-secondary-source^ argument for Jira issue references.
Use ^--jira-trailer^ to read issue keys exclusively from a named git trailer line instead
(e.g. ^Jira: PROJ-42^); when set, the commit message body and branch name are not scanned.
^--jira-trailer^ and ^--jira-secondary-source^ are mutually exclusive.

Jira issue references have the form:
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
identifier) will be filtered out. Use ^--jira-trailer^ to read issue keys from a dedicated
git trailer line (e.g. ^Jira: CVE-42^), which confines scanning to the trailer value and
removes collisions caused by surrounding commit text; write the issue key alone in the
trailer value, not embedded in a longer hyphenated string (e.g. ^Jira: CVE-2026-41284^
would still be filtered out). Alternatively, use ^--jira-secondary-source^ with a different
identifier format.

If you want to restrict the Jira issue matching to a specific project, use the
^--jira-project-key^ flag to specify your own project key. You can specify multiple project keys if needed.

If the ^--ignore-branch-match^ is set, the branch name is not parsed for a match.
^--ignore-branch-match^ has no effect when ^--jira-trailer^ is set, since the branch is
never scanned in trailer mode.

The found issue references will be checked against Jira to confirm their existence.
The attestation is reported in all cases, and its compliance status depends on referencing
existing Jira issues.

A reachable but wrong base URL still surfaces as a non-existent issue, because Jira answers 404
both for an issue that does not exist and for one you may not view. A base URL that cannot be
reached is reported as not confirmed, with the transport error as the reason. A credential
rejection is likewise detected and reported as not confirmed, with a warning identifying the
credentials. Use ^--debug^ to see the status Jira returned per issue.

The ^--jira-issue-fields^ can be used to include fields from the jira issue. By default no fields
are included. ^*all^ will give all fields. Using ^--jira-issue-fields "*all" --dry-run^ will give you
the complete list so you can select the once you need. The issue fields uses the jira API that is documented here:
https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issues/#api-rest-api-2-issue-issueidorkey-get-request
` + attestationBindingDesc + `

` + kosliIgnoreDesc + `

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

# read the jira issue key exclusively from a git trailer line (e.g. "Jira: PROJ-42")
# bypasses commit message and branch scanning entirely — useful when project keys
# collide with patterns like CVE identifiers
kosli attest jira \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--jira-trailer Jira \
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

			err = MuXRequiredFlags(cmd, []string{"jira-trailer", "jira-secondary-source"}, false)
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("jira-trailer") && strings.TrimRight(strings.TrimSpace(o.trailerKey), ":") == "" {
				return fmt.Errorf("--jira-trailer cannot be empty")
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
	cmd.Flags().StringVar(&o.trailerKey, "jira-trailer", "", jiraTrailerFlag)
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

	// Find Jira issue keys either from a named git trailer or by scanning the
	// commit message, branch name, and secondary source.
	var issueIDs []string
	if o.trailerKey != "" {
		if o.ignoreBranchMatch {
			logger.Warn("--ignore-branch-match has no effect when --jira-trailer is set")
		}
		trailerValues := gitview.GetTrailerValues(commitInfo.Message, o.trailerKey)
		combinedTrailerText := strings.Join(trailerValues, "\n")
		issueIDs = jira.FindJiraIssueKeys(combinedTrailerText, o.projectKeys)
		logger.Debug("Checked for Jira issue references in trailer '%s' of Git commit %s: %v", o.trailerKey, commitInfo.Sha1, trailerValues)
		if len(trailerValues) > 0 && len(issueIDs) == 0 {
			logger.Warn("trailer '%s' was found but contained no valid Jira issue keys: %v", o.trailerKey, trailerValues)
		}
	} else {
		searchTexts := []string{commitInfo.Message}
		if !o.ignoreBranchMatch {
			searchTexts = append(searchTexts, commitInfo.Branch)
		}
		if o.secondarySource != "" {
			searchTexts = append(searchTexts, o.secondarySource)
		}
		combinedText := strings.Join(searchTexts, "\n")
		issueIDs = jira.FindJiraIssueKeys(combinedText, o.projectKeys)
		logger.Debug("Checked for Jira issue references in Git commit %s on branch %s commit message:\n%s", commitInfo.Sha1, commitInfo.Branch, commitInfo.Message)
	}
	logger.Debug("the following Jira references are found: %v", issueIDs)

	issueSource := "commit message or branch name"
	if o.trailerKey != "" {
		issueSource = fmt.Sprintf("trailer '%s'", o.trailerKey)
	}

	issueLog := ""
	issueFoundCount := 0
	unconfirmedIDs := []string{}
	unconfirmedReasons := []string{}
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
			unconfirmedIDs = append(unconfirmedIDs, issueID)
			// deduplicated: every issue in a run reports the same run-level cause
			if !slices.Contains(unconfirmedReasons, result.LookupReason) {
				unconfirmedReasons = append(unconfirmedReasons, result.LookupReason)
			}
		case jira.IssueMissing:
			// counted as neither, and listed as not found below
		}
		issueLog += fmt.Sprintf("\n\t%s: %s", result.IssueID, jiraIssueLogLine(result))
	}
	if len(unconfirmedIDs) > 0 {
		logger.Warn("%s", jiraUnconfirmedWarning(unconfirmedIDs, unconfirmedReasons))
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
		err = fmt.Errorf("%sno Jira references are found in %s", errString, issueSource)
	}

	if issueFoundCount != len(issueIDs) && o.assert && !global.DryRun {
		errString := ""
		if err != nil {
			errString = fmt.Sprintf("%s\nError: ", err.Error())
		}
		// prefixed, so it does not read as another entry in the issue list it follows
		reasonLog := ""
		for _, reason := range unconfirmedReasons {
			reasonLog += fmt.Sprintf("\n\treason: %s", reason)
		}
		err = fmt.Errorf("%s%s from references found in %s%s%s", errString,
			jiraAssertHeadline(len(issueIDs)-issueFoundCount-len(unconfirmedIDs), len(unconfirmedIDs)), issueSource, issueLog, reasonLog)
	}
	return wrapAttestationError(err)
}

// jiraAssertHeadline names what went wrong in the first line of an --assert failure. A
// lookup Jira never answered is not evidence of a missing issue; when every lookup was
// answered the wording is unchanged.
func jiraAssertHeadline(missing, unconfirmed int) string {
	switch {
	case unconfirmed == 0:
		return "missing Jira issues"
	case missing == 0:
		return "unconfirmed Jira issues"
	default:
		return "missing or unconfirmed Jira issues"
	}
}

// jiraUnconfirmedWarning builds the one warning covering every lookup Jira did not answer.
// The cause is run-level - an expired token, an unreachable host - so warning per issue would
// repeat one sentence per issue key.
func jiraUnconfirmedWarning(issueIDs, reasons []string) string {
	return fmt.Sprintf("could not confirm %s in Jira: %s. Unconfirmed issues are counted as not found in the attestation.",
		strings.Join(issueIDs, ", "), strings.Join(reasons, "; "))
}

// jiraIssueLogLine describes one lookup for the per-issue list in the log and the --assert
// error. An unanswered lookup reads as not confirmed but still counts as not found
// everywhere else, so compliance and the --assert exit code are unchanged.
func jiraIssueLogLine(result *jira.JiraIssueInfo) string {
	switch result.LookupStatus {
	case jira.IssueFound:
		return "issue found"
	case jira.LookupUnverified:
		return "issue not confirmed"
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
