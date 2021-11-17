package main

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

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
	f.StringVar(&o.dest, "dir", "./", "The directory to which documentation is written.")
	f.BoolVar(&o.generateHeaders, "generate-headers", true, "Generate standard headers for markdown files.")

	return cmd
}

func (o *docsOptions) run(out io.Writer) error {
	if o.generateHeaders {
		standardLinks := func(s string) string { return s }

		hdrFunc := func(filename string) string {
			base := filepath.Base(filename)
			name := strings.TrimSuffix(base, path.Ext(base))
			title := strings.Title(strings.Replace(name, "_", " ", -1))
			return fmt.Sprintf("---\ntitle: \"%s\"\n---\n\n", strings.ToLower(title))
		}

		return doc.GenMarkdownTreeCustom(o.topCmd, o.dest, hdrFunc, standardLinks)
	}
	return doc.GenMarkdownTree(o.topCmd, o.dest)
	// var err = doc.GenMarkdownTree(o.topCmd, o.dest)
	// if err != nil {
	// 	return err
	// }

	// return generateReSTFiles(o.topCmd, "docs/rst")
	// linkHandler := func(name, ref string) string {
	// 	return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
	// }
	//return doc.GenReSTTreeCustom(o.topCmd, "docs/rst", func(filename string) string { return "" }, linkHandler)
}

// func generateReSTFiles(cmd *cobra.Command, dir string) error {
// 	if len(cmd.Commands()) == 0 && (cmd.Name() == "k8s" || cmd.Name() == "ecs" || cmd.Name() == "server") {
// 		basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".rst"
// 		filename := filepath.Join(dir, basename)
// 		file, err := os.Create(filename)
// 		if err != nil {
// 			return err
// 		}
// 		defer file.Close()

// 		lines := []string{}

// 		lines = append(lines, fmt.Sprintf(".. list-table:: %s", cmd.CommandPath()))
// 		lines = append(lines, "   :header-rows: 1")
// 		lines = append(lines, "")
// 		lines = append(lines, "   * - ENV_VAR_NAME")
// 		lines = append(lines, "     - Required?")
// 		lines = append(lines, "     - Notes")
// 		cmd.Flags().VisitAll(func(f *pflag.Flag) {
// 			if f.Name != "help" {
// 				lines = append(lines, fmt.Sprintf("   * - %s", merkelyEnvVar(f.Name)))
// 				lines = append(lines, fmt.Sprintf("     - %s", required(f.DefValue)))
// 				lines = append(lines, fmt.Sprintf("     - %s", usage(f.Usage, f.DefValue)))
// 			}
// 		})
// 		for _, line := range lines {
// 			fmt.Fprintf(file, "%s\n", line)
// 		}
// 	} else {
// 		for _, c := range cmd.Commands() {
// 			err := generateReSTFiles(c, dir)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func merkelyEnvVar(s string) string {
// 	s = strings.Replace(s, "-", "_", -1)
// 	s = strings.ToUpper(s)
// 	return "MERKELY_" + s
// }

// func required(s string) string {
// 	if len(s) == 0 {
// 		return "yes"
// 	} else {
// 		return "no"
// 	}
// }

// func usage(usage string, def string) string {
// 	var result string
// 	result += usage
// 	if required(def) == "no" {
// 		result += " Defaults to :code:`" + def + "`."
// 	}
// 	return result
// }
