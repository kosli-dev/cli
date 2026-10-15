package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/kosli"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/spf13/cobra"
)

const getArtifactShortDesc = `Get artifact from a specified flow`

const getArtifactLongDesc = getArtifactShortDesc + `
You can get an artifact by its fingerprint or by its git commit sha.
In case of using the git commit, it is possible to get multiple artifacts matching the git commit.

The expected argument is an expression to specify the artifact to get.
It has the format <FLOW_NAME><SEPARATOR><COMMIT_SHA1|ARTIFACT_FINGERPRINT> 

Expression can be specified as follows:
- flowName@<fingerprint>  artifact with a given fingerprint. The fingerprint can be short or complete.
- flowName:<commit_sha>   artifact with a given commit SHA. The commit sha can be short or complete.

Examples of valid expressions are:
- flow@184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0
- flow@184c7
- flow:110d048bf1fce72ba546cbafc4427fb21b958dee
- flow:110d0
`

const getArtifactExample = `
# get an artifact with a given fingerprint from a flow
kosli get artifact flowName@fingerprint \
	--api-token yourAPIToken \
	--org orgName

# get the latest artifact with a given fingerprint from a flow in a specific trail
kosli get artifact flowName@fingerprint \
	--api-token yourAPIToken \
	--org orgName
	--trail trailName

# get an artifact with a given commit SHA from a flow
kosli get artifact flowName:commitSHA \
	--api-token yourAPIToken \
	--org orgName

# get a list of artifacts with a given commit SHA from a flow in a particular trail
kosli get artifact flowName:commitSHA \
	--api-token yourAPIToken \
	--org orgName
	--trail trailName`

type getArtifactOptions struct {
	output string
	trail  string
}

func newGetArtifactCmd(out io.Writer) *cobra.Command {
	o := new(getArtifactOptions)
	cmd := &cobra.Command{
		Use:     "artifact EXPRESSION",
		Short:   getArtifactShortDesc,
		Long:    getArtifactLongDesc,
		Example: getArtifactExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd.Context(), out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().StringVarP(&o.trail, "trail", "t", "", trailNameFlagOptional)

	return cmd
}

func (o *getArtifactOptions) run(ctx context.Context, out io.Writer, args []string) error {
	flowName, id, separator, err := handleArtifactExpression(args[0])
	if err != nil {
		return err
	}

	var raw []byte
	var artifacts func() ([]kosli.Artifact, error)
	if separator == "@" {
		artifact, err := kosliAPI().GetArtifactByFingerprint(ctx, flowName, id, o.trail)
		if err != nil {
			return err
		}
		raw = artifact.Raw
		artifacts = func() ([]kosli.Artifact, error) {
			value, err := artifact.Value()
			return []kosli.Artifact{value}, err
		}
	} else {
		listed, err := kosliAPI().ListArtifactsByCommitSHA(ctx, flowName, id, o.trail)
		if err != nil {
			return err
		}
		raw, artifacts = listed.Raw, listed.Value
	}

	return output.FormattedPrint(string(raw), o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": func(_ string, out io.Writer, pageNumber int) error {
				value, err := artifacts()
				if err != nil {
					return err
				}
				return printArtifactsAsTable(value, out, pageNumber)
			},
			"json": output.PrintJson,
		})
}

func printArtifactsAsTable(artifacts []kosli.Artifact, out io.Writer, pageNumber int) error {
	separator := ""
	for _, artifact := range artifacts {
		rows, err := artifactTableRows(artifact)
		if err != nil {
			return err
		}
		fmt.Print(separator)
		separator = "\n"
		tabFormattedPrint(out, []string{}, rows)
	}
	return nil
}

func artifactTableRows(artifact kosli.Artifact) ([]string, error) {
	rows := []string{
		fmt.Sprintf("Name:\t%s", artifact.Filename),
		fmt.Sprintf("Flow:\t%s", artifact.FlowName),
	}
	// pointers, so a key sent with an empty value still prints its row
	if artifact.TrailName != nil {
		rows = append(rows, fmt.Sprintf("Trail:\t%s", *artifact.TrailName))
	}
	if artifact.TemplateReferenceName != nil {
		rows = append(rows, fmt.Sprintf("Name in template:\t%s", *artifact.TemplateReferenceName))
	}
	rows = append(rows, fmt.Sprintf("Fingerprint:\t%s", artifact.Fingerprint))
	createdAt, err := formattedTimestamp(artifact.CreatedAt, false)
	if err != nil {
		return nil, err
	}
	rows = append(rows,
		fmt.Sprintf("Created on:\t%s", createdAt),
		fmt.Sprintf("Git commit:\t%s", artifact.GitCommit),
		fmt.Sprintf("Commit URL:\t%s", artifact.CommitURL),
		fmt.Sprintf("Build URL:\t%s", artifact.BuildURL),
		fmt.Sprintf("Artifact URL:\t%s", artifact.HTMLURL),
		fmt.Sprintf("State:\t%s", artifact.State),
	)

	if len(artifact.Running) > 0 {
		names := environmentRefNames(artifact.Running)
		sort.Strings(names)
		rows = append(rows, fmt.Sprintf("Running in environments:\t%s", strings.Join(names, ", ")))
	}
	if len(artifact.Exited) > 0 {
		rows = append(rows, fmt.Sprintf("Exited from environments:\t%s", strings.Join(environmentRefNames(artifact.Exited), ", ")))
	}
	if len(artifact.History) > 0 {
		rows = append(rows, "History:")
		for _, event := range artifact.History {
			timestamp, err := formattedTimestamp(event.Timestamp, true)
			if err != nil {
				return nil, err
			}
			rows = append(rows, fmt.Sprintf("    %s\t%s", event.Event, timestamp))
		}
	}
	return rows, nil
}

// environmentRefNames renders each reference as name#snapshot, e.g. "prod#12".
func environmentRefNames(refs []kosli.EnvironmentRef) []string {
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, fmt.Sprintf("%s#%.0f", ref.EnvironmentName, ref.SnapshotIndex))
	}
	return names
}
