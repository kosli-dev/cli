package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

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

	return cmd
}

func (o *docsOptions) run() error {
	if o.generateHeaders {
		linkHandler := func(name string) string {
			base := strings.TrimSuffix(name, path.Ext(name))
			return "/client_reference/" + strings.ToLower(base) + "/"
		}

		hdrFunc := func(filename string, beta, deprecated bool, summary string) string {
			base := filepath.Base(filename)
			name := strings.TrimSuffix(base, path.Ext(base))
			title := strings.ToLower(strings.Replace(name, "_", " ", -1))
			return fmt.Sprintf("---\ntitle: \"%s\"\nbeta: %t\ndeprecated: %t\nsummary: \"%s\"\n---\n\n", title, beta, deprecated, summary)
		}

		return MereklyGenMarkdownTreeCustom(o.topCmd, o.dest, hdrFunc, linkHandler)
	}
	return doc.GenMarkdownTree(o.topCmd, o.dest)
}

func MereklyGenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender func(string, bool, bool, string) string, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		// skip all unavailable commands except deprecated ones
		if (!c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand()) && c.Deprecated == "" {
			continue
		}
		if err := MereklyGenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	if !cmd.HasParent() || !cmd.HasSubCommands() {
		basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
		filename := filepath.Join(dir, basename)
		summary := cmd.Short
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.WriteString(f, filePrepender(filename, isBeta(cmd), isDeprecated(cmd), summary)); err != nil {
			return err
		}
		if err := KosliGenMarkdownCustom(cmd, f, linkHandler); err != nil {
			return err
		}
	}
	return nil
}

// KosliGenMarkdownCustom creates custom markdown output.
func KosliGenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	buf.WriteString("# " + name + "\n\n")

	if isBeta(cmd) {
		buf.WriteString("{{% hint warning %}}\n")
		buf.WriteString(fmt.Sprintf("**%s** is a beta feature. ", name))
		buf.WriteString("Beta features provide early access to product functionality. ")
		buf.WriteString("These features may change between releases without warning, or can be removed in a ")
		buf.WriteString("future release.\n")
		buf.WriteString("Please contact us to enable this feature for your organization.\n")
		// buf.WriteString("You can enable beta features by using the `kosli enable beta` command.")
		buf.WriteString("{{% /hint %}}\n")
	}

	if isDeprecated(cmd) {
		buf.WriteString("{{% hint danger %}}\n")
		buf.WriteString(fmt.Sprintf("**%s** is deprecated. %s  ", name, cmd.Deprecated))
		buf.WriteString("Deprecated commands will be removed in a future release.\n")
		buf.WriteString("{{% /hint %}}\n")
	}

	if len(cmd.Long) > 0 {
		buf.WriteString("## Synopsis\n\n")
		if cmd.Runnable() {
			buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.UseLine()))
		}
		buf.WriteString(strings.Replace(cmd.Long, "^", "`", -1) + "\n\n")
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}

	urlSafeName := url.QueryEscape(name)
	liveExamplesBuf := new(bytes.Buffer)
	for _, ci := range []string{"GitHub", "GitLab"} {
		if liveYamlDocExists(ci, urlSafeName) {
			liveExamplesBuf.WriteString(fmt.Sprintf("{{< tab \"%v\" >}}", ci))
			liveExamplesBuf.WriteString(fmt.Sprintf("View an example of the `%s` command in %s.\n\n", name, ci))
			liveExamplesBuf.WriteString(fmt.Sprintf("In [this YAML file](%v)", yamlURL(ci, urlSafeName)))
			if liveEventDocExists(ci, urlSafeName) {
				liveExamplesBuf.WriteString(fmt.Sprintf(", which created [this Kosli Event](%v).", eventURL(ci, urlSafeName)))
			}
			liveExamplesBuf.WriteString("{{< /tab >}}")
		}
	}
	liveExamples := liveExamplesBuf.String()
	if len(liveExamples) > 0 {
		buf.WriteString("## Live Examples in different CI systems\n\n")
		buf.WriteString("{{< tabs \"live-examples\" \"col-no-wrap\" >}}")
		buf.WriteString(liveExamples)
		buf.WriteString("{{< /tabs >}}\n\n")
	}

	liveCliFullCommand, liveCliURL, liveCliExists := liveCliDocExists(name)
	if liveCliExists {
		buf.WriteString("## Live Example\n\n")
		buf.WriteString("{{< raw-html >}}")
		buf.WriteString(fmt.Sprintf("To view a live example of '%s' you can run the commands below (for the <a href=\"https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/\">cyber-dojo</a> demo organization).<br/><a href=\"%s\">Run the commands below and view the output.</a>", name, liveCliURL))
		buf.WriteString("<pre>")
		buf.WriteString("export KOSLI_ORG=cyber-dojo\n")
		buf.WriteString("export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only\n")
		buf.WriteString(liveCliFullCommand)
		buf.WriteString("</pre>")
		buf.WriteString("{{< / raw-html >}}\n\n")
	}

	if len(cmd.Example) > 0 {
		// This is an attempt to tidy up the non-live examples, so they each have their own title.
		// Note: The contents of the title lines could also contain < and > characters which will
		// be lost if simply embedded in a md ## section.
		buf.WriteString("## Examples Use Cases\n\n")
		url := "https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables"
		message := fmt.Sprintf("These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](%v). \n\n", url)
		buf.WriteString(message)

		// Some non-title lines contain a # character, (eg in a snappish) so we have to
		// split on newlines first and then only split on # in the first position
		example := strings.TrimSpace(cmd.Example)
		lines := strings.Split(example, "\n")

		// Some commands have #titles spanning several lines (that is, each title line starts with a # character)
		if name == "kosli report approval" {
			buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", example))
		} else if name == "kosli request approval" {
			buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", example))
		} else if name == "kosli snapshot server" {
			buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", example))
		} else if lines[0][0] != '#' {
			// Some commands, eg 'kosli assert snapshot' have no #title
			// and their example starts immediately with the kosli command.
			buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", example))
		} else {
			// The rest we can format nicely
			all := hashTitledExamples(lines)
			for i := 0; i < len(all); i++ {
				exampleLines := all[i]
				// Some titles have a trailing colon, some don't
				title := strings.Trim(exampleLines[0], ":")
				if len(title) > 0 {
					buf.WriteString(fmt.Sprintf("##### %s\n\n", strings.TrimSpace(title[1:])))
					buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", strings.Join(exampleLines[1:], "\n")))
				}
			}
		}
	}

	_, err := buf.WriteTo(w)
	return err
}

