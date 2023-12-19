package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlPackage "net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/gitview"
	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/utils"
	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

const (
	bitbucket   = "Bitbucket"
	github      = "Github"
	teamcity    = "Teamcity"
	gitlab      = "Gitlab"
	azureDevops = "Azure Devops"
	circleci    = "CircleCI"
	unknown     = "Unknown"
)

// supportedCIs the set of CI tools that are supported for defaulting
var supportedCIs = []string{bitbucket, github, teamcity, gitlab, azureDevops, circleci}

// ciTemplates a map of kosli flags and corresponding default templates in supported CI tools
var ciTemplates = map[string]map[string]string{
	github: {
		"git-commit": "${GITHUB_SHA}",
		"repository": "${GITHUB_REPOSITORY}",
		"org":        "${GITHUB_REPOSITORY_OWNER}",
		"commit-url": "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA}",
		"build-url":  "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}",
	},
	bitbucket: {
		"git-commit": "${BITBUCKET_COMMIT}",
		"repository": "${BITBUCKET_REPO_SLUG}",
		"workspace":  "${BITBUCKET_WORKSPACE}",
		"commit-url": "https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT}",
		"build-url":  "https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER}",
	},
	teamcity: {
		"git-commit": "${BUILD_VCS_NUMBER}",
	},
	gitlab: {
		"git-commit": "${CI_COMMIT_SHA}",
		"repository": "${CI_PROJECT_NAME}",
		"build-url":  "${CI_JOB_URL}",
		"commit-url": "${CI_PROJECT_URL}/-/commit/${CI_COMMIT_SHA}",
		"namespace":  "${CI_PROJECT_NAMESPACE}",
	},
	azureDevops: {
		"git-commit": "${BUILD_SOURCEVERSION}",
		"repository": "${BUILD_REPOSITORY_NAME}",
		"build-url":  "${SYSTEM_COLLECTIONURI}${SYSTEM_TEAMPROJECT}/_build/results?buildId=${BUILD_BUILDID}",
		"commit-url": "${SYSTEM_COLLECTIONURI}${SYSTEM_TEAMPROJECT}/_git/${BUILD_REPOSITORY_NAME}/commit/${BUILD_SOURCEVERSION}",
		"org-url":    "${SYSTEM_COLLECTIONURI}",
		"project":    "${SYSTEM_TEAMPROJECT}",
	},
	circleci: {
		"git-commit": "${CIRCLE_SHA1}",
		"repository": "${CIRCLE_PROJECT_REPONAME}",
		"commit-url": "${CIRCLE_REPOSITORY_URL}/commit/${CIRCLE_SHA1}",
		"build-url":  "${CIRCLE_BUILD_URL}",
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
	} else if _, ok := os.LookupEnv("GITLAB_CI"); ok {
		return gitlab
	} else if _, ok := os.LookupEnv("TF_BUILD"); ok {
		return azureDevops
	} else if _, ok := os.LookupEnv("CIRCLECI"); ok {
		return circleci
	} else {
		return unknown
	}
}

// DefaultValue looks up the default value of a given flag in a given CI tool
// if the DOCS env variable is set, return empty string to avoid
// having irrelevant defaults in the docs
// if the KOSLI_TESTS env variable is set, return empty string to allow
// testing missing flags in CI
func DefaultValue(ci, flag string) string {
	_, ok1 := os.LookupEnv("DOCS")
	_, ok2 := os.LookupEnv("KOSLI_TESTS")
	if !ok1 && !ok2 {
		if v, ok := ciTemplates[ci][flag]; ok {
			result := os.ExpandEnv(v)
			// github and gitlab use ../commit/.. , bitbucket uses ../commits/..
			if ci == circleci && flag == "commit-url" {
				result = gitview.ExtractRepoURLFromRemote(result)
				if strings.Contains(result, "bitbucket.org") {
					return strings.Replace(result, "commit", "commits", 1)
				}
			}
			return result
		}
		// when not in a known CI, default some values
		if flag == "git-commit" {
			return "HEAD"
		}
	}
	return ""
}

