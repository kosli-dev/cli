package main

import (
	"io"

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
		formatter := docgen.MintlifyFormatter{}

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

		return docgen.GenMarkdownTree(o.topCmd, o.dest, formatter, metaFn)
	}
	return doc.GenMarkdownTree(o.topCmd, o.dest)
}

