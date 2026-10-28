package main

import (
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kosli-dev/cli/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const maxCliFlagsHeaderBytes = 1024

// invocationID is shared by every request this process makes, so the server
// can group the requests of one CLI run.
var invocationID = uuid.NewString()

// connectionFlags say where to send a request rather than what the command
// does, and several of them carry credentials or customer hostnames.
var connectionFlags = map[string]bool{
	"api-token":   true,
	"host":        true,
	"org":         true,
	"config-file": true,
	"http-proxy":  true,
}

var ciVendors = map[string]string{
	github:      "github",
	gitlab:      "gitlab",
	bitbucket:   "bitbucket",
	azureDevops: "azure-devops",
	circleci:    "circleci",
	teamcity:    "teamcity",
	codeBuild:   "codebuild",
	jenkins:     "jenkins",
}

// usageHeaders describes the running command to the Kosli server. Values never
// include flag values or positional arguments, since many flags carry secrets.
func usageHeaders(cmd *cobra.Command) map[string]string {
	return map[string]string{
		"X-Kosli-Client":        "cli/" + version.GetVersion(),
		"X-Kosli-Operation":     strings.TrimPrefix(strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()), " "),
		"X-Kosli-Invocation-Id": invocationID,
		"X-Kosli-Cli-Flags":     flagNamesInEffect(cmd),
		"X-Kosli-Ci":            ciVendor(),
	}
}

// flagNamesInEffect lists the flags given on the command line, in the
// environment or in the config file: bindFlags sets the latter two on the flag,
// which marks them changed. The list is cut at a flag boundary to fit
// maxCliFlagsHeaderBytes.
func flagNamesInEffect(cmd *cobra.Command) string {
	var names []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !connectionFlags[f.Name] {
			names = append(names, f.Name)
		}
	})
	sort.Strings(names)

	value := ""
	for _, name := range names {
		next := name
		if value != "" {
			next = value + "," + name
		}
		if len(next) > maxCliFlagsHeaderBytes {
			break
		}
		value = next
	}
	return value
}

func ciVendor() string {
	if vendor, ok := ciVendors[WhichCI()]; ok {
		return vendor
	}
	if isCI, _ := strconv.ParseBool(os.Getenv("CI")); isCI {
		return "other-ci"
	}
	return "none"
}
