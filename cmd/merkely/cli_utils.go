package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/merkely-development/reporter/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const bitbucket = "Bitbucket"
const github = "Github"
const teamcity = "Teamcity"
const unknown = "Unknown"

// supportedCIs the set of CI tools that are supported for defaulting
var supportedCIs = []string{bitbucket, github, teamcity}

// ciTemplates a map of merkely flags and corresponding default templates in supported CI tools
var ciTemplates = map[string]map[string]string{
	github: {
		"git-commit": "${GITHUB_SHA}",
		"commit-url": "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA}",
		"build-url":  "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}",
	},
	bitbucket: {
		"git-commit": "${BITBUCKET_COMMIT}",
		"commit-url": "https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT}",
		"build-url":  "https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER}",
	},
	teamcity: {
		"git-commit": "${BUILD_VCS_NUMBER}",
	},
}

// GetCIDefaultsTemplates returns the templates used in a given CI
// to calculate the input list of keys
func GetCIDefaultsTemplates(ciTools, keys []string) string {
	result := `The following flags are defaulted as follows in the CI list below:

   `
	for _, ci := range ciTools {
		result += fmt.Sprintf(`
	| %s 
	|---------------------------------------------------------------------------`, ci)
		for _, key := range keys {
			if value, ok := ciTemplates[ci][key]; ok {
				result += fmt.Sprintf(`
	| %s : %s`, key, value)
			}
		}
		result += `
	|---------------------------------------------------------------------------`
	}
	return result
}

// WhichCI detects which CI tool we are in based on env variables
func WhichCI() string {
	if _, ok := os.LookupEnv("BITBUCKET_BUILD_NUMBER"); ok {
		return bitbucket
	} else if _, ok := os.LookupEnv("GITHUB_RUN_NUMBER"); ok {
		return github
	} else if _, ok := os.LookupEnv("TEAMCITY_VERSION"); ok {
		return teamcity
	} else {
		return unknown
	}
}

// DefaultValue looks up the default value of a given flag in a given CI tool
func DefaultValue(ci, flag string) string {
	if v, ok := ciTemplates[ci][flag]; ok {
		return os.ExpandEnv(v)
	}
	return ""
}

