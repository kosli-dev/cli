package docgen

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// MintlifyFormatter generates Mintlify-compatible MDX documentation.
type MintlifyFormatter struct{}

func (MintlifyFormatter) Title(name string) string {
	return "" // Mintlify renders title from front matter
}

func (MintlifyFormatter) FrontMatter(meta CommandMeta) string {
	desc := sanitizeDescription(meta.Summary)
	return fmt.Sprintf("---\ntitle: \"%s\"\nbeta: %t\ndeprecated: %t\ndescription: \"%s\"\n---\n\n",
		meta.Name, meta.Beta, meta.Deprecated, desc)
}

func (MintlifyFormatter) BetaWarning(name string) string {
	var b strings.Builder
	b.WriteString("<Warning>\n")
	fmt.Fprintf(&b, "**%s** is a beta feature. ", name)
	fmt.Fprintf(&b, "Beta features provide early access to product functionality. ")
	fmt.Fprintf(&b, "These features may change between releases without warning, or can be removed in a ")
	fmt.Fprintf(&b, "future release.\n")
	fmt.Fprintf(&b, "Please contact us to enable this feature for your organization.\n")
	b.WriteString("</Warning>\n")
	return b.String()
}

func (MintlifyFormatter) DeprecatedWarning(name, message string) string {
	var b strings.Builder
	b.WriteString("<Warning>\n")
	fmt.Fprintf(&b, "**%s** is deprecated. %s  ", name, message)
	fmt.Fprintf(&b, "Deprecated commands will be removed in a future release.\n")
	b.WriteString("</Warning>\n")
	return b.String()
}

func (MintlifyFormatter) Synopsis(meta CommandMeta) string {
	var b strings.Builder
	if len(meta.Long) > 0 {
		b.WriteString("## Synopsis\n\n")
		if meta.Runnable {
			fmt.Fprintf(&b, "```shell\n%s\n```\n\n", meta.UseLine)
		}
		long := strings.ReplaceAll(meta.Long, "^", "`")
		b.WriteString(escapeMintlifyProse(long) + "\n\n")
	}
	return b.String()
}

func (MintlifyFormatter) FlagsSection(flags, inherited string) string {
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

func (MintlifyFormatter) LiveCIExamples(examples []CIExample, commandName string) string {
	if len(examples) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Live Examples in different CI systems\n\n")
	b.WriteString("<Tabs>\n")
	for _, ex := range examples {
		fmt.Fprintf(&b, "\t<Tab title=\"%v\">\n", ex.CI)
		fmt.Fprintf(&b, "\tView an example of the `%s` command in %s.\n\n", commandName, ex.CI)
		fmt.Fprintf(&b, "\tIn [this YAML file](%v)", ex.YamlURL)
		if ex.EventURL != "" {
			fmt.Fprintf(&b, ", which created [this Kosli Event](%v).", ex.EventURL)
		}
		b.WriteString("\n\t</Tab>\n")
	}
	b.WriteString("</Tabs>\n\n")
	return b.String()
}

func (MintlifyFormatter) LiveCLIExample(commandName, fullCommand, url string) string {
	var b strings.Builder
	b.WriteString("## Live Example\n\n")
	fmt.Fprintf(&b, "To view a live example of '%s' you can run the commands below (for the <a href=\"https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/\">cyber-dojo</a> demo organization).<br/><a href=\"%s\">Run the commands below and view the output.</a>", commandName, url)
	b.WriteString("<pre>")
	b.WriteString("export KOSLI_ORG=cyber-dojo\n")
	b.WriteString("export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only\n")
	b.WriteString(fullCommand)
	b.WriteString("</pre>\n\n")
	return b.String()
}

func (MintlifyFormatter) ExampleUseCases(commandName, example string) string {
	var b strings.Builder
	b.WriteString("## Examples Use Cases\n\n")
	url := "/getting_started/install/#assigning-flags-via-environment-variables"
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
		b.WriteString("<AccordionGroup>\n")
		for i := 0; i < len(all); i++ {
			exampleLines := all[i]
			title := strings.Trim(exampleLines[0], ":")
			if len(title) > 0 {
				fmt.Fprintf(&b, "<Accordion title=\"%s\">\n", strings.TrimSpace(title[1:]))
				fmt.Fprintf(&b, "```shell\n%s\n```\n", strings.Join(exampleLines[1:], "\n"))
				b.WriteString("</Accordion>\n")
			}
		}
		b.WriteString("</AccordionGroup>\n\n")
	}
	return b.String()
}

func (MintlifyFormatter) LinkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "/client_reference/" + strings.ToLower(base)
}

// sanitizeDescription replaces ^text^ with `text` and truncates to 200 chars.
func sanitizeDescription(s string) string {
	s = strings.ReplaceAll(s, "^", "`")
	s = strings.ReplaceAll(s, "\"", "'")
	if len(s) > 200 {
		s = s[:197] + "..."
	}
	return s
}

// escapeMintlifyProse escapes JSX-problematic characters in prose text
// (outside of code fences). It converts {expr} to \{expr\} and <WORD> to `WORD`.
func escapeMintlifyProse(s string) string {
	// Split on code fences to only escape prose sections
	parts := strings.Split(s, "```")
	for i := 0; i < len(parts); i += 2 {
		// Only process prose sections (even indices)
		if i < len(parts) {
			parts[i] = escapeProseFragment(parts[i])
		}
	}
	return strings.Join(parts, "```")
}

var angleBracketPattern = regexp.MustCompile(`<([A-Z][A-Z0-9_-]*)>`)

func escapeProseFragment(s string) string {
	// Escape curly braces: {expr} -> \{expr\}
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")

	// Escape <UPPERCASE_WORD> patterns -> `UPPERCASE_WORD`
	// but leave HTML tags like <a>, <br/>, <pre> alone
	s = angleBracketPattern.ReplaceAllString(s, "`$1`")

	return s
}
