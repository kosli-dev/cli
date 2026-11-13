package main

import (
	"bytes"
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
Only unread updates are listed; use ^--all^ to list the ones you have seen too, unread first,
then the newest. An update about a feature that is not enabled for the org is left out.
The first 10 are listed, as in the Kosli app; use ^--limit^ to list more, or ^--limit 0^ for
no limit. ^--all^ and ^--limit^ apply to the JSON output as well, which also holds each
update's summary. Its ^unread_count^ is always the total number of unread updates, however
many are listed.

An update is unread when you have not seen it and it was published in the last 12 months.
A user who signed up less than 7 days before first listing product updates, in the CLI or in
the Kosli app, starts with every update seen. Listing leaves updates unread: use ^--dismiss^
to mark the unread updates listed as seen, or ^kosli dismiss product-updates^ later.

Product updates are tracked per user, so the API token must belong to a member of the org.`

const listProductUpdatesExample = `
# list unread product updates:
kosli list product-updates \
	--api-token yourAPIToken \
	--org yourOrgName

# list unread product updates and mark them as seen:
kosli list product-updates \
	--dismiss \
	--api-token yourAPIToken \
	--org yourOrgName

# list every product update, read ones included:
kosli list product-updates \
	--all \
	--limit 0 \
	--api-token yourAPIToken \
	--org yourOrgName

# list unread product updates in JSON:
kosli list product-updates \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listProductUpdatesOptions struct {
	output  string
	limit   int
	all     bool
	dismiss bool
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
			return o.validate(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().IntVar(&o.limit, "limit", 10, productUpdatesLimitFlag)
	cmd.Flags().BoolVar(&o.all, "all", false, productUpdatesAllFlag)
	cmd.Flags().BoolVar(&o.dismiss, "dismiss", false, productUpdatesDismissFlag)

	return cmd
}

func (o *listProductUpdatesOptions) validate(cmd *cobra.Command) error {
	if o.limit < 0 {
		return ErrorBeforePrintingUsage(cmd, "--limit must be 0 or more")
	}
	return nil
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

	selected, err := selectProductUpdates(response.Body, o.all, o.limit)
	if err != nil {
		return err
	}

	err = output.FormattedPrint(selected.body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": func(raw string, out io.Writer, page int) error {
				return printProductUpdatesListAsTable(raw, out, selected.matching, o.all)
			},
			"json": output.PrintJson,
		})
	if err != nil || !o.dismiss || len(selected.unreadIDs) == 0 {
		return err
	}

	if err := markProductUpdatesSeen(selected.unreadIDs); err != nil {
		return err
	}
	// A count after the JSON would make the output unparseable.
	if !global.DryRun && o.output != "json" {
		logger.Info("%d product update(s) dismissed", len(selected.unreadIDs))
	}
	return nil
}

type selectedProductUpdates struct {
	// body is the list response holding only the updates selected.
	body string
	// matching counts the updates that qualified before the limit cut them.
	matching int
	// unreadIDs are the ids of the unread updates in body.
	unreadIDs []string
}

// selectProductUpdates keeps the unread updates of a list response, or all of
// them with includeRead, and of those the first limit, or all when limit is 0.
func selectProductUpdates(raw string, includeRead bool, limit int) (*selectedProductUpdates, error) {
	// Only product_updates is decoded, so the rest of the server's answer
	// reaches the JSON output as it was sent.
	response := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return nil, err
	}
	updates := []json.RawMessage{}
	if listed, ok := response["product_updates"]; ok {
		if err := json.Unmarshal(listed, &updates); err != nil {
			return nil, err
		}
	}

	selected := &selectedProductUpdates{unreadIDs: []string{}}
	kept := []json.RawMessage{}
	for _, rawUpdate := range updates {
		update := productUpdate{}
		if err := json.Unmarshal(rawUpdate, &update); err != nil {
			return nil, err
		}
		if !includeRead && !update.Unread {
			continue
		}
		selected.matching++
		if limit != 0 && len(kept) == limit {
			continue
		}
		kept = append(kept, rawUpdate)
		if update.Unread {
			selected.unreadIDs = append(selected.unreadIDs, update.ID)
		}
	}

	keptJSON, err := marshalKeepingHTMLCharacters(kept)
	if err != nil {
		return nil, err
	}
	response["product_updates"] = keptJSON
	body, err := marshalKeepingHTMLCharacters(response)
	if err != nil {
		return nil, err
	}
	selected.body = string(body)
	return selected, nil
}

// marshalKeepingHTMLCharacters encodes v as json.Marshal does, but leaves the
// &, < and > of a title or summary as they are instead of as \u escapes.
func marshalKeepingHTMLCharacters(v any) ([]byte, error) {
	encoded := &bytes.Buffer{}
	encoder := json.NewEncoder(encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(encoded.Bytes()), nil
}

func printProductUpdatesListAsTable(raw string, out io.Writer, matching int, includeRead bool) error {
	response := &listProductUpdatesResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	shown := len(response.ProductUpdates)
	if shown == 0 {
		if includeRead {
			logger.Info("No product updates were found.")
		} else {
			logger.Info("No unread product updates. Use --all to list the ones you have seen.")
		}
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

	rows = append(rows, "\n"+productUpdatesFooter(shown, matching, response.UnreadCount, includeRead))

	tabFormattedPrint(out, header, rows)

	return nil
}

func productUpdatesFooter(shown, matching, unreadCount int, includeRead bool) string {
	const more = " Use --limit to list more."
	switch {
	case includeRead && shown < matching:
		return fmt.Sprintf("Showing %d of %d product updates, %d unread in total.", shown, matching, unreadCount) + more
	case includeRead:
		return fmt.Sprintf("%d unread of %d product updates", unreadCount, matching)
	case shown < matching:
		return fmt.Sprintf("Showing %d of %d unread product updates.", shown, matching) + more
	default:
		return fmt.Sprintf("%d unread product update(s)", matching)
	}
}