// RequireFlags decalres a list of flags as required for a given command
func RequireFlags(cmd *cobra.Command, flagNames []string) error {
	for _, name := range flagNames {
		if cmd.Flags().Lookup(name).DefValue == "" {
			err := cobra.MarkFlagRequired(cmd.Flags(), name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RequireGlobalFlags validates that a set of global fields have been assigned a value
func RequireGlobalFlags(global *GlobalOpts, fields []string) error {
	v := reflect.ValueOf(*global)
	typeOfGlobal := v.Type()

	for _, field := range fields {
		for i := 0; i < v.NumField(); i++ {
			if typeOfGlobal.Field(i).Name == field {
				if v.Field(i).Interface() == "" {
					return fmt.Errorf("%s is not set", GetFlagFromVarName(field))
				}
			}
		}
	}

	return nil
}

// GetFlagFromVarName returns a POSIX cmd flag from a camelCase variable name
func GetFlagFromVarName(varName string) string {
	if varName == "" {
		return ""
	}
	result := "--"
	for pos, char := range varName {
		if pos == 0 {
			result += string(unicode.ToLower(char))
			continue
		}
		if unicode.IsLetter(char) && unicode.IsUpper(char) {
			result += fmt.Sprintf("-%c", unicode.ToLower(char))
		} else {
			result += string(char)
		}
	}
	return result
}

// NoArgs returns an error if any args are included.
func NoArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf(
			"%q accepts no arguments\n\nUsage:  %s",
			cmd.CommandPath(),
			cmd.UseLine(),
		)
	}
	return nil
}

type registryProviderEndpoints struct {
	mainApi string
	authApi string
	service string
}

func getRegistryEndpointForProvider(provider string) (*registryProviderEndpoints, error) {
	switch provider {
	case "dockerhub":
		return &registryProviderEndpoints{
			mainApi: "https://registry-1.docker.io/v2",
			authApi: "https://auth.docker.io",
			service: "registry.docker.io",
		}, nil
	case "github":
		return &registryProviderEndpoints{
			mainApi: "https://ghcr.io/v2",
			authApi: "https://ghcr.io",
			service: "ghcr.io",
		}, nil

	default:
		return getRegistryEndpoint(provider)
		// return &registryProviderEndpoints{}, fmt.Errorf("%s is not a supported docker registry provider", provider)
	}

}

func getRegistryEndpoint(url string) (*registryProviderEndpoints, error) {
	return &registryProviderEndpoints{
		mainApi: "https://" + url + "/v2",
		authApi: "https://" + url + "/oauth2",
		service: url,
	}, nil
}

// getDockerRegistryAPIToken returns a short-lived read-only api token for a docker registry api
func getDockerRegistryAPIToken(providerInfo *registryProviderEndpoints, username, password, imageName string) (string, error) {
	url := fmt.Sprintf("%s/token?scope=repository:%s:pull&service=%s", providerInfo.authApi, imageName, providerInfo.service)
	res, err := requests.DoBasicAuthRequest([]byte{}, url, username, password, 3, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return "", fmt.Errorf("failed to create an authentication token for the docker registry: %v", err)
	}

	var responseData map[string]interface{}
	err = json.Unmarshal([]byte(res.Body), &responseData)
	if err != nil {
		return "", err
	}
	token := responseData["token"]
	if token == nil {
		token = responseData["access_token"]
	}
	return token.(string), nil
}

// GetSha256Digest calculates the sha256 digest of an artifact.
// Supported artifact types are: dir, file, docker
func GetSha256Digest(artifactName string, o *fingerprintOptions) (string, error) {
	var err error
	var fingerprint string
	switch o.artifactType {
	case "file":
		fingerprint, err = digest.FileSha256(artifactName)
	case "dir":
		fingerprint, err = digest.DirSha256(artifactName, log)
	case "docker":
		if o.registryProvider != "" {
			var providerInfo *registryProviderEndpoints
			providerInfo, err = getRegistryEndpointForProvider(o.registryProvider)
			if err != nil {
				return "", err
			}
			nameSlice := strings.Split(artifactName, ":")
			if len(nameSlice) < 2 {
				nameSlice = append(nameSlice, "latest")
			}
			token := ""
			if !strings.Contains(providerInfo.mainApi, "jfrog.io") {
				token, err = getDockerRegistryAPIToken(providerInfo, o.registryUsername, o.registryPassword, nameSlice[0])
				if err != nil {
					return "", err
				}
			} else {
				token = o.registryPassword
			}
			fingerprint, err = digest.RemoteDockerImageSha256(nameSlice[0], nameSlice[1], providerInfo.mainApi, token)

		} else {
			fingerprint, err = digest.DockerImageSha256(artifactName)
		}
	default:
		return "", fmt.Errorf("%s is not a supported artifact type", o.artifactType)
	}

	return fingerprint, err
}

// LoadUserData reads a user data file and validates that it contains JSON
func LoadUserData(filepath string) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	content := `{}`
	if filepath != "" {
		content, err = utils.LoadFileContent(filepath)
		if err != nil {
			return result, err
		}
		if !utils.IsJSON(content) {
			return result, fmt.Errorf("%s does not contain a valid JSON", filepath)
		}
	}
	err = json.Unmarshal([]byte(content), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// ValidateArtifactArg validates the artifact name or path argument
func ValidateArtifactArg(args []string, artifactType, inputSha256 string) error {
	if len(args) > 1 {
		return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
	}
	if len(args) == 0 || args[0] == "" {
		return fmt.Errorf("docker image name or file/dir path is required")
	}

	if artifactType == "" && inputSha256 == "" {
		return fmt.Errorf("either --type or --sha256 must be specified")
	}

	if inputSha256 != "" {
		if err := digest.ValidateDigest(inputSha256); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRegisteryFlags validates that you provide all registery information necessary for
// remote digest.
func ValidateRegisteryFlags(o *fingerprintOptions) error {
	if o.registryProvider != "" && (o.registryPassword == "" || o.registryUsername == "") {
		return fmt.Errorf("both --registry-username and registry-password are required when --registry-provider is used")
	}
	if o.registryProvider == "" && (o.registryPassword != "" || o.registryUsername != "") {
		return fmt.Errorf("--registry-username and registry-password are only used when --registry-provider is used")
	}
	return nil
}
