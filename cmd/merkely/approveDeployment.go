package main

import (
	"fmt"
	"io"
	"net/http"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type approvalOptions struct {
	artifactType    string
	pipelineName    string
	oldestSrcCommit string
	newestSrcCommit string
	srcRepoRoot     string
	userDataFile    string
	payload         ApprovalPayload
}

type ApprovalPayload struct {
	ArtifactSha256 string                 `json:"artifact_sha256"`
	Description    string                 `json:"description"`
	CommitList     []string               `json:"src_commit_list"`
	Reviews        []map[string]string    `json:"approvals"`
	UserData       map[string]interface{} `json:"user_data"`
}

func newApproveDeploymentCmd(out io.Writer) *cobra.Command {
	o := new(approvalOptions)
	cmd := &cobra.Command{
		Use:   "approval ARTIFACT-NAME-OR-PATH",
		Short: "Approve deploying an artifact in Merkely. ",
		Long:  approveDeploymentDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return ValidateArtifactArg(args, o.artifactType, o.payload.ArtifactSha256)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, false)
		},
	}

	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact to be approved. Options are [dir, file, docker]. Only required if you don't specify --sha256.")
	cmd.Flags().StringVarP(&o.payload.ArtifactSha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", "[optional] The approval description.")
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", "[optional] The path to a JSON file containing additional data you would like to attach to this approval.")
	cmd.Flags().StringVar(&o.oldestSrcCommit, "oldest-commit", "", "The source commit sha for the oldest change in the deployment approval.")
	cmd.Flags().StringVar(&o.newestSrcCommit, "newest-commit", "HEAD", "The source commit sha for the newest change in the deployment approval.")
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", "The directory where the source git repository is volume-mounted.")

	err := RequireFlags(cmd, []string{"pipeline", "oldest-commit"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *approvalOptions) run(args []string, request bool) error {
	var err error
	if o.payload.ArtifactSha256 == "" {
		o.payload.ArtifactSha256, err = GetSha256Digest(o.artifactType, args[0])
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/approvals/", global.Host, global.Owner, o.pipelineName)
	if !request {
		o.payload.Reviews = []map[string]string{
			{
				"state":        "APPROVED",
				"comment":      o.payload.Description,
				"approved_by":  "External",
				"approval_url": "undefined",
			},
		}
	} else {
		o.payload.Reviews = []map[string]string{}
	}

	o.payload.UserData, err = LoadUserData(o.userDataFile)
	if err != nil {
		return err
	}
	o.payload.CommitList, err = listCommitsBetween(o.srcRepoRoot, o.oldestSrcCommit, o.newestSrcCommit)
	if err != nil {
		return err
	}

	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPost, log)
	return err
}

func approveDeploymentDesc() string {
	return `
   Approve a deployment of an artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   `
}

// listCommitsBetween list all commits that have happened between two commits in a git repo
func listCommitsBetween(repoRoot, oldest, newest string) ([]string, error) {
	newestHash := plumbing.NewHash(newest)
	oldestHash := plumbing.NewHash(oldest)
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return []string{}, err
	}

	commits := []string{}

	commitsIter, err := repo.Log(&git.LogOptions{From: newestHash, Order: git.LogOrderCommitterTime})
	if err != nil {
		return []string{}, err
	}

	for ok := true; ok; {
		commit, err := commitsIter.Next()
		if err != nil {
			return []string{}, err
		}
		if commit.Hash != oldestHash {
			commits = append(commits, commit.Hash.String())
		} else {
			break
		}
	}

	return commits, nil
}