func hashTitledExamples(lines []string) [][]string {
	// Some non-title lines contain a # character, so we have split on newlines first
	// and then split on # which are the first character in their line
	result := make([][]string, 0)
	example := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			result = append(result, example) // See result[1:] at end
			example = make([]string, 0)
		}
		if !isSetWithEnvVar(line) {
			example = append(example, choppedLineContinuation(line))
		}
	}
	result = append(result, example)
	return result[1:]
}

func isSetWithEnvVar(line string) bool {
	trimmed_line := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed_line, "--api-token ") {
		return true
	} else if strings.HasPrefix(trimmed_line, "--host ") {
		return true
	} else if strings.HasPrefix(trimmed_line, "--org ") {
		return true
	} else if strings.HasPrefix(trimmed_line, "--flow ") {
		return true
	} else if strings.HasPrefix(trimmed_line, "--trail ") {
		return true
	} else {
		return false
	}
}

func choppedLineContinuation(line string) string {
	trimmed_line := strings.TrimRightFunc(line, unicode.IsSpace)
	return strings.TrimSuffix(trimmed_line, "\\")
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Flags\n")
		buf.WriteString("| Flag | Description |\n")
		buf.WriteString("| :--- | :--- |\n")
		usages := CommandsInTable(flags)
		fmt.Fprint(buf, usages)
		buf.WriteString("\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("## Flags inherited from parent commands\n")
		buf.WriteString("| Flag | Description |\n")
		buf.WriteString("| :--- | :--- |\n")
		usages := CommandsInTable(parentFlags)
		fmt.Fprint(buf, usages)
		buf.WriteString("\n\n")
	}
	return nil
}

const baseURL = "https://app.kosli.com/api/v2/livedocs/cyber-dojo"

func liveYamlDocExists(ci string, command string) bool {
	url := fmt.Sprintf("%v/yaml_exists?ci=%v&command=%v", baseURL, strings.ToLower(ci), command)
	return liveDocExists(url)
}

func liveEventDocExists(ci string, command string) bool {
	url := fmt.Sprintf("%v/event_exists?ci=%v&command=%v", baseURL, strings.ToLower(ci), command)
	return liveDocExists(url)
}

func liveCliDocExists(command string) (string, string, bool) {
	fullCommand, ok := liveCliMap[command]
	if ok {
		plussed := strings.Replace(fullCommand, " ", "+", -1)
		exists_url := fmt.Sprintf("%v/cli_exists?command=%v", baseURL, plussed)
		url := fmt.Sprintf("%v/cli?command=%v", baseURL, plussed)
		return fullCommand, url, liveDocExists(exists_url)
	} else {
		return "", "", false
	}
}

func liveDocExists(url string) bool {
	response, err := http.Get(url)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	var exists bool
	err = decoder.Decode(&exists)
	if err != nil {
		return false
	}
	return exists
}

func yamlURL(ci string, command string) string {
	return fmt.Sprintf("%v/yaml?ci=%v&command=%v", baseURL, strings.ToLower(ci), command)
}

func eventURL(ci string, command string) string {
	return fmt.Sprintf("%v/event?ci=%v&command=%v", baseURL, strings.ToLower(ci), command)
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
