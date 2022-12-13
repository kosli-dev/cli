package main

import (
	"io"

	"github.com/kosli-dev/cli/internal/gitview"

	"github.com/spf13/cobra"
)

const commitsDesc = `Print a list of commits within a range.`

type commitsOptions struct {
	oldestSrcCommit string
	newestSrcCommit string
}

func newCommitsCmd(out io.Writer) *cobra.Command {
	o := new(commitsOptions)
	cmd := &cobra.Command{
		Use:    "commits",
		Short:  commitsDesc,
		Long:   commitsDesc,
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, out)
		},
	}

	cmd.Flags().StringVarP(&o.oldestSrcCommit, "oldest-commit", "o", "", oldestCommitFlag)
	cmd.Flags().StringVarP(&o.newestSrcCommit, "newest-commit", "n", "HEAD", newestCommitFlag)

	err := RequireFlags(cmd, []string{"oldest-commit"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}
	return cmd
}

//goland:noinspection GoUnusedParameter
func (o *commitsOptions) run(args []string, out io.Writer) error {

	gitView, err := gitview.New(".")
	if err != nil {
		return err
	}

	commits, err := gitView.CommitsBetween(o.oldestSrcCommit, o.newestSrcCommit, logger)
	if err != nil {
		return err
	}
	for _, commit := range commits {
		logger.Info(commit.Sha1)
		logger.Info("%s %s %s %d", commit.Branch, commit.Author, commit.Message, commit.Timestamp)
	}
	return nil
}
