package docgen

import (
	"fmt"
	"path"
	"strings"
)

// HugoFormatter generates Hugo-compatible markdown documentation.
type HugoFormatter struct{}

func (HugoFormatter) Title(name string) string {
	return "# " + name + "\n\n"
}

func (HugoFormatter) FrontMatter(meta CommandMeta) string {
	return fmt.Sprintf("---\ntitle: \"%s\"\nbeta: %t\ndeprecated: %t\nsummary: \"%s\"\n---\n\n",
		meta.Name, meta.Beta, meta.Deprecated, meta.Summary)
}

func (HugoFormatter) BetaWarning(name string) string {
	var b strings.Builder
	b.WriteString("{{% hint warning %}}\n")
	fmt.Fprintf(&b, "**%s** is a beta feature. ", name)
	fmt.Fprintf(&b, "Beta features provide early access to product functionality. ")
	fmt.Fprintf(&b, "These features may change between releases without warning, or can be removed in a ")
	fmt.Fprintf(&b, "future release.\n")
	fmt.Fprintf(&b, "Please contact us to enable this feature for your organization.\n")
	b.WriteString("{{% /hint %}}\n")
	return b.String()
}

func (HugoFormatter) DeprecatedWarning(name, message string) string {
	var b strings.Builder
	b.WriteString("{{% hint danger %}}\n")
	fmt.Fprintf(&b, "**%s** is deprecated. %s  ", name, message)
	fmt.Fprintf(&b, "Deprecated commands will be removed in a future release.\n")
	b.WriteString("{{% /hint %}}\n")
	return b.String()
}

func (HugoFormatter) Synopsis(meta CommandMeta) string {
	var b strings.Builder
	if len(meta.Long) > 0 {
		b.WriteString("## Synopsis\n\n")
		if meta.Runnable {
			fmt.Fprintf(&b, "```shell\n%s\n```\n\n", meta.UseLine)
		}
		b.WriteString(strings.ReplaceAll(meta.Long, "^", "`") + "\n\n")
	}
	return b.String()
}

func (HugoFormatter) FlagsSection(flags, inherited string) string {
	var b strings.Builder
	if flags != "" {
		b.WriteString("## Flags\n")
		b.WriteString("| Flag | Description |\n")
		b.WriteString("| :--- | :--- |\n")
		b.WriteString(flags)
		b.WriteString("\n\n")
	}
	if inherited != "" {
		b.WriteString("## Flags inherited from parent commands\n")
		b.WriteString("| Flag | Description |\n")
		b.WriteString("| :--- | :--- |\n")
		b.WriteString(inherited)
		b.WriteString("\n\n")
	}
	return b.String()
}

func (HugoFormatter) LiveCIExamples(examples []CIExample, commandName string) string {
	if len(examples) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Live Examples in different CI systems\n\n")
	b.WriteString("{{< tabs \"live-examples\" \"col-no-wrap\" >}}")
	for _, ex := range examples {
		fmt.Fprintf(&b, "{{< tab \"%v\" >}}", ex.CI)
		fmt.Fprintf(&b, "View an example of the `%s` command in %s.\n\n", commandName, ex.CI)
		fmt.Fprintf(&b, "In [this YAML file](%v)", ex.YamlURL)
		if ex.EventURL != "" {
			fmt.Fprintf(&b, ", which created [this Kosli Event](%v).", ex.EventURL)
		}
		b.WriteString("{{< /tab >}}")
	}
	b.WriteString("{{< /tabs >}}\n\n")
	return b.String()
}

func (HugoFormatter) LiveCLIExample(commandName, fullCommand, url string) string {
	var b strings.Builder
	b.WriteString("## Live Example\n\n")
	b.WriteString("{{< raw-html >}}")
	fmt.Fprintf(&b, "To view a live example of '%s' you can run the commands below (for the <a href=\"https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/\">cyber-dojo</a> demo organization).<br/><a href=\"%s\">Run the commands below and view the output.</a>", commandName, url)
	b.WriteString("<pre>")
	b.WriteString("export KOSLI_ORG=cyber-dojo\n")
	b.WriteString("export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only\n")
	b.WriteString(fullCommand)
	b.WriteString("</pre>")
	b.WriteString("{{< / raw-html >}}\n\n")
	return b.String()
}

func (HugoFormatter) ExampleUseCases(commandName, example string) string {
	var b strings.Builder
	b.WriteString("## Examples Use Cases\n\n")
	url := "https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables"
	fmt.Fprintf(&b, "These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](%v). \n\n", url)

	example = strings.TrimSpace(example)
	lines := strings.Split(example, "\n")

	if commandName == "kosli report approval" ||
		commandName == "kosli request approval" ||
		commandName == "kosli snapshot server" {
		fmt.Fprintf(&b, "```shell\n%s\n```\n\n", example)
	} else if lines[0][0] != '#' {
		fmt.Fprintf(&b, "```shell\n%s\n```\n\n", example)
	} else {
		all := HashTitledExamples(lines)
		for i := 0; i < len(all); i++ {
			exampleLines := all[i]
			title := strings.Trim(exampleLines[0], ":")
			if len(title) > 0 {
				fmt.Fprintf(&b, "##### %s\n\n", strings.TrimSpace(title[1:]))
				fmt.Fprintf(&b, "```shell\n%s\n```\n\n", strings.Join(exampleLines[1:], "\n"))
			}
		}
	}
	return b.String()
}

func (HugoFormatter) LinkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "/client_reference/" + strings.ToLower(base) + "/"
}
