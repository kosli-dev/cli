package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type searchOptions struct {
}

// const artifactCreationExample = `
// # Report to a Kosli pipeline that a file type artifact has been created
// kosli pipeline artifact report creation FILE.tgz \
// 	--api-token yourApiToken \
// 	--artifact-type file \
// 	--build-url https://exampleci.com \
// 	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--owner yourOrgName \
// 	--pipeline yourPipelineName

// # Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
// kosli pipeline artifact report creation \
// 	--api-token yourApiToken \
// 	--build-url https://exampleci.com \
// 	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--owner yourOrgName \
// 	--pipeline yourPipelineName \
// 	--sha256 yourSha256
// `

func newSearchCmd(out io.Writer) *cobra.Command {
	o := new(searchOptions)
	cmd := &cobra.Command{
		Use:   "search GIT-COMMIT",
		Short: "Search for a git commit in Kosli.",
		// Example: artifactCreationExample,
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "git commit argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}
	return cmd
}

func (o *searchOptions) run(args []string) error {
	var err error
	commit := args[0]

	url := fmt.Sprintf("%s/api/v1/%s/commits/%s", global.Host, global.Owner, commit)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, log)
	if err != nil {
		return err
	}

	fmt.Println(response.Body)
	return nil
}
