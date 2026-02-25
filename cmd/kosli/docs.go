package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsShortDesc = `Generate documentation files for Kosli CLI. `

const docsLongDesc = docsShortDesc + `
This command can generate documentation in the following formats: Markdown.
`

type docsOptions struct {
	dest            string
	topCmd          *cobra.Command
	generateHeaders bool
	mintlify        bool
}

func newDocsCmd(out io.Writer) *cobra.Command {
	o := &docsOptions{}

	cmd := &cobra.Command{
		Use:    "docs",
		Short:  docsShortDesc,
		Long:   docsLongDesc,
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.topCmd = cmd.Root()
			return o.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&o.dest, "dir", "./", "The directory to which documentation is written.")
	f.BoolVar(&o.generateHeaders, "generate-headers", true, "Generate standard headers for markdown files.")
	f.BoolVar(&o.mintlify, "mintlify", false, "Generate Mintlify-compatible MDX output instead of Hugo.")

	return cmd
}

func (o *docsOptions) run() error {
	if o.generateHeaders {
		var formatter docgen.Formatter
		if o.mintlify {
			formatter = docgen.MintlifyFormatter{}
		} else {
			formatter = docgen.HugoFormatter{}
		}

		metaFn := func(cmd *cobra.Command) docgen.CommandMeta {
			return docgen.CommandMeta{
				Name:       cmd.CommandPath(),
				Beta:       isBeta(cmd),
				Deprecated: isDeprecated(cmd),
				DeprecMsg:  cmd.Deprecated,
				Summary:    cmd.Short,
				Long:       cmd.Long,
				UseLine:    cmd.UseLine(),
				Runnable:   cmd.Runnable(),
				Example:    cmd.Example,
			}
		}

		return docgen.GenMarkdownTree(o.topCmd, o.dest, formatter, metaFn, &kosliLiveDocProvider{})
	}
	return doc.GenMarkdownTree(o.topCmd, o.dest)
}

// kosliLiveDocProvider implements docgen.LiveDocProvider using HTTP calls
// to the Kosli live docs API.
type kosliLiveDocProvider struct{}

const baseURL = "https://app.kosli.com/api/v2/livedocs/cyber-dojo"

func (p *kosliLiveDocProvider) YamlDocExists(ci, command string) bool {
	url := fmt.Sprintf("%s/yaml_exists?ci=%s&command=%s", baseURL, strings.ToLower(ci), command)
	return liveDocExists(url)
}

func (p *kosliLiveDocProvider) EventDocExists(ci, command string) bool {
	url := fmt.Sprintf("%s/event_exists?ci=%s&command=%s", baseURL, strings.ToLower(ci), command)
	return liveDocExists(url)
}

func (p *kosliLiveDocProvider) YamlURL(ci, command string) string {
	return fmt.Sprintf("%s/yaml?ci=%s&command=%s", baseURL, strings.ToLower(ci), command)
}

func (p *kosliLiveDocProvider) EventURL(ci, command string) string {
	return fmt.Sprintf("%s/event?ci=%s&command=%s", baseURL, strings.ToLower(ci), command)
}

func (p *kosliLiveDocProvider) CLIDocExists(command string) (string, string, bool) {
	fullCommand, ok := liveCliMap[command]
	if ok {
		plussed := strings.ReplaceAll(fullCommand, " ", "+")
		existsURL := fmt.Sprintf("%s/cli_exists?command=%s", baseURL, plussed)
		url := fmt.Sprintf("%s/cli?command=%s", baseURL, plussed)
		return fullCommand, url, liveDocExists(existsURL)
	}
	return "", "", false
}

func liveDocExists(url string) bool {
	response, err := http.Get(url) //nolint:gosec
	if err != nil {
		return false
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			logger.Warn("failed to close response body: %v", err)
		}
	}()
	decoder := json.NewDecoder(response.Body)
	var exists bool
	err = decoder.Decode(&exists)
	if err != nil {
		return false
	}
	return exists
}

var liveCliMap = map[string]string{
	"kosli list environments": "kosli list environments --output=json",
	"kosli get environment":   "kosli get environment aws-prod --output=json",
	"kosli log environment":   "kosli log environment aws-prod --output=json",
	"kosli list snapshots":    "kosli list snapshots aws-prod --output=json",
	"kosli get snapshot":      "kosli get snapshot aws-prod --output=json",
	"kosli diff snapshots":    "kosli diff snapshots aws-beta aws-prod --output=json",
	"kosli list flows":        "kosli list flows --output=json",
	"kosli get flow":          "kosli get flow dashboard-ci --output=json",
	//"kosli list trails":       "kosli list trails dashboard-ci --output=json",  // Produces too much output
	"kosli get trail":       "kosli get trail dashboard-ci 1159a6f1193150681b8484545150334e89de6c1c --output=json",
	"kosli get attestation": "kosli get attestation snyk-container-scan --flow=differ-ci --fingerprint=0cbbe3a6e73e733e8ca4b8813738d68e824badad0508ff20842832b5143b48c0 --output=json",
}
