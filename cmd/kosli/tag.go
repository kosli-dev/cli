package main

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type tagOptions struct {
	resourceType string
	resourceID   string
	payload      TagResourcePayload
}

type TagResourcePayload struct {
	SetTags    map[string]string `json:"set_tags"`
	RemoveTags []string          `json:"remove_tags"`
}

const tagShortDesc = `Tag a resource in Kosli with key-value pairs.  `

const tagLongDesc = tagShortDesc + `
use --set to add or update tags, and --unset to remove tags.
`

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
`

func newTagCmd(out io.Writer) *cobra.Command {
	o := new(tagOptions)
	cmd := &cobra.Command{
		Use:     "tag RESOURCE-TYPE RESOURCE-ID",
		Short:   tagShortDesc,
		Long:    tagLongDesc,
		Example: tagExample,
		Args:    cobra.ExactArgs(2),
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

	cmd.Flags().StringToStringVar(&o.payload.SetTags, "set", map[string]string{}, setTagsFlag)
	cmd.Flags().StringSliceVar(&o.payload.RemoveTags, "unset", []string{}, unsetTagsFlag)

	addDryRunFlag(cmd)

	return cmd
}

func (o *tagOptions) run(args []string) error {
	o.resourceType = args[0]
	o.resourceID = args[1]

	err := validateResourceType(o.resourceType)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/v2/tags/%s/%s/%s", global.Host, global.Org, o.resourceType, o.resourceID)

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

func validateResourceType(resourceType string) error {
	options := []string{"flow", "flows", "env", "environment", "environments"}
	match := false
	for _, opt := range options {
		if resourceType == opt {
			match = true
			break
		}
	}

	if !match {
		return fmt.Errorf("%s is not a valid resource type. Valid resource types are: %s", resourceType, options)
	}
	return nil
}
