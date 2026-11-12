package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listProductUpdatesShortDesc = `List Kosli product updates.`

const listProductUpdatesLongDesc = listProductUpdatesShortDesc + `
Product updates are the new features and improvements published in Kosli's changelog.
Unread updates are listed first, then the newest. An update about a feature that is not
enabled for the org is left out. The first time you list them, in the CLI or in the Kosli app,
every update is marked as seen, so none is shown as unread; only updates published after that
start as unread. Dismiss those with ^kosli dismiss product-updates^.
The update's summary is included in the JSON output.

Product updates are tracked per user, so the API token must belong to a member of the org.`

const listProductUpdatesExample = `
# list product updates:
kosli list product-updates \
	--api-token yourAPIToken \
	--org yourOrgName

# list product updates in JSON:
kosli list product-updates \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listProductUpdatesOptions struct {
	output string
}

type listProductUpdatesResponse struct {
	ProductUpdates []productUpdate `json:"product_updates"`
	UnreadCount    int             `json:"unread_count"`
}

type productUpdate struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	PublishedAt json.Number `json:"published_at"`
	Unread      bool        `json:"unread"`
}

func newListProductUpdatesCmd(out io.Writer) *cobra.Command {
	o := new(listProductUpdatesOptions)
	cmd := &cobra.Command{
		Use:         "product-updates",
		Aliases:     []string{"pu", "pr-up"},
		Short:       listProductUpdatesShortDesc,
		Long:        listProductUpdatesLongDesc,
		Example:     listProductUpdatesExample,
		Args:        cobra.NoArgs,
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *listProductUpdatesOptions) run(out io.Writer) error {
	reqURL, err := url.JoinPath(global.Host, "api/v2/product-updates", global.Org)
	if err != nil {
		return err
	}

	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	})
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printProductUpdatesListAsTable,
			"json":  output.PrintJson,
		})
}

func printProductUpdatesListAsTable(raw string, out io.Writer, page int) error {
	response := &listProductUpdatesResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	if len(response.ProductUpdates) == 0 {
		logger.Info("No product updates were found.")
		return nil
	}

	header := []string{"ID", "TITLE", "PUBLISHED", "UNREAD", "URL"}
	rows := []string{}
	for _, update := range response.ProductUpdates {
		publishedAt, err := formattedTimestamp(update.PublishedAt, true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%t\t%s",
			update.ID, update.Title, publishedAt, update.Unread, update.URL))
	}

	rows = append(rows, fmt.Sprintf("\n%d unread of %d product updates",
		response.UnreadCount, len(response.ProductUpdates)))

	tabFormattedPrint(out, header, rows)

	return nil
}
