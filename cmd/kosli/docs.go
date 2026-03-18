package main

import (
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// Exit code sets used across command doc generation.
var (
	// exitCodesDefault applies to all commands that call the Kosli API.
	exitCodesDefault = []docgen.ExitCodeEntry{
		{Code: 0, Meaning: "No error."},
		{Code: 2, Meaning: "Kosli server is unreachable or returned a server error."},
		{Code: 3, Meaning: "Invalid API token or unauthorized access."},
		{Code: 4, Meaning: "CLI usage error (e.g. missing or invalid flags)."},
	}

	// exitCodesAssert applies to assert commands that can signal compliance violations.
	exitCodesAssert = []docgen.ExitCodeEntry{
		{Code: 0, Meaning: "No error."},
		{Code: 1, Meaning: "Assertion/compliance violation."},
		{Code: 2, Meaning: "Kosli server is unreachable or returned a server error."},
		{Code: 3, Meaning: "Invalid API token or unauthorized access."},
		{Code: 4, Meaning: "CLI usage error (e.g. missing or invalid flags)."},
	}

	// exitCodesAssertStatus applies to `kosli assert status` which only checks reachability.
	exitCodesAssertStatus = []docgen.ExitCodeEntry{
		{Code: 0, Meaning: "Kosli server is responsive."},
		{Code: 2, Meaning: "Kosli server is unreachable or down."},
	}

	// exitCodesNoAPI applies to commands that do not call the Kosli API (version, completion, docs).
	exitCodesNoAPI = []docgen.ExitCodeEntry{
		{Code: 0, Meaning: "No error."},
		{Code: 4, Meaning: "CLI usage error."},
	}
)

// commandExitCodes maps command paths to their exit code sets.
// Commands not in this map receive exitCodesDefault.
var commandExitCodes = map[string][]docgen.ExitCodeEntry{
	"kosli assert artifact": exitCodesAssert,
	"kosli assert approval": exitCodesAssert,
	"kosli assert snapshot": exitCodesAssert,
	"kosli assert status":   exitCodesAssertStatus,
	"kosli version":         exitCodesNoAPI,
	"kosli completion":      exitCodesNoAPI,
	"kosli docs":            exitCodesNoAPI,
}

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
			path := cmd.CommandPath()
			exitCodes, ok := commandExitCodes[path]
			if !ok && cmd.Runnable() {
				exitCodes = exitCodesDefault
			}
			return docgen.CommandMeta{
				Name:       path,
				Beta:       isBeta(cmd),
				Deprecated: isDeprecated(cmd),
				DeprecMsg:  cmd.Deprecated,
				Summary:    cmd.Short,
				Long:       cmd.Long,
				UseLine:    cmd.UseLine(),
				Runnable:   cmd.Runnable(),
				Example:    cmd.Example,
				ExitCodes:  exitCodes,
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

func buildDocURL(base, path, ci, command string) string {
	u, err := neturl.JoinPath(base, path)
	if err != nil {
		return ""
	}
	q := neturl.Values{}
	q.Set("ci", strings.ToLower(ci))
	q.Set("command", command)
	return u + "?" + q.Encode()
}

func buildCLIDocURL(base, path, command string) string {
	u, err := neturl.JoinPath(base, path)
	if err != nil {
		return ""
	}
	q := neturl.Values{}
	q.Set("command", command)
	return u + "?" + q.Encode()
}

func (p *kosliLiveDocProvider) YamlDocExists(ci, command string) bool {
	return liveDocExists(buildDocURL(baseURL, "yaml_exists", ci, command))
}

func (p *kosliLiveDocProvider) EventDocExists(ci, command string) bool {
	return liveDocExists(buildDocURL(baseURL, "event_exists", ci, command))
}

func (p *kosliLiveDocProvider) YamlURL(ci, command string) string {
	return buildDocURL(baseURL, "yaml", ci, command)
}

func (p *kosliLiveDocProvider) EventURL(ci, command string) string {
	return buildDocURL(baseURL, "event", ci, command)
}

func (p *kosliLiveDocProvider) CLIDocExists(command string) (string, string, bool) {
	fullCommand, ok := liveCliMap[command]
	if ok {
		plussed := strings.ReplaceAll(fullCommand, " ", "+")
		existsURL := buildCLIDocURL(baseURL, "cli_exists", plussed)
		url := buildCLIDocURL(baseURL, "cli", plussed)
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
