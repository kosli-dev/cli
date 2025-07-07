package main

import (
	"encoding/json"
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
# get trail attestations of a given type and name for a range of commits in a flow (requires the --flow flag)
kosli get attestations \ 
	--flow flowName \
	--type attestationType \
	--name attestationName \
	--commit-range commitRange 

# get artifact attestations of a given type and name for a range of commits in a flow (requires the --flow flag)
kosli get attestations \
	--flow flowName \
	--type attestationType \
	--name slotName.attestationName \
	--commit-range commitRange 

# get all attestations of a given type from a flow
kosli get attestations \
	--flow flowName \
	--type attestationType
`

type getAttestationsOptions struct {
	output          string
	flowName        string
	attestationType string
	attestationName string
	commitRange     string
	repositoryRoot  string
}

type SingleAttestation struct {
	Name                string  `json:"attestation_name"`
	Type                string  `json:"attestation_type"`
	Compliance          bool    `json:"is_compliant"`
	ArtifactFingerprint string  `json:"artifact_fingerprint,omitempty"`
	CreatedAt           float64 `json:"created_at"`
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
	var commitDict map[string][]SingleAttestation
	err := json.Unmarshal([]byte(raw), &commitDict)
	if err != nil {
		return err
	}

	if len(commitDict) == 0 {
		logger.Info("No commits found.")
		return nil
	}

	header := []string{"COMMIT", "ATTESTATION NAME", "ATTESTATION TYPE", "COMPLIANT", "CREATED_AT"}
	rows := []string{}
	for commit, attestations := range commitDict {
		for i, attestation := range attestations {
			createdAt, err := formattedTimestamp(attestation.CreatedAt, true)
			if err != nil {
				return err
			}
			if i == 0 {
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%t\t%s", commit, attestation.Name, attestation.Type, attestation.Compliance, createdAt))
			} else {
				rows = append(rows, fmt.Sprintf("\t%s\t%s\t%t\t%s", attestation.Name, attestation.Type, attestation.Compliance, createdAt))
			}
		}
		rows = append(rows, "\t\t\t")
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
