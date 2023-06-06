package main

import (
	"bytes"
	"fmt"
	"io"
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

		hdrFunc := func(filename string, beta bool) string {
			base := filepath.Base(filename)
			name := strings.TrimSuffix(base, path.Ext(base))
			title := strings.ToLower(strings.Replace(name, "_", " ", -1))
			return fmt.Sprintf("---\ntitle: \"%s\"\nbeta: %t\n---\n\n", title, beta)
		}

		return MereklyGenMarkdownTreeCustom(o.topCmd, o.dest, hdrFunc, linkHandler)
	}
	return doc.GenMarkdownTree(o.topCmd, o.dest)
}

func MereklyGenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender func(string, bool) string, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
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

		if _, err := io.WriteString(f, filePrepender(filename, isBeta(cmd))); err != nil {
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
		buf.WriteString(fmt.Sprintf("**%s** is a beta feature. \n", name))
		buf.WriteString("Beta features provide early access to product functionality. These ")
		buf.WriteString("features may change between releases without warning, or can be removed from a ")
		buf.WriteString("future release.\n")
		buf.WriteString("You can enable beta features by using the `kosli enable beta` command.")
		buf.WriteString("{{< /hint >}}\n")
	}

	if len(cmd.Long) > 0 {
		buf.WriteString("## Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.UseLine()))
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.Example))
	}
	_, err := buf.WriteTo(w)
	return err
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
		buf.WriteString("## Options inherited from parent commands\n")
		buf.WriteString("| Flag | Description |\n")
		buf.WriteString("| :--- | :--- |\n")
		usages := CommandsInTable(parentFlags)
		fmt.Fprint(buf, usages)
		buf.WriteString("\n\n")
	}
	return nil
}
