package main

import (
	"fmt"
	"io"

	"github.com/kosli-dev/cli/internal/version"
	"github.com/spf13/cobra"
)

const versionShortDesc = `Print the version of a Kosli CLI.  `
const versionLongDesc = versionShortDesc + `
The output will look something like this:  
version.BuildInfo{Version:"v0.0.1", GitCommit:"fe51cd1e31e6a202cba7dead9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.16.3"}

- Version is the semantic version of the release.
- GitCommit is the SHA for the commit that this version was built from.
- GitTreeState is "clean" if there are no local code changes when this binary was
  built, and "dirty" if the binary was built from locally modified code.
- GoVersion is the version of Go that was used to compile Kosli CLI.
`

type versionOptions struct {
	short bool
}

func newVersionCmd(out io.Writer) *cobra.Command {
	o := new(versionOptions)
	cmd := &cobra.Command{
		Use:   "version",
		Short: versionShortDesc,
		Long:  versionLongDesc,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.run(out)
		},
	}

	cmd.Flags().BoolVarP(&o.short, "short", "s", false, shortFlag)

	return cmd
}

func (o *versionOptions) run(out io.Writer) {
	logger.Info(formatVersion(o.short))
}

func formatVersion(short bool) string {
	if short {
		return version.GetVersion()
	}
	return fmt.Sprintf("%#v", version.Get())
}
