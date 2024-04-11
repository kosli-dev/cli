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

		hdrFunc := func(filename string, beta, deprecated bool) string {
			base := filepath.Base(filename)
			name := strings.TrimSuffix(base, path.Ext(base))
			title := strings.ToLower(strings.Replace(name, "_", " ", -1))
			return fmt.Sprintf("---\ntitle: \"%s\"\nbeta: %t\ndeprecated: %t\n---\n\n", title, beta, deprecated)
		}

		return MereklyGenMarkdownTreeCustom(o.topCmd, o.dest, hdrFunc, linkHandler)
	}
	return doc.GenMarkdownTree(o.topCmd, o.dest)
}

func MereklyGenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender func(string, bool, bool) string, linkHandler func(string) string) error {
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
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.WriteString(f, filePrepender(filename, isBeta(cmd), isDeprecated(cmd))); err != nil {
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
		buf.WriteString("{{< hint warning >}}")
		buf.WriteString(fmt.Sprintf("**%s** is a beta feature. ", name))
		buf.WriteString("Beta features provide early access to product functionality.  ")
		buf.WriteString("These features may change between releases without warning, or can be removed in a ")
		buf.WriteString("future release.\n")
		buf.WriteString("Please contact us to enable this feature for your organization.")
		// buf.WriteString("You can enable beta features by using the `kosli enable beta` command.")
		buf.WriteString("{{< /hint >}}\n")
	}

	if isDeprecated(cmd) {
		buf.WriteString("{{< hint danger >}}")
		buf.WriteString(fmt.Sprintf("**%s** is deprecated. %s  ", name, cmd.Deprecated))
		buf.WriteString("Deprecated commands will be removed in a future release.")
		buf.WriteString("{{< /hint >}}\n")
	}

	if len(cmd.Long) > 0 {
		buf.WriteString("## Synopsis\n\n")
		buf.WriteString(strings.Replace(cmd.Long, "^", "`", -1) + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.UseLine()))
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

	if len(cmd.Example) > 0 {
		// This is an attempt to tidy up the non-live examples, so they each have their own title.
		// Note: The contents of the title lines could also contain < and > characters which will
		// be lost if simply embedded in a md ## section.
		buf.WriteString("## Examples Use Cases\n\n")

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
					buf.WriteString(fmt.Sprintf("**%s**\n\n", strings.TrimSpace(title[1:])))
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
		example = append(example, line)
	}
	result = append(result, example)
	return result[1:]
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
