package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const dismissProductUpdatesShortDesc = `Dismiss Kosli product updates, marking them as seen.`

const dismissProductUpdatesLongDesc = dismissProductUpdatesShortDesc + `
A dismissed product update is no longer unread for you. Give the ids of the updates to
dismiss, as listed by ^kosli list product-updates^, or use ^--all^ to dismiss every unread update.
An id of an update the org cannot see is ignored.

Product updates are tracked per user, so the API token must belong to a member of the org.`

const dismissProductUpdatesExample = `
# dismiss one product update:
kosli dismiss product-updates yourProductUpdateID \
	--api-token yourAPIToken \
	--org yourOrgName

# dismiss several product updates:
kosli dismiss product-updates yourProductUpdateID1 yourProductUpdateID2 \
	--api-token yourAPIToken \
	--org yourOrgName

# dismiss every unread product update:
kosli dismiss product-updates \
	--all \
	--api-token yourAPIToken \
	--org yourOrgName
`

// The most ids the server accepts in one request.
const maxProductUpdateIDsPerRequest = 200

type productUpdatesSeenPayload struct {
	IDs []string `json:"ids"`
}

type dismissProductUpdatesOptions struct {
	all bool
}

func newDismissProductUpdatesCmd(out io.Writer) *cobra.Command {
	o := new(dismissProductUpdatesOptions)
	cmd := &cobra.Command{
		Use:         "product-updates {PRODUCT-UPDATE-ID [PRODUCT-UPDATE-ID...] | --all}",
		Aliases:     []string{"pu", "pr-up"},
		Short:       dismissProductUpdatesShortDesc,
		Long:        dismissProductUpdatesLongDesc,
		Example:     dismissProductUpdatesExample,
		Annotations: map[string]string{betaCLIAnnotation: ""},
		Args: func(cmd *cobra.Command, args []string) error {
			if o.all && len(args) > 0 {
				return fmt.Errorf("--all cannot be combined with PRODUCT-UPDATE-ID arguments")
			}
			if !o.all && len(uniqueNonEmpty(args)) == 0 {
				return fmt.Errorf("provide at least one PRODUCT-UPDATE-ID, or --all")
			}
			return nil
		},
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
	cmd.Flags().BoolVar(&o.all, "all", false, dismissAllProductUpdatesFlag)
	addDryRunFlag(cmd)
	return cmd
}

func (o *dismissProductUpdatesOptions) run(ids []string) error {
	if o.all {
		// The ids come from the server, and a dry run must work without one.
		if global.DryRun {
			logger.Info("############### THIS IS A DRY-RUN  ###############")
			logger.Info("every unread product update would be dismissed")
			return nil
		}
		var err error
		ids, err = unreadProductUpdateIDs()
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			logger.Info("No unread product updates to dismiss.")
			return nil
		}
	}

	ids = uniqueNonEmpty(ids)
	if err := markProductUpdatesSeen(ids); err != nil {
		return err
	}
	if !global.DryRun {
		logger.Info("%d product update(s) dismissed", len(ids))
	}
	return nil
}

// markProductUpdatesSeen sends ids in requests of at most
// maxProductUpdateIDsPerRequest.
func markProductUpdatesSeen(ids []string) error {
	reqURL, err := url.JoinPath(global.Host, "api/v2/product-updates", global.Org, "seen")
	if err != nil {
		return err
	}

	for start := 0; start < len(ids); start += maxProductUpdateIDsPerRequest {
		end := min(start+maxProductUpdateIDsPerRequest, len(ids))
		_, err := kosliClient.Do(&requests.RequestParams{
			Method:  http.MethodPost,
			URL:     reqURL,
			Payload: productUpdatesSeenPayload{IDs: ids[start:end]},
			DryRun:  global.DryRun,
			Token:   global.ApiToken,
		})
		if err != nil {
			if start > 0 {
				return fmt.Errorf("dismissed %d of %d product update(s) before failing: %w", start, len(ids), err)
			}
			return err
		}
	}
	return nil
}

func unreadProductUpdateIDs() ([]string, error) {
	reqURL, err := url.JoinPath(global.Host, "api/v2/product-updates", global.Org)
	if err != nil {
		return nil, err
	}

	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	})
	if err != nil {
		return nil, err
	}

	listed := &listProductUpdatesResponse{}
	if err := json.Unmarshal([]byte(response.Body), listed); err != nil {
		return nil, err
	}

	ids := []string{}
	for _, update := range listed.ProductUpdates {
		if update.Unread {
			ids = append(ids, update.ID)
		}
	}
	return ids, nil
}

// uniqueNonEmpty returns values in order, without empty strings or repeats.
func uniqueNonEmpty(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
