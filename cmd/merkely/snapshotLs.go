package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const snapshotLsDesc = `
List snapshot.
`

type snapshotLsOptions struct {
	// long bool
}

func newSnapshotLsCmd(out io.Writer) *cobra.Command {
	o := new(snapshotLsOptions)
	cmd := &cobra.Command{
		Use:   "snap",
		Short: snapshotLsDesc,
		Long:  snapshotLsDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	// cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)

	return cmd
}

func (o *snapshotLsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/snapshots/-1", global.Host, global.Owner, args[0])
	var outErr error
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		// if o.assert {
		// 	return fmt.Errorf("merkely server %s is unresponsive", global.Host)
		// }
		_, outErr = out.Write([]byte(err.Error()))
	} else {
		type Annotation struct {
			Type string `json:"type"`
			Was  int
			Now  int
		}

		type Owner struct {
			ApiVersion         string
			Kind               string
			Name               string
			Uid                string
			Controller         bool
			BlockOwnerDeletion bool
		}

		type PodContent struct {
			Namespace         string
			CreationTimestamp int64
			Owners            []Owner
		}

		type Artifact struct {
			Name              string
			Pipeline_name     string
			Compliant         bool
			Deployments       []int
			Sha256            string
			CreationTimestamp []int64
			Pods              map[string]PodContent
			Annotation        Annotation
		}

		type Snapshot struct {
			Index     int
			Timestamp float32
			User_id   string
			User_name string
			Artifacts []Artifact
			Compliant bool
		}

		var snapshot Snapshot
		fmt.Println(snapshot)
		err = json.Unmarshal([]byte(response.Body), &snapshot)
		if err != nil {
			return err
		}

		// fmt.Println(response.Body)
		formatStringHead := "%-8s %-30s %-10s %-19s %-26s %-10s\n"
		formatStringLine := "%-8s %-30s %-10s %-19s %-26s %-10d\n"
		fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "TAG", "SHA256", "SINCE", "REPLICAS")

		for _, artifact := range snapshot.Artifacts {
			since := time.Unix(artifact.CreationTimestamp[0], 0).Format(time.RFC3339)
			artifactSplit := strings.Split(artifact.Name, ":")
			shortSha := artifact.Sha256[:7] + "..." + artifact.Sha256[64-7:]
			fmt.Printf(formatStringLine, "xxxx", artifactSplit[0], artifactSplit[1], shortSha, since, len(artifact.CreationTimestamp))
		}

		// if o.long {
		// 	fmt.Printf("%-15s %-10s %-27s %s\n", "NAME", "TYPE", "LAST REPORT", "LAST MODIFIED")
		// }

	}
	if outErr != nil {
		return outErr
	}
	return nil
}
