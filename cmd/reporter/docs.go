package main

import (
	"fmt"
	"io"
	//"path"
	//"path/filepath"
	//"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsDesc = `
Generate documentation files for Merkely CLI.
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
		Short:  "generate documentation as markdown",
		Long:   docsDesc,
		Hidden: true,
		Args:   NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.topCmd = cmd.Root()
			return o.run(out)
		},
	}

	f := cmd.Flags()
	f.StringVar(&o.dest, "dir", "./", "directory to which documentation is written")
	f.BoolVar(&o.generateHeaders, "generate-headers", true, "generate standard headers for markdown files")

	return cmd
}

func (o *docsOptions) run(out io.Writer) error {
	// if o.generateHeaders {
	// 	standardLinks := func(s string) string { return s }

	// 	hdrFunc := func(filename string) string {
	// 		base := filepath.Base(filename)
	// 		name := strings.TrimSuffix(base, path.Ext(base))
	// 		title := strings.Title(strings.Replace(name, "_", " ", -1))
	// 		return fmt.Sprintf("---\ntitle: \"%s\"\n---\n\n", title)
	// 	}

	// 	return doc.GenMarkdownTreeCustom(o.topCmd, o.dest, hdrFunc, standardLinks)
	// }
	var err = doc.GenMarkdownTree(o.topCmd, o.dest)
	if err != nil {
		return err
	}
	linkHandler := func(name, ref string) string {
		return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
	}

	return doc.GenReSTTreeCustom(o.topCmd, "docs/rst", func(filename string) string { return "" }, linkHandler)
}
