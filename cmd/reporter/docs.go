package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
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

	return explore(o.topCmd, "docs/rst")
	// linkHandler := func(name, ref string) string {
	// 	return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
	// }
	//return doc.GenReSTTreeCustom(o.topCmd, "docs/rst", func(filename string) string { return "" }, linkHandler)
}

func explore(cmd *cobra.Command, dir string) error {
	if len(cmd.Commands()) == 0 && (cmd.Name() == "k8s" || cmd.Name() == "ecs") {
		basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".rst"
		filename := filepath.Join(dir, basename)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		var buffer strings.Builder
		buffer.WriteString(fmt.Sprintf("Command name: %s\n", cmd.Name()))
		buffer.WriteString(fmt.Sprintf("Command path: %s\n", cmd.CommandPath()))
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "help" {
				buffer.WriteString(fmt.Sprintf("     Name: %s. Def Value: %s\n", f.Name, f.DefValue))
			}
		})
		fmt.Fprintf(file, "%s", buffer.String())
	} else {
		for _, c := range cmd.Commands() {
			explore(c, dir)
		}
	}
	return nil
}
