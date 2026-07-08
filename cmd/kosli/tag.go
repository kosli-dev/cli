package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type tagOptions struct {
	resourceType string
	resourceID   string
	provider     string
	repoID       string
	payload      TagResourcePayload
}

type TagResourcePayload struct {
	SetTags    map[string]string `json:"set_tags"`
	RemoveTags []string          `json:"remove_tags"`
}

const tagShortDesc = `Tag a resource in Kosli with key-value pairs.  `

var validTagResourceTypes = []string{"flow", "flows", "env", "environment", "environments", "control", "controls", "repo", "repos"}

var tagLongDesc = tagShortDesc + fmt.Sprintf(`
use --set to add or update tags, and --unset to remove tags.

Valid resource types are: %s.

Repos are identified by their name. If multiple repos share the same name
across VCS providers, use --provider to disambiguate, or tag the repo
unambiguously by its internal ID with --repo-id (see: kosli get repo).
`, strings.Join(validTagResourceTypes, ", "))

const tagExample = `
# add/update tags to a flow
kosli tag flow yourFlowName \
	--set key1=value1 \
	--set key2=value2 \
	--api-token yourApiToken \
	--org yourOrgName

# tag an environment
kosli tag env yourEnvironmentName \
	--set key1=value1 \
	--set key2=value2 \
	--api-token yourApiToken \
	--org yourOrgName

# add/update tags to an environment
kosli tag env yourEnvironmentName \
	--set key1=value1 \
	--set key2=value2 \
	--api-token yourApiToken \
	--org yourOrgName

# remove tags from an environment
kosli tag env yourEnvironmentName \
	--unset key1=value1 \
	--api-token yourApiToken \
	--org yourOrgName

# tag a control
kosli tag control yourControlIdentifier \
	--set key1=value1 \
	--api-token yourApiToken \
	--org yourOrgName

# tag a repo
kosli tag repo yourOrg/yourRepoName \
	--set key1=value1 \
	--api-token yourApiToken \
	--org yourOrgName

# tag a repo whose name exists across multiple VCS providers
kosli tag repo yourOrg/yourRepoName \
	--provider github \
	--set key1=value1 \
	--api-token yourApiToken \
	--org yourOrgName

# tag a repo by its internal ID (see: kosli get repo)
kosli tag repo --repo-id yourRepoID \
	--set key1=value1 \
	--api-token yourApiToken \
	--org yourOrgName
`

func newTagCmd(out io.Writer) *cobra.Command {
	o := new(tagOptions)
	cmd := &cobra.Command{
		Use:     "tag RESOURCE-TYPE [RESOURCE-ID]",
		Short:   tagShortDesc,
		Long:    tagLongDesc,
		Example: tagExample,
		Args:    cobra.RangeArgs(1, 2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringToStringVarP(&o.payload.SetTags, "set", "s", map[string]string{}, setTagsFlag)
	cmd.Flags().StringSliceVarP(&o.payload.RemoveTags, "unset", "u", []string{}, unsetTagsFlag)
	cmd.Flags().StringVar(&o.provider, "provider", "", "[optional] The VCS provider of the repo (e.g. github, gitlab). Only valid when tagging repos; required when multiple repos share the same name across providers.")
	cmd.Flags().StringVar(&o.repoID, "repo-id", "", "[optional] The repo's internal ID (see: kosli get repo). Only valid when tagging repos; replaces the RESOURCE-ID argument and identifies the repo unambiguously.")

	addDryRunFlag(cmd)

	return cmd
}

func (o *tagOptions) run(args []string) error {
	o.resourceType = args[0]

	err := validateResourceType(o.resourceType)
	if err != nil {
		return err
	}
	isRepo := o.resourceType == "repo" || o.resourceType == "repos"
	if !isRepo {
		if o.provider != "" {
			return fmt.Errorf("--provider is only valid when tagging repos")
		}
		if o.repoID != "" {
			return fmt.Errorf("--repo-id is only valid when tagging repos")
		}
	}
	if o.provider != "" && o.repoID != "" {
		return fmt.Errorf("--provider cannot be combined with --repo-id")
	}

	// the tags endpoint identifies repos by their internal id (repo names are
	// not unique across VCS providers): use --repo-id as-is, or resolve the
	// repo name to it first
	var urlResourceID string
	switch {
	case len(args) == 2 && o.repoID != "":
		return fmt.Errorf("exactly one of the RESOURCE-ID argument or --repo-id must be provided")
	case len(args) == 2:
		o.resourceID = args[1]
		urlResourceID = o.resourceID
		if isRepo && !global.DryRun {
			urlResourceID, err = o.resolveRepoID()
			if err != nil {
				return err
			}
		}
	case o.repoID != "":
		o.resourceID = o.repoID
		urlResourceID = o.repoID
	default:
		return fmt.Errorf("the RESOURCE-ID argument is required unless tagging a repo with --repo-id")
	}

	url, err := url.JoinPath(global.Host, "api/v2/tags", global.Org, o.resourceType, urlResourceID)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPatch,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		var addMsg, removedMsg, msg string
		if len(o.payload.SetTags) > 0 {
			keys := make([]string, 0, len(o.payload.SetTags))
			for key := range o.payload.SetTags {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			keysStr := strings.Join(keys, ", ")
			addMsg = fmt.Sprintf("Tag(s) [%s] added", keysStr)
		}
		if len(o.payload.RemoveTags) > 0 {
			sort.Strings(o.payload.RemoveTags)
			keysStr := strings.Join(o.payload.RemoveTags, ", ")
			removedMsg = fmt.Sprintf("Tag(s) [%s] removed", keysStr)
		}
		if addMsg != "" {
			msg = addMsg
			if removedMsg != "" {
				msg += fmt.Sprintf(", and %s", removedMsg)
			}

		} else {
			msg = removedMsg
		}
		if msg != "" {
			msg += fmt.Sprintf(" for %s '%s'", o.resourceType, o.resourceID)
		} else {
			msg = fmt.Sprintf("No tags were applied for %s '%s'", o.resourceType, o.resourceID)
		}
		logger.Info(msg)
	}
	return err
}

// resolveRepoID resolves the repo name in o.resourceID to the repo's internal
// id via the get repo endpoint, passing --provider through for disambiguation.
func (o *tagOptions) resolveRepoID() (string, error) {
	reqURL, err := url.JoinPath(global.Host, "api/v2/repos", global.Org, o.resourceID)
	if err != nil {
		return "", err
	}
	if o.provider != "" {
		params := url.Values{}
		params.Set("provider", o.provider)
		reqURL += "?" + params.Encode()
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return "", err
	}

	var repo struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(response.Body), &repo); err != nil {
		return "", err
	}
	if repo.ID == "" {
		return "", fmt.Errorf("could not resolve repo %q to an id", o.resourceID)
	}
	return repo.ID, nil
}

func validateResourceType(resourceType string) error {
	match := false
	for _, opt := range validTagResourceTypes {
		if resourceType == opt {
			match = true
			break
		}
	}

	if !match {
		return fmt.Errorf("%s is not a valid resource type. Valid resource types are: %s", resourceType, validTagResourceTypes)
	}
	return nil
}