// RequireFlags declares a list of flags as required for a given command
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

// DeprecateFlags declares a list of flags as deprecated for a given command
func DeprecateFlags(cmd *cobra.Command, flags map[string]string) error {
	for name, message := range flags {
		err := cmd.Flags().MarkDeprecated(name, message)
		if err != nil {
			return err
		}
	}
	return nil
}

// MuXRequiredFlags returns an error if more than one or none (if atLeastOne is true) of
// mutually-exclusive required flags are set
func MuXRequiredFlags(cmd *cobra.Command, flagNames []string, atLeastOne bool) error {
	flagsSet := 0
	for _, name := range flagNames {
		flag := cmd.Flags().Lookup(name)
		if flag.Changed {
			flagsSet++
			if flagsSet > 1 {
				return fmt.Errorf("only one of %s is allowed", JoinFlagNames(flagNames))
			}
		}
	}
	if atLeastOne {
		if flagsSet == 0 {
			return fmt.Errorf("at least one of %s is required", JoinFlagNames(flagNames))
		}
	}
	return nil
}

func RequireAtLeastOneOfFlags(cmd *cobra.Command, flagNames []string) error {
	flagsSet := 0
	for _, name := range flagNames {
		flag := cmd.Flags().Lookup(name)
		if flag.Changed {
			flagsSet++
		}
	}
	if flagsSet == 0 {
		return fmt.Errorf("at least one of %s is required", JoinFlagNames(flagNames))
	}
	return nil
}

// JoinFlagNames returns a comma-separated string of flag names with "--" prefix
// from a list of plain names
func JoinFlagNames(flagNames []string) string {
	posixFlagNames := []string{}
	for _, flagName := range flagNames {
		posixFlagNames = append(posixFlagNames, fmt.Sprintf("--%s", flagName))
	}
	return strings.Join(posixFlagNames, ", ")
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
	}
}

func getRegistryEndpoint(url string) (*registryProviderEndpoints, error) {
	url = strings.TrimPrefix(url, "https://")
	url = strings.Split(url, "/")[0]

	return &registryProviderEndpoints{
		mainApi: "https://" + url + "/v2",
		authApi: "https://" + url + "/oauth2",
		service: url,
	}, nil
}

// getDockerRegistryAPIToken returns a short-lived read-only api token for a docker registry api
func getDockerRegistryAPIToken(providerInfo *registryProviderEndpoints, username, password, imageName string) (string, error) {
	var res *requests.HTTPResponse
	var err error

	if strings.Contains(providerInfo.service, "jfrog") {
		url := "https://" + providerInfo.service + "/artifactory/api/security/token"

		form := urlPackage.Values{}
		form.Add("username", username)
		form.Add("scope", "member-of-groups:readers")
		form.Add("expires_in", "60")

		reqParams := &requests.RequestParams{
			Method:            http.MethodPost,
			URL:               url,
			Payload:           form.Encode(),
			Username:          username,
			Password:          password,
			AdditionalHeaders: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		}
		res, err = kosliClient.Do(reqParams)
	} else {
		url := fmt.Sprintf("%s/token?scope=repository:%s:pull&service=%s", providerInfo.authApi, imageName, providerInfo.service)
		reqParams := &requests.RequestParams{
			Method:   http.MethodGet,
			URL:      url,
			Username: username,
			Password: password,
		}
		res, err = kosliClient.Do(reqParams)
	}

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
func GetSha256Digest(artifactName string, o *fingerprintOptions, logger *log.Logger) (string, error) {
	var err error
	var fingerprint string
	switch o.artifactType {
	case "file":
		fingerprint, err = digest.FileSha256(artifactName)
	case "dir":
		fingerprint, err = digest.DirSha256(artifactName, o.excludePaths, logger)
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
			imageName := nameSlice[0]
			imageTag := nameSlice[1]

			if strings.Contains(nameSlice[0], "/") {
				strSlice := strings.Split(nameSlice[0], "/")
				urlOrRepo := strSlice[0]
				if strings.Contains(urlOrRepo, ".") {
					imageName = strings.TrimPrefix(nameSlice[0], urlOrRepo+"/")
				}
			}

			if !strings.Contains(imageName, "/") && o.registryProvider == "dockerhub" {
				imageName = fmt.Sprintf("library/%s", imageName)
			}

			token, err := getDockerRegistryAPIToken(providerInfo, o.registryUsername, o.registryPassword, imageName)
			if err != nil {
				return "", err
			}

			fingerprint, err = digest.RemoteDockerImageSha256(imageName, imageTag, providerInfo.mainApi, token, logger)
			if err != nil {
				return "", err
			}

		} else {
			fingerprint, err = digest.DockerImageSha256(artifactName)
		}
	default:
		return "", fmt.Errorf("%s is not a supported artifact type", o.artifactType)
	}

	logger.Debug("calculated fingerprint: %s for artifact: %s", fingerprint, artifactName)
	return fingerprint, err
}

