package docgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenMarkdownTreeCreatesFiles(t *testing.T) {
	dir := t.TempDir()

	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{
		Use:   "child",
		Short: "A child command",
		Long:  "A child command with a longer description.",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(child)

	metaFn := func(cmd *cobra.Command) CommandMeta {
		return CommandMeta{
			Name:     cmd.CommandPath(),
			Summary:  cmd.Short,
			Long:     cmd.Long,
			UseLine:  cmd.UseLine(),
			Runnable: cmd.Runnable(),
			Example:  cmd.Example,
		}
	}

	err := GenMarkdownTree(root, dir, HugoFormatter{}, metaFn, NullLiveDocProvider{})
	if err != nil {
		t.Fatalf("GenMarkdownTree error: %v", err)
	}

	// Should create a file for the leaf command
	expected := filepath.Join(dir, "root_child.md")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", expected)
	}

	content, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), "# root child") {
		t.Error("expected command name as heading")
	}
	if !strings.Contains(string(content), "## Synopsis") {
		t.Error("expected Synopsis section")
	}
}

func TestGenMarkdownTreeSkipsHiddenCommands(t *testing.T) {
	dir := t.TempDir()

	root := &cobra.Command{Use: "root"}
	hidden := &cobra.Command{
		Use:    "hidden",
		Hidden: true,
		RunE:   func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(hidden)

	metaFn := func(cmd *cobra.Command) CommandMeta {
		return CommandMeta{Name: cmd.CommandPath()}
	}

	err := GenMarkdownTree(root, dir, HugoFormatter{}, metaFn, NullLiveDocProvider{})
	if err != nil {
		t.Fatalf("GenMarkdownTree error: %v", err)
	}

	unexpected := filepath.Join(dir, "root_hidden.md")
	if _, err := os.Stat(unexpected); !os.IsNotExist(err) {
		t.Error("expected hidden command to be skipped")
	}
}

func TestGenMarkdownTreeIncludesDeprecatedCommands(t *testing.T) {
	dir := t.TempDir()

	root := &cobra.Command{Use: "root"}
	deprecated := &cobra.Command{
		Use:        "deprecated",
		Short:      "A deprecated cmd",
		Long:       "Deprecated long desc.",
		Deprecated: "use something else",
		RunE:       func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(deprecated)

	metaFn := func(cmd *cobra.Command) CommandMeta {
		return CommandMeta{
			Name:       cmd.CommandPath(),
			Deprecated: cmd.Deprecated != "",
			DeprecMsg:  cmd.Deprecated,
			Summary:    cmd.Short,
			Long:       cmd.Long,
			UseLine:    cmd.UseLine(),
			Runnable:   cmd.Runnable(),
		}
	}

	err := GenMarkdownTree(root, dir, HugoFormatter{}, metaFn, NullLiveDocProvider{})
	if err != nil {
		t.Fatalf("GenMarkdownTree error: %v", err)
	}

	expected := filepath.Join(dir, "root_deprecated.md")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Error("expected deprecated command to be included")
	}

	content, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), "hint danger") {
		t.Error("expected deprecation warning in output")
	}
}

func TestGenMarkdownTreeWithMintlifyFormatter(t *testing.T) {
	dir := t.TempDir()

	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{
		Use:   "child",
		Short: "A child command",
		Long:  "A child command with a longer description.",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(child)

	metaFn := func(cmd *cobra.Command) CommandMeta {
		return CommandMeta{
			Name:     cmd.CommandPath(),
			Summary:  cmd.Short,
			Long:     cmd.Long,
			UseLine:  cmd.UseLine(),
			Runnable: cmd.Runnable(),
		}
	}

	err := GenMarkdownTree(root, dir, MintlifyFormatter{}, metaFn, NullLiveDocProvider{})
	if err != nil {
		t.Fatalf("GenMarkdownTree error: %v", err)
	}

	expected := filepath.Join(dir, "root_child.md")
	content, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "description:") {
		t.Error("expected Mintlify description in front matter")
	}
}
