package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getAttestationShortDesc = `Get attestation by name from a specified trail or artifact.  `

const getAttestationLongDesc = getAttestationShortDesc + `
You can get an attestation from a trail or artifact using its name. The attestation name should be given
WITHOUT dot-notation.

To get an attestation from a trail, specify the trail name using the --trail flag.  
To get an attestation from an artifact, specify the artifact fingerprint using the --fingerprint flag.
The fingerprint can be short or complete.

In both cases the flow must also be specified using the --flow flag.

If there are multiple attestations with the same name on the trail or artifact, a list of all will be returned.
`

const getAttestationExample = `
# get an attestation from a trail (requires the --trail flag)
kosli get attestation attestationName \
	--flow flowName \
	--trail trailName 

# get an attestation from an artifact 
kosli get attestation attestationName \
	--flow flowName \
	--fingerprint fingerprint 
`

type getAttestationOptions struct {
	output      string
	flow        string
	trail       string
	fingerprint string
}

type Attestation struct {
	Name                string         `json:"attestation_name"`
	Type                string         `json:"attestation_type"`
	Compliance          bool           `json:"is_compliant"`
	ArtifactFingerprint string         `json:"artifact_fingerprint,omitempty"`
	CreatedAt           float64        `json:"created_at"`
	GitCommitInfo       *GitCommitInfo `json:"git_commit_info,omitempty"`
	HtmlUrl             string         `json:"html_url"`
}

type GitCommitInfo struct {
	Sha1      string  `json:"sha1"`
	Author    string  `json:"author"`
	Message   string  `json:"message"`
	Branch    string  `json:"branch"`
	Url       string  `json:"url,omitempty"`
	Timestamp float64 `json:"timestamp"`
}

func newGetAttestationCmd(out io.Writer) *cobra.Command {
	o := new(getAttestationOptions)
	cmd := &cobra.Command{
		Use:     "attestation ATTESTATION-NAME",
		Short:   getAttestationShortDesc,
		Long:    getAttestationLongDesc,
		Example: getAttestationExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			err = MuXRequiredFlags(cmd, []string{"trail", "fingerprint"}, true)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().StringVarP(&o.flow, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.trail, "trail", "t", "", getAttestationTrailFlag)
	cmd.Flags().StringVarP(&o.fingerprint, "fingerprint", "F", "", getAttestationFingerprintFlag)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *getAttestationOptions) run(out io.Writer, args []string) error {
	var url string
	baseUrl := fmt.Sprintf("%s/api/v2/attestations/%s/%s", global.Host, global.Org, o.flow)
	if o.trail != "" {
		url = fmt.Sprintf("%s/trail/%s", baseUrl, o.trail)
	}

	if o.fingerprint != "" {
		url = fmt.Sprintf("%s/artifact/%s", baseUrl, o.fingerprint)
	}

	url = fmt.Sprintf("%s/%s", url, args[0])

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}

	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printAttestationsAsTable,
			"json":  output.PrintJson,
		})
}

func printAttestationsAsTable(raw string, out io.Writer, pageNumber int) error {
	var attestations []Attestation
	err := json.Unmarshal([]byte(raw), &attestations)
	if err != nil {
		return err
	}

	if len(attestations) == 0 {
		logger.Info("No attestations found.")
		return nil
	}

	separator := ""
	for _, attestation := range attestations {
		rows := []string{}
		rows = append(rows, fmt.Sprintf("Name:\t%s", attestation.Name))
		rows = append(rows, fmt.Sprintf("Type:\t%s", attestation.Type))
		rows = append(rows, fmt.Sprintf("Compliance:\t%t", attestation.Compliance))

		createdAt, err := formattedTimestamp(attestation.CreatedAt, false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAt))

		if attestation.ArtifactFingerprint != "" {
			rows = append(rows, fmt.Sprintf("Artifact fingerprint:\t%s", attestation.ArtifactFingerprint))
		}
		if attestation.GitCommitInfo != nil {
			rows = append(rows, "Git Commit Info:")
			rows = append(rows, fmt.Sprintf("    Sha1:\t%s", attestation.GitCommitInfo.Sha1))
			rows = append(rows, fmt.Sprintf("    Author:\t%s", attestation.GitCommitInfo.Author))
			rows = append(rows, fmt.Sprintf("    Branch:\t%s", attestation.GitCommitInfo.Branch))
			rows = append(rows, fmt.Sprintf("    Commit URL:\t%s", attestation.GitCommitInfo.Url))
			timestamp, err := formattedTimestamp(attestation.GitCommitInfo.Timestamp, false)
			if err != nil {
				return err
			}
			rows = append(rows, fmt.Sprintf("    Timestamp:\t%s", timestamp))
		}

		if attestation.HtmlUrl != "" {
			rows = append(rows, fmt.Sprintf("Attestation URL:\t%s", attestation.HtmlUrl))
		}

		fmt.Print(separator)
		separator = "\n"
		tabFormattedPrint(out, []string{}, rows)
	}
	return nil
}
