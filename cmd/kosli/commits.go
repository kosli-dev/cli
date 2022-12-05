package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const commitsDesc = `
Print a list of commits within a range.
`

type commitsOptions struct {
	oldestSrcCommit string
	newestSrcCommit string
}

func newCommitsCmd(out io.Writer) *cobra.Command {
	o := new(commitsOptions)
	cmd := &cobra.Command{
		Use:    "commits",
		Short:  "Print the a list of commits between two commits.",
		Long:   commitsDesc,
		Hidden: true,
		Args:   NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, out)
		},
	}

	cmd.Flags().StringVarP(&o.oldestSrcCommit, "oldest-commit", "o", "", oldestCommitFlag)
	cmd.Flags().StringVarP(&o.newestSrcCommit, "newest-commit", "n", "HEAD", newestCommitFlag)

	err := RequireFlags(cmd, []string{"oldest-commit"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}
	return cmd
}

func (o *commitsOptions) run(args []string, out io.Writer) error {

	gitRepository, err := gitRepository(".")
	if err != nil {
		return err
	}

	commits, err := listCommitsBetween(gitRepository, o.oldestSrcCommit, o.newestSrcCommit)
	if err != nil {
		return err
	}
	for _, commit := range commits {
		fmt.Fprintf(out, "%s\n", commit.Sha1)
		fmt.Fprintf(out, "%s %s %s %d\n", commit.Branch, commit.Author, commit.Message, commit.Timestamp)
		fmt.Fprint(out, "\n")
	}
	return nil
}
