package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createEnvironmentShortDesc = `Create or update a Kosli environment.`

const createEnvironmentLongDesc = createEnvironmentShortDesc + `

^^--type^^ must match the type of environment you wish to record snapshots from.
The following types are supported:
  - k8s        - Kubernetes
  - ecs        - Amazon Elastic Container Service
  - s3         - Amazon S3 object storage
  - lambda     - AWS Lambda serverless
  - docker     - Docker images
  - azure-apps - Azure app services
  - server     - Generic type

By default, the environment does not require artifacts provenance (i.e. environment snapshots will not 
become non-compliant because of artifacts that do not have provenance). You can require provenance for all artifacts
by setting --require-provenance=true

Also, by default, kosli will not make new snapshots for scaling events (change in number of instances running).
For large clusters the scaling events will often outnumber the actual change of SW.

It is possible to enable new snapshots for scaling events with the --include-scaling flag, or turn
it off again with the --exclude-scaling.
`

const createEnvironmentExample = `
# create a Kosli environment:
kosli create environment yourEnvironmentName
	--type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type createEnvOptions struct {
	payload        CreateEnvironmentPayload
	excludeScaling bool
	includeScaling bool
}

type CreateEnvironmentPayload struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	Description       string `json:"description"`
	IncludeScaling    *bool  `json:"include_scaling,omitempty"`
	RequireProvenance bool   `json:"require_provenance"`
}

func newCreateEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(createEnvOptions)
	cmd := &cobra.Command{
		Use:     "environment ENVIRONMENT-NAME",
		Aliases: []string{"env"},
		Short:   createEnvironmentShortDesc,
		Long:    createEnvironmentLongDesc,
		Example: createEnvironmentExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			err = MuXRequiredFlags(cmd, []string{"exclude-scaling", "include-scaling"}, false)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.payload.Type, "type", "t", "", newEnvTypeFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", envDescriptionFlag)
	cmd.Flags().BoolVar(&o.excludeScaling, "exclude-scaling", false, excludeScalingFlag)
	cmd.Flags().BoolVar(&o.includeScaling, "include-scaling", false, includeScalingFlag)
	cmd.Flags().BoolVar(&o.payload.RequireProvenance, "require-provenance", false, requireProvenanceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createEnvOptions) run(args []string) error {
	o.payload.Name = args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s", global.Host, global.Org)

	if o.includeScaling {
		var myTrue = true
		o.payload.IncludeScaling = &myTrue
	}
	if o.excludeScaling {
		var myFalse = false
		o.payload.IncludeScaling = &myFalse
	}
	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("environment %s was created", o.payload.Name)
	}
	return err
}
