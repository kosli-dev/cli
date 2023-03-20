package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitGenericOptions struct {
	userDataFile string
	evidenceFile string
	payload      GenericEvidencePayload
}

// type GenericEvidencePayloadWithFile struct {
// 	GenericEvidencePayload `json:"evidence_json"`
// 	File                   *bytes.Buffer `json:"evidence_file,omitempty"`
// }

const reportEvidenceCommitGenericShortDesc = `Report Generic evidence for a commit in Kosli flows.`

const reportEvidenceCommitGenericLongDesc = reportEvidenceCommitGenericShortDesc

const reportEvidenceCommitGenericExample = `
# report Generic evidence for a commit related to one Kosli flow:
kosli report evidence commit generic \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--description "some description" \
	--compliant \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName

# report Generic evidence for a commit related to multiple Kosli flows with user-data:
kosli report evidence commit generic \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--description "some description" \
	--compliant \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--user-data /path/to/json/file.json
`

func newReportEvidenceCommitGenericCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceCommitGenericOptions)
	cmd := &cobra.Command{
		Use:     "generic",
		Short:   reportEvidenceCommitGenericShortDesc,
		Long:    reportEvidenceCommitGenericLongDesc,
		Example: reportEvidenceCommitGenericExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addCommitEvidenceFlags(cmd, &o.payload.TypedEvidencePayload, ci)
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", false, evidenceCompliantFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)

	cmd.Flags().StringVar(&o.evidenceFile, "evidence-file", "", evidenceFileFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitGenericOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v1/projects/%s/evidence/commit/generic", global.Host, global.Owner)
	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(o.payload, "", "    ")
	if err != nil {
		return err
	}

	contentType, body, err := createEvidenceMultipartForm(jsonBytes, o.evidenceFile)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:            http.MethodPost,
		URL:               url,
		AdditionalHeaders: map[string]string{"Content-Type": contentType},
		DryRun:            global.DryRun,
		Password:          global.ApiToken,
	}

	_, err = kosliClient.DoWithFile(reqParams, body)
	if err == nil && !global.DryRun {
		logger.Info("generic evidence '%s' is reported to commit: %s", o.payload.EvidenceName, o.payload.CommitSHA)
	}
	return err
}

func createEvidenceMultipartForm(evidenceJson []byte, evidenceFile string) (string, *bytes.Buffer, error) {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormField("evidence_json")
	if err != nil {
		return "", body, err
	}

	_, err = part.Write(evidenceJson)
	if err != nil {
		return "", body, err
	}

	if evidenceFile != "" {
		file, err := os.Open(evidenceFile)
		if err != nil {
			return "", body, err
		}
		defer file.Close()

		part, err = writer.CreateFormFile("evidence_file", filepath.Base(evidenceFile))
		if err != nil {
			return "", body, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return "", body, err
		}
	}

	contentType := writer.FormDataContentType()

	return contentType, body, nil
}
