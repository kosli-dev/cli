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
		long = linkifyKosliDocsURLs(long)
		b.WriteString(escapeMintlifyProse(long) + "\n\n")
	}
	return b.String()
}

func (MintlifyFormatter) FlagsSection(flags, inherited string) string {
	flags = linkifyKosliDocsURLs(flags)
	flags = escapeMintlifyProse(flags)
	inherited = linkifyKosliDocsURLs(inherited)
	inherited = escapeMintlifyProse(inherited)
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

var kosliDocsURLPattern = regexp.MustCompile(`https://docs\.kosli\.com(/[^\s),]*)`)

// linkifyKosliDocsURLs converts bare https://docs.kosli.com/path URLs
// into markdown links [docs](/path).
func linkifyKosliDocsURLs(s string) string {
	return kosliDocsURLPattern.ReplaceAllString(s, "[docs]($1)")
}

// angleBracketPattern matches placeholder patterns in angle brackets that MDX
// would interpret as JSX tags. Matches:
//   - uppercase placeholders like <IMAGE-NAME>
//   - patterns with pipes like <hours|days|weeks|months> or <COMMIT_SHA1|FINGERPRINT>
//   - lowercase placeholders like <fingerprint>, <commit_sha>
//
// Standard HTML tags like <a>, <br/>, <pre>, <code> are filtered out in escapeProseFragment.
var angleBracketPattern = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9_|-]*)>`)

var htmlTags = map[string]bool{
	"a": true, "br": true, "pre": true, "code": true, "em": true,
	"strong": true, "p": true, "div": true, "span": true, "ul": true,
	"ol": true, "li": true, "img": true, "table": true, "tr": true,
	"td": true, "th": true, "thead": true, "tbody": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// singleQuotedURLPattern matches single-quoted URLs like 'http://example.com'
// so they can be rendered as inline code in Mintlify instead of clickable links.
// Real links should use markdown syntax [text](url) in the flag description;
// single quotes are reserved for example/placeholder URLs.
var singleQuotedURLPattern = regexp.MustCompile(`'(https?://[^\s']+)'`)

func escapeProseFragment(s string) string {
	// Escape curly braces: {expr} -> \{expr\}
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")

	// Convert single-quoted URLs to backtick-wrapped inline code
	// so Mintlify renders them as code instead of clickable links.
	s = singleQuotedURLPattern.ReplaceAllString(s, "`$1`")

	// Escape angle-bracket placeholders -> backtick-wrapped
	// but leave standard HTML tags alone
	s = angleBracketPattern.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[1 : len(match)-1]
		// Check the base tag name (before any pipe) against HTML tags
		baseName := strings.SplitN(inner, "|", 2)[0]
		if htmlTags[strings.ToLower(baseName)] {
			return match
		}
		return "`" + inner + "`"
	})

	return s
}
