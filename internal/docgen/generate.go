package docgen

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// GenMarkdownTree walks the cobra command tree and generates a doc file for each
// leaf command using the provided Formatter.
func GenMarkdownTree(cmd *cobra.Command, dir string, formatter Formatter, metaFn CommandMetaFunc, liveDocs LiveDocProvider) error {
	for _, c := range cmd.Commands() {
		// skip all unavailable commands except deprecated ones
		if (!c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand()) && c.Deprecated == "" {
			continue
		}
		if err := GenMarkdownTree(c, dir, formatter, metaFn, liveDocs); err != nil {
			return err
		}
	}

	// Only generate docs for leaf commands (not root, not parent commands)
	if !cmd.HasParent() || !cmd.HasSubCommands() {
		basename := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
		filename := filepath.Join(dir, basename)
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("warning: failed to close file %s: %v", filename, err)
			}
		}()

		if err := genMarkdownCustom(cmd, f, formatter, metaFn, liveDocs); err != nil {
			return err
		}
	}
	return nil
}

func genMarkdownCustom(cmd *cobra.Command, w io.Writer, formatter Formatter, metaFn CommandMetaFunc, liveDocs LiveDocProvider) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	meta := metaFn(cmd)
	name := meta.Name

	var buf strings.Builder

	// Front matter
	buf.WriteString(formatter.FrontMatter(meta))

	// Title
	buf.WriteString(formatter.Title(name))

	// Beta warning
	if meta.Beta {
		buf.WriteString(formatter.BetaWarning(name))
	}

	// Deprecated warning
	if meta.Deprecated {
		buf.WriteString(formatter.DeprecatedWarning(name, meta.DeprecMsg))
	}

	// Synopsis
	buf.WriteString(formatter.Synopsis(meta))

	// Flags
	flags, inherited := RenderFlagsTables(cmd)
	buf.WriteString(formatter.FlagsSection(flags, inherited))

	// Live CI examples
	urlSafeName := url.QueryEscape(name)
	var ciExamples []CIExample
	for _, ci := range []string{"GitHub", "GitLab"} {
		if liveDocs.YamlDocExists(ci, urlSafeName) {
			ex := CIExample{
				CI:      ci,
				YamlURL: liveDocs.YamlURL(ci, urlSafeName),
			}
			if liveDocs.EventDocExists(ci, urlSafeName) {
				ex.EventURL = liveDocs.EventURL(ci, urlSafeName)
			}
			ciExamples = append(ciExamples, ex)
		}
	}
	buf.WriteString(formatter.LiveCIExamples(ciExamples, name))

	// Live CLI example
	fullCommand, cliURL, cliExists := liveDocs.CLIDocExists(name)
	if cliExists {
		buf.WriteString(formatter.LiveCLIExample(name, fullCommand, cliURL))
	}

	// Example use cases
	if len(meta.Example) > 0 {
		buf.WriteString(formatter.ExampleUseCases(name, meta.Example))
	}

	_, err := fmt.Fprint(w, buf.String())
	return err
}