// LoadJsonData loads json data from a file
func LoadJsonData(filepath string) (interface{}, error) {
	var err error
	var result interface{}
	content := `{}`
	if filepath != "" {
		content, err = utils.LoadFileContent(filepath)
		if err != nil {
			return result, err
		}
		if !utils.IsJSON(content) {
			return result, fmt.Errorf("%s does not contain a valid JSON", filepath)
		}
		logger.Debug("loaded user data file content from: %s", filepath)
	}
	err = json.Unmarshal([]byte(content), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// ValidateArtifactArg validates the artifact name or path argument
func ValidateArtifactArg(args []string, artifactType, inputSha256 string, alwaysRequireArtifactName bool) error {
	if len(args) > 1 {
		suppliedArgs := []string{}
		argsWithLeadingSpace := false
		for _, arg := range args {
			if arg == " " {
				argsWithLeadingSpace = true
			}
			suppliedArgs = append(suppliedArgs, arg)
		}
		errMsg := []string{}
		errMsg = append(errMsg, "only one argument (docker image name or file/dir path) is allowed.")
		errMsg = append(errMsg, fmt.Sprintf("The %d supplied arguments are: [%v]", len(args), strings.Join(suppliedArgs, ", ")))
		if argsWithLeadingSpace {
			errMsg = append(errMsg, "Arguments with a leading space are probably caused by a lone backslash that has a space after it.")
		}
		return fmt.Errorf(strings.Join(errMsg, "\n"))
	}

	if len(args) == 0 || args[0] == "" {
		if alwaysRequireArtifactName {
			return fmt.Errorf("docker image name or file/dir path is required")
		} else if inputSha256 == "" {
			return fmt.Errorf("docker image name or file/dir path is required when --fingerprint is not provided")
		}
	}

	if artifactType == "" && inputSha256 == "" {
		return fmt.Errorf("either --artifact-type or --fingerprint must be specified")
	}

	if inputSha256 != "" {
		if err := digest.ValidateDigest(inputSha256); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAttestationArtifactArg validates the artifact name or path argument and fingerprint flag
func ValidateAttestationArtifactArg(args []string, artifactType, inputSha256 string) error {
	if artifactType != "" && (len(args) == 0 || args[0] == "") {
		return fmt.Errorf("artifact name argument is required when --artifact-type is set")
	}

	if inputSha256 != "" {
		if err := digest.ValidateDigest(inputSha256); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRegistryFlags validates that you provide all registry information necessary for
// remote digest.
func ValidateRegistryFlags(cmd *cobra.Command, o *fingerprintOptions) error {
	if o.artifactType != "docker" && (o.registryPassword != "" || o.registryUsername != "") {
		return ErrorBeforePrintingUsage(cmd, "--registry-provider, --registry-username and registry-password are only applicable when --artifact-type is 'docker'")
	}
	if o.registryProvider != "" && (o.registryPassword == "" || o.registryUsername == "") {
		return ErrorBeforePrintingUsage(cmd, "both --registry-username and registry-password are required when --registry-provider is used")
	}
	if o.registryProvider == "" && (o.registryPassword != "" || o.registryUsername != "") {
		return ErrorBeforePrintingUsage(cmd, "--registry-username and registry-password are only used when --registry-provider is used")
	}
	return nil
}

// ErrorBeforePrintingUsage
func ErrorBeforePrintingUsage(cmd *cobra.Command, errMsg string) error {
	return fmt.Errorf(
		"%s\nUsage: %s",
		errMsg,
		cmd.UseLine(),
	)
}

// tabFormattedPrint prints data in a tabular format. Takes header titles in a string slice
// and rows as a slice of strings
func tabFormattedPrint(out io.Writer, header []string, rows []string) {
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(out, 5, 12, 2, ' ', 0)
	if len(header) > 0 {
		fmt.Fprintln(w, strings.Join(header, "\t"))
	}
	for _, row := range rows {
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

// formattedTimestamp formats a float timestamp into something like "Mon, 22 Aug 2022 11:34:59 CEST • 10 days ago"
// time is formatted using RFC1123
func formattedTimestamp(timestamp interface{}, short bool) (string, error) {
	var intTimestamp int64
	var shortFormat string
	var unixTime time.Time

	switch t := timestamp.(type) {
	case int64:
		intTimestamp = timestamp.(int64)
	case float64:
		intTimestamp = int64(timestamp.(float64))
	case string:
		floatTimestamp, err := strconv.ParseFloat(timestamp.(string), 64)
		if err != nil {
			return "", err
		}
		intTimestamp = int64(floatTimestamp)
	case nil:
		return "N/A", nil
	default:
		return "", fmt.Errorf("unsupported timestamp type %s", t)
	}

	// use a fixed timestamp when running tests
	// also set timezone to UTC to make tests pass everywhere
	_, inTests := os.LookupEnv("KOSLI_TESTS")
	_, testingThisFunc := os.LookupEnv("KOSLI_TESTS_FORMATTED_TIMESTAMP")
	if inTests && testingThisFunc {
		unixTime = time.Unix(intTimestamp, 0).UTC()
	} else if inTests {
		unixTime = time.Unix(int64(1452902400), 0).UTC()
	} else {
		unixTime = time.Unix(intTimestamp, 0)
	}
	shortFormat = unixTime.Format(time.RFC1123)

	if short {
		return shortFormat, nil
	} else {
		timeago.English.Max = 36 * timeago.Month
		timeAgoFormat := timeago.English.Format(unixTime)
		return fmt.Sprintf("%s \u2022 %s", shortFormat, timeAgoFormat), nil
	}
}

// getPathOfEvidenceFileToUpload returns the path of an evidence file to upload based
// on the provided evidencePaths.
// - if one path is provided and it is a file, that path is returned as it
// - if one path is provided and it is a directory, the directory is tarred and the
// path of the generated tar file is returned
// - if multiple paths are provided, they are packaged into a tar file and the
// path of the generated tar file is returned
func getPathOfEvidenceFileToUpload(evidencePaths []string) (string, bool, error) {
	cleanupNeeded := false
	if len(evidencePaths) == 0 {
		return "", cleanupNeeded, fmt.Errorf("no evidence paths provided")
	}
	dirToTar := ""
	if len(evidencePaths) == 1 {
		ok, err := utils.IsFile(evidencePaths[0])
		if err != nil {
			return "", cleanupNeeded, err
		}
		if ok {
			logger.Debug("file %s is provided as evidence", evidencePaths[0])
			return evidencePaths[0], cleanupNeeded, nil
		}

		ok, err = utils.IsDir(evidencePaths[0])
		if err != nil {
			return "", cleanupNeeded, err
		}
		if ok {
			logger.Debug("dir %s is provided as evidence. It will be tarred", evidencePaths[0])
			dirToTar = evidencePaths[0]
		}

	} else { // there are multiple paths
		// copy all paths to a new temp dir
		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			return "", cleanupNeeded, err
		}

		logger.Debug("[%d] paths are provided as evidence. They will be tarred from %s", len(evidencePaths), tmpDir)

		for _, path := range evidencePaths {
			err := cp.Copy(path, filepath.Join(tmpDir, filepath.Base(path)))
			if err != nil {
				return "", cleanupNeeded, err
			}
		}
		dirToTar = tmpDir
		defer os.RemoveAll(tmpDir)
	}

	// tar the required dir and return the path of the tar file
	tarFilePath, err := utils.Tar(dirToTar, "evidence.tgz")
	if err != nil {
		return "", cleanupNeeded, err
	}
	cleanupNeeded = true
	return tarFilePath, cleanupNeeded, nil
}

// handleExpressions parses ~ and # expressions and returns
// a name (usually flow name), an ID (positive for fixed IDs or negative for reverse IDs),
// and an error if the expression is invalid
func handleExpressions(expression string) (string, int, error) {
	separator := ""
	hasTilda := strings.Contains(expression, "~")
	hasHash := strings.Contains(expression, "#")
	if hasTilda && hasHash {
		return "", 0, fmt.Errorf("invalid expression: %s. Both '~' and '#' are present", expression)
	} else if hasTilda {
		separator = "~"
	} else if hasHash {
		separator = "#"
	} else {
		return expression, -1, nil
	}

	items := strings.SplitN(expression, separator, 2)
	if items[0] == "" {
		return "", 0, fmt.Errorf("invalid expression: %s. Flow name is missing", expression)
	}
	id, err := strconv.Atoi(items[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid expression: %s. '%s' is not an integer", expression, items[1])
	}
	if separator == "~" {
		id = (-1 * id) - 1
	}
	return items[0], id, nil
}

// handleSnapshotExpressions parses ~, # and @ expressions and returns
// an environment name, a url encoded snapshot fragment,
// and an error if the expression is invalid
func handleSnapshotExpressions(expression string) (string, string, error) {
	separator := ""
	hasTilda := strings.Contains(expression, "~")
	hasHash := strings.Contains(expression, "#")
	hasAt := strings.Contains(expression, "@")
	if (hasTilda && hasHash) || (hasTilda && hasAt) || (hasHash && hasAt) {
		return "", "", fmt.Errorf("invalid expression: %s. Only one of '@', '~' or '#' can be present", expression)
	} else if hasTilda {
		separator = "~"
	} else if hasHash {
		separator = "#"
	} else if hasAt {
		separator = "@"
	} else {
		return expression, "-1", nil
	}

	items := strings.SplitN(expression, separator, 2)
	if items[0] == "" {
		return "", "", fmt.Errorf("invalid expression: %s. Environment name is missing", expression)
	}
	return items[0], urlPackage.PathEscape(separator + items[1]), nil
}

// handleArtifactExpression parses artifact expressions (with @ and :) and returns
// flow name, an ID (fingerprint or commit sha), the separator found, and an error
// if the expression is invalid
func handleArtifactExpression(expression string) (string, string, string, error) {
	separator := ""
	hasAt := strings.Contains(expression, "@")
	hasColon := strings.Contains(expression, ":")
	if hasAt && hasColon {
		return "", "", separator, fmt.Errorf("invalid expression: %s. Both '@' and ':' are present", expression)
	} else if hasAt {
		separator = "@"
	} else if hasColon {
		separator = ":"
	} else {
		return "", "", "", fmt.Errorf("invalid expression: %s", expression)
	}

	items := strings.SplitN(expression, separator, 2)
	if items[0] == "" {
		return "", "", "", fmt.Errorf("invalid expression: %s. Flow name is missing", expression)
	}
	if items[1] == "" {
		return "", "", "", fmt.Errorf("invalid expression: %s. Artifact identity is missing", expression)
	}

	return items[0], items[1], separator, nil
}
