package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type artifactCreationOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	srcRepoRoot        string
	payload            ArtifactPayload
}

type ArtifactPayload struct {
	Sha256      string            `json:"sha256"`
	Filename    string            `json:"filename"`
	Description string            `json:"description"`
	GitCommit   string            `json:"git_commit"`
	BuildUrl    string            `json:"build_url"`
	CommitUrl   string            `json:"commit_url"`
	RepoUrl     string            `json:"repo_url"`
	CommitsList []*ArtifactCommit `json:"commits_list"`
}

type ArtifactCommit struct {
	Sha1      string `json:"sha1"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Timestamp int64  `json:"timestamp"`
	Branch    string `json:"branch"`
}

const artifactCreationExample = `
# Report to a Kosli pipeline that a file type artifact has been created
kosli pipeline artifact report creation FILE.tgz \
	--api-token yourApiToken \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
kosli pipeline artifact report creation \
	--api-token yourApiToken \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 
`

//goland:noinspection GoUnusedParameter
func newArtifactCreationCmd(out io.Writer) *cobra.Command {
	o := new(artifactCreationOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "creation ARTIFACT-NAME-OR-PATH",
		Short:   "Report an artifact creation to a Kosli pipeline. ",
		Long:    artifactCreationDesc(),
		Example: artifactCreationExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Sha256, true)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegisteryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", artifactDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.GitCommit, "git-commit", "g", DefaultValue(ci, "git-commit"), gitCommitFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), commitUrlFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"pipeline", "git-commit", "build-url", "commit-url"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *artifactCreationOptions) run(args []string) error {
	if o.payload.Sha256 != "" {
		o.payload.Filename = args[0]
	} else {
		var err error
		o.payload.Sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
		if o.fingerprintOptions.artifactType == "dir" || o.fingerprintOptions.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	previousCommit, err := previousCommit(o)
	if err != nil {
		return err
	}

	o.payload.CommitsList, err = changeLog(o, previousCommit)
	if err != nil {
		return err
	}

	o.payload.RepoUrl, err = getRepoUrl(o.srcRepoRoot)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/", global.Host, global.Owner, o.pipelineName)
	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func changeLog(o *artifactCreationOptions, previousCommit string) ([]*ArtifactCommit, error) {
	if previousCommit != "" {
		commitsList, err := listCommitsBetween(o.srcRepoRoot, previousCommit, o.payload.GitCommit)
		if err != nil {
			fmt.Printf("Warning: %s\n", err)
		}
		return commitsList, nil
	}

	if len(o.payload.CommitsList) == 0 {
		currentArtifactCommit, err := o.currentArtifactCommit()
		if err != nil {
			return []*ArtifactCommit{}, err
		}
		return []*ArtifactCommit{currentArtifactCommit}, nil
	}
	return []*ArtifactCommit{}, nil
}

func previousCommit(o *artifactCreationOptions) (string, error) {
	previousCommitUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/latest_commit",
		global.Host, global.Owner, o.pipelineName, o.payload.Sha256)

	response, err := requests.DoBasicAuthRequest([]byte{}, previousCommitUrl, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, log)
	if err != nil {
		return "", err
	}

	var previousCommitResponse map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &previousCommitResponse)
	if err != nil {
		return "", err
	}
	previousCommit := previousCommitResponse["latest_commit"]
	if previousCommit == nil {
		return "", nil
	} else {
		return previousCommit.(string), nil
	}
}

func (o *artifactCreationOptions) currentArtifactCommit() (*ArtifactCommit, error) {
	repo, err := git.PlainOpen(o.srcRepoRoot)
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("failed to open git repository at %s: %v", o.srcRepoRoot, err)
	}

	branchName, err := branchName(repo)
	if err != nil {
		return &ArtifactCommit{}, err
	}

	currentHash, err := repo.ResolveRevision(plumbing.Revision(o.payload.GitCommit))
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("failed to resolve %s: %v", o.payload.GitCommit, err)
	}
	currentCommit, err := repo.CommitObject(*currentHash)
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("could not retrieve commit for %s: %v", *currentHash, err)
	}

	return &ArtifactCommit{
		Sha1:      currentCommit.Hash.String(),
		Message:   strings.TrimSpace(currentCommit.Message),
		Author:    currentCommit.Author.String(),
		Timestamp: currentCommit.Author.When.UTC().Unix(),
		Branch:    branchName,
	}, nil
}

func artifactCreationDesc() string {
	return `
   Report an artifact creation to a Kosli pipeline. 
   ` + sha256Desc
}

func getRepoUrl(repoRoot string) (string, error) {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return "", fmt.Errorf("failed to open git repository at %s: %v",
			repoRoot, err)
	}
	repoRemote, err := repo.Remote("origin") // TODO: We hard code this for now. Should we have a flag to set it from the cmdline?
	if err != nil {
		fmt.Printf("Warning: Repo URL will not be reported since there is no remote('origin') in git repository (%s)\n", repoRoot)
		return "", nil
	}
	remoteUrl := repoRemote.Config().URLs[0]
	if strings.HasPrefix(remoteUrl, "git@") {
		remoteUrl = strings.Replace(remoteUrl, ":", "/", 1)
		remoteUrl = strings.Replace(remoteUrl, "git@", "https://", 1)
	}
	remoteUrl = strings.TrimSuffix(remoteUrl, ".git")
	return remoteUrl, nil
}

// listCommitsBetween list all commits that have happened between two commits in a git repo
func listCommitsBetween(repoRoot, oldest, newest string) ([]*ArtifactCommit, error) {
	var commits []*ArtifactCommit
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return commits, fmt.Errorf("failed to open git repository at %s: %v",
			repoRoot, err)
	}

	branchName, err := branchName(repo)
	if err != nil {
		return commits, err
	}

	newestHash, err := repo.ResolveRevision(plumbing.Revision(newest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve %s: %v", newest, err)
	}
	oldestHash, err := repo.ResolveRevision(plumbing.Revision(oldest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve %s: %v", oldest, err)
	}

	log.Debugf("This is the newest commit hash %s", newestHash.String())
	log.Debugf("This is the oldest commit hash %s", oldestHash.String())

	commitsIter, err := repo.Log(&git.LogOptions{From: *newestHash, Order: git.LogOrderCommitterTime})
	if err != nil {
		return commits, fmt.Errorf("failed to git log: %v", err)
	}

	for true {
		commit, err := commitsIter.Next()
		if err != nil {
			return commits, fmt.Errorf("failed to get next commit: %v", err)
		}
		if commit.Hash != *oldestHash {
			currentCommit := &ArtifactCommit{
				Sha1:      commit.Hash.String(),
				Message:   strings.TrimSpace(commit.Message),
				Author:    commit.Author.String(),
				Timestamp: commit.Author.When.UTC().Unix(),
				Branch:    branchName,
			}
			commits = append(commits, currentCommit)
		} else {
			break
		}
	}

	return commits, nil
}

func branchName(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get the current HEAD of the git repository: %v", err)
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "", nil
}
