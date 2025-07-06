package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getAttestationsShortDesc = `Get attestations.  `

const getAttestationsLongDesc = getAttestationShortDesc + ``

const getAttestationsExample = `
# get an attestation from a trail (requires the --trail flag)
kosli get attestation attestationName \
	--flow flowName \
	--trail trailName 

# get an attestation from an artifact 
kosli get attestation attestationName \
	--flow flowName \
	--fingerprint fingerprint 
`

type getAttestationsOptions struct {
	output          string
	flowName        string
	attestationType string
	attestationName string
	commitRange     string
	repositoryRoot  string
}

func newGetAttestationsCmd(out io.Writer) *cobra.Command {
	o := new(getAttestationsOptions)
	cmd := &cobra.Command{
		Use:     "attestations",
		Hidden:  true,
		Short:   getAttestationsShortDesc,
		Long:    getAttestationsLongDesc,
		Example: getAttestationsExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVar(&o.attestationType, "type", "", "attestationTypeFlag")
	cmd.Flags().StringVar(&o.attestationName, "name", "", "attestationNameFlag")
	cmd.Flags().StringVar(&o.commitRange, "commit-range", "", "commitRangeFlag")
	cmd.Flags().StringVar(&o.repositoryRoot, "repository-root", ".", "repositoryRootFlag")
	return cmd
}

func (o *getAttestationsOptions) run(out io.Writer, args []string) error {
	baseURL := fmt.Sprintf("%s/api/v2/attestations/%s/list_attestations_for_criteria", global.Host, global.Org)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	queryParams := parsedURL.Query()

	if o.flowName != "" {
		queryParams.Add("flow_name", o.flowName)
	}
	if o.attestationType != "" {
		queryParams.Add("attestation_type", o.attestationType)
	}
	if o.attestationName != "" {
		queryParams.Add("attestation_name", o.attestationName)
	}
	if o.commitRange != "" {
		commitRange := strings.Split(o.commitRange, "..")
		if len(commitRange) != 2 {
			return fmt.Errorf("invalid commit range: %s", o.commitRange)
		}
		baseRef := commitRange[0]
		targetRef := commitRange[1]

		gitView, err := gitview.New(o.repositoryRoot)
		if err != nil {
			return err
		}
		commits, err := gitView.CommitsBetween(baseRef, targetRef, logger)
		if err != nil {
			return err
		}
		for _, commit := range commits {
			queryParams.Add("commit_list", commit.Sha1)
		}
	}

	parsedURL.RawQuery = queryParams.Encode()
	finalURL := parsedURL.String()

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    finalURL,
		Token:  global.ApiToken,
	}

	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printFilteredAttestationsAsTable,
			"json":  output.PrintJson,
		})
}

func printFilteredAttestationsAsTable(raw string, out io.Writer, pageNumber int) error {
	// var attestations []Attestation
	// err := json.Unmarshal([]byte(raw), &attestations)
	// if err != nil {
	// 	return err)
	// }

	// if len(attestations) == 0 {
	// 	logger.Info("No attestations found.")
	// 	return nil
	// }

	// separator := ""
	// for _, attestation := range attestations {
	// 	rows := []string{}
	// 	rows = append(rows, fmt.Sprintf("Name:\t%s", attestation.Name))
	// 	rows = append(rows, fmt.Sprintf("Type:\t%s", attestation.Type))
	// 	rows = append(rows, fmt.Sprintf("Compliance:\t%t", attestation.Compliance))

	// 	createdAt, err := formattedTimestamp(attestation.CreatedAt, false)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAt))

	// 	if attestation.ArtifactFingerprint != "" {
	// 		rows = append(rows, fmt.Sprintf("Artifact fingerprint:\t%s", attestation.ArtifactFingerprint))
	// 	}
	// 	if attestation.GitCommitInfo != nil {
	// 		rows = append(rows, "Git Commit Info:")
	// 		rows = append(rows, fmt.Sprintf("    Sha1:\t%s", attestation.GitCommitInfo.Sha1))
	// 		rows = append(rows, fmt.Sprintf("    Author:\t%s", attestation.GitCommitInfo.Author))
	// 		rows = append(rows, fmt.Sprintf("    Branch:\t%s", attestation.GitCommitInfo.Branch))
	// 		rows = append(rows, fmt.Sprintf("    Commit URL:\t%s", attestation.GitCommitInfo.Url))
	// 		timestamp, err := formattedTimestamp(attestation.GitCommitInfo.Timestamp, false)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		rows = append(rows, fmt.Sprintf("    Timestamp:\t%s", timestamp))
	// 	}

	// 	if attestation.HtmlUrl != "" {
	// 		rows = append(rows, fmt.Sprintf("Attestation URL:\t%s", attestation.HtmlUrl))
	// 	}

	// 	fmt.Print(separator)
	// 	separator = "\n"
	// 	tabFormattedPrint(out, []string{}, rows)
	// }
	return nil
}
