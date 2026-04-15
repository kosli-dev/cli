package sonar

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/kosli-dev/cli/internal/logger"
)

type SonarConfig struct {
	APIToken    string
	WorkingDir  string
	CETaskUrl   string
	revision    string
	projectKey  string
	serverURL   string
	pullRequest string
	maxWait     int
}

// Structs to build the JSON for our attestation payload
type SonarResults struct {
	ServerUrl   string       `json:"serverUrl"`
	TaskID      string       `json:"taskId"`
	Status      string       `json:"status"`
	AnalysedAt  string       `json:"analysedAt"`
	Revision    string       `json:"revision"`
	Project     Project      `json:"project"`
	Branch      *Branch      `json:"branch,omitempty"`
	PullRequest string       `json:"pullRequest,omitempty"`
	QualityGate *QualityGate `json:"qualityGate,omitempty"`
}

type Project struct {
	Key  string `json:"key"`
	Name string `json:"name,omitempty"`
	Url  string `json:"url"`
}

type Branch struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type QualityGate struct {
	Status     string      `json:"status"`
	Conditions []Condition `json:"conditions"`
}

type Condition struct {
	Metric         string `json:"metric"`
	ErrorThreshold string `json:"errorThreshold"`
	Operator       string `json:"operator"`
	Value          string `json:"value,omitempty"`
	Status         string `json:"status"`
}

// These are the structs for the response from the qualitygates/project_status API
type QualityGateResponse struct {
	ProjectStatus ProjectStatus `json:"projectStatus"`
	Errors        []Error       `json:"errors,omitempty"`
}

type ProjectStatus struct {
	Status     string       `json:"status"`
	Conditions []Conditions `json:"conditions"`
}

type Conditions struct {
	Status         string `json:"status"`
	MetricKey      string `json:"metricKey"`
	Comparator     string `json:"comparator"`
	ErrorThreshold string `json:"errorThreshold"`
	ActualValue    string `json:"actualValue"`
}

// These are the structs for the response from the ceTaskURL
type TaskResponse struct {
	Task   Task    `json:"task"`
	Errors []Error `json:"errors,omitempty"`
}
type Task struct {
	TaskID        string `json:"id"`
	ComponentName string `json:"componentName"`
	ComponentKey  string `json:"componentKey"`
	AnalysisID    string `json:"analysisId"`
	Status        string `json:"status"`
	Branch        string `json:"branch"`
	BranchType    string `json:"branchType"`
	PullRequest   string `json:"pullRequest"`
}

type ActivityResponse struct {
	Tasks []Task `json:"tasks"`
}

// These are the structs for the response from the project_analyses/search API
type ProjectAnalyses struct {
	Analyses []Analysis `json:"analyses"`
	Errors   []Error    `json:"errors,omitempty"`
}

type Analysis struct {
	Key      string `json:"key"`
	Date     string `json:"date"`
	Revision string `json:"revision"`
}

// These are the structs for the response from the project_pull_requests/list API
type PullRequestsResponse struct {
	PullRequests []PullRequestInfo `json:"pullRequests"`
	Errors       []Error           `json:"errors,omitempty"`
}

type PullRequestInfo struct {
	Key          string   `json:"key"`
	Branch       string   `json:"branch"`
	AnalysisDate string   `json:"analysisDate"`
	Commit       PRCommit `json:"commit"`
}

type PRCommit struct {
	SHA string `json:"sha"`
}

// Struct for error messages from sonar APIs
type Error struct {
	Msg string `json:"msg"`
}

func NewSonarConfig(apiToken, workingDir, ceTaskUrl, projectKey, serverURL, revision, pullRequest string, maxWait int) *SonarConfig {
	return &SonarConfig{
		APIToken:    apiToken,
		WorkingDir:  workingDir,
		CETaskUrl:   ceTaskUrl,
		revision:    revision,
		projectKey:  projectKey,
		serverURL:   serverURL,
		pullRequest: pullRequest,
		maxWait:     maxWait,
	}
}

func sonarURL(serverURL, apiPath string, params url.Values) (string, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}
	u = u.JoinPath(apiPath)
	u.RawQuery = params.Encode()
	return u.String(), nil
}

func (sc *SonarConfig) GetSonarResults(logger *log.Logger) (*SonarResults, error) {
	httpClient := &http.Client{}
	var analysisID, tokenHeader string
	var err error
	project := &Project{}
	qualityGate := &QualityGate{}
	sonarResults := &SonarResults{}

	//Check if the API token is given
	if sc.APIToken != "" {
		tokenHeader = fmt.Sprintf("Bearer %s", sc.APIToken)
	} else {
		return nil, fmt.Errorf("API token must be given to retrieve data from SonarQube")
	}

	// If explicit pull-request flag was given, set it on the results
	if sc.pullRequest != "" {
		sonarResults.PullRequest = sc.pullRequest
	}

	// Read the report-task.txt file (if it exists) to get the server URL, dashboard URL and ceTaskURL
	err = sc.readFile(project, sonarResults, logger)
	if err != nil {
		if sc.CETaskUrl != "" {
			// If the CE task URL is provided directly (e.g. via --sonar-ce-task-url), we can skip the report-task.txt
			// and use the CE task URL to get the data. Extract the server URL from the CE task URL.
			parsedURL, parseErr := url.Parse(sc.CETaskUrl)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse CE task URL: %s", parseErr)
			}
			sonarResults.ServerUrl = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		} else if sc.projectKey == "" || (sc.revision == "" && sc.pullRequest == "") {
			return nil, fmt.Errorf("%s. Alternatively provide the project key and either revision or pull-request ID for the scan to attest", err)
		} else {
			// If the report-task.txt does not exist, but we've been given the project key and revision (or PR ID), we can still get the data
			project.Key = sc.projectKey
			sonarResults.ServerUrl = sc.serverURL
			sonarResults.Revision = sc.revision
			project.Url, err = sonarURL(sonarResults.ServerUrl, "dashboard", url.Values{"id": {project.Key}})
			if err != nil {
				return nil, err
			}
			if sonarResults.PullRequest == "" {
				analysisID, err = GetProjectAnalysisFromRevision(httpClient, sonarResults, project, sc.revision, tokenHeader, logger)
				if err != nil {
					return nil, err
				}
			}

			err = GetTaskID(httpClient, sonarResults, project, analysisID, tokenHeader, logger)
			if err != nil {
				return nil, err
			}
		}
	}

	if analysisID == "" && sc.CETaskUrl != "" {
		//Get the analysis ID, status, project name and branch data from the ceTaskURL (ce API)
		analysisID, err = GetCETaskData(httpClient, project, sonarResults, sc.CETaskUrl, tokenHeader, sc.maxWait, logger)
		if err != nil {
			return nil, err
		}

		if sonarResults.PullRequest == "" {
			//Get project revision and scan date/time from the projectAnalyses API
			err = GetProjectAnalysisFromAnalysisID(httpClient, sonarResults, project, analysisID, tokenHeader)
			if err != nil {
				return nil, err
			}
		}
		// PR case falls through to the block below
	}

	// If we have a PR get PR analysis data
	if sonarResults.PullRequest != "" {
		err = GetPRAnalysisData(httpClient, sonarResults, project, sonarResults.PullRequest, tokenHeader)
		if err != nil {
			return nil, err
		}
	}

	//Get the quality gate status from the qualitygates/project_status API
	qualityGate, err = GetQualityGate(httpClient, sonarResults, qualityGate, analysisID, project.Key, sonarResults.PullRequest, tokenHeader)
	if err != nil {
		return nil, err
	}

	sonarResults.Project = *project
	sonarResults.QualityGate = qualityGate

	return sonarResults, nil
}

func (sc *SonarConfig) readFile(project *Project, results *SonarResults, logger *log.Logger) error {
	metadata, err := os.Open(filepath.Join(sc.WorkingDir, "report-task.txt"))
	if err != nil {
		return fmt.Errorf("%s. Check your working directory is set correctly", err)
	}
	defer func() {
		if err := metadata.Close(); err != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close metadata file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(metadata)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "serverUrl=") {
			results.ServerUrl = strings.TrimPrefix(line, "serverUrl=")
			continue
		}
		if strings.HasPrefix(line, "dashboardUrl=") {
			project.Url = strings.TrimPrefix(line, "dashboardUrl=")
			continue
		}
		if strings.HasPrefix(line, "ceTaskUrl=") {
			sc.CETaskUrl = strings.TrimPrefix(line, "ceTaskUrl=")
			continue
		}
	}
	return nil
}

func GetCETaskData(httpClient *http.Client, project *Project, sonarResults *SonarResults, ceTaskURL, tokenHeader string, maxWait int, logger *log.Logger) (string, error) {
	taskRequest, err := http.NewRequest("GET", ceTaskURL, nil)
	if err != nil {
		return "", err
	}
	taskRequest.Header.Add("Authorization", tokenHeader)

	wait := 1    // start wait period
	retries := 0 // number of retries so far
	elapsed := 0 // seconds elapsed
	taskResponseData := &TaskResponse{}

	for elapsed < maxWait {
		taskResponse, err := httpClient.Do(taskRequest)
		if err != nil {
			return "", err
		}
		err = json.NewDecoder(taskResponse.Body).Decode(taskResponseData)
		if err != nil {
			_ = taskResponse.Body.Close()
			return "", fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
		}
		// If the CETaskURL from the report-task.txt file gives a 404, the CE task does not exist, or SonarQube is down.
		if taskResponseData.Errors != nil {
			_ = taskResponse.Body.Close()
			return "", fmt.Errorf("%s on %s. \nSonarQube may be experiencing problems, please check https://status.sonarqube.com/ and try again later. \nOtherwise if you are attesting an older scan, the snapshot may have been deleted by SonarQube", taskResponseData.Errors[0].Msg, sonarResults.ServerUrl)
		}

		if err := taskResponse.Body.Close(); err != nil {
			logger.Warn("failed to close task response body: %v", err)
		}

		if taskResponseData.Task.Status == "PENDING" || taskResponseData.Task.Status == "IN_PROGRESS" {
			// So that we don't wait longer than maxWait
			if (elapsed + wait) > maxWait {
				wait = maxWait - elapsed
			}

			logger.Info("retry %d: waiting %ds for SonarQube scan to be processed...", retries+1, wait)
			time.Sleep(time.Duration(wait) * time.Second)

			if elapsed > 300 { // If we've waited 5 minutes, we'll wait 300 seconds (5 minutes) before checking again. This is so that we don't end up with extremely long waiting intervals with the doubling approach.
				elapsed += wait
				retries++
				wait += 300
			} else { // Otherwise, we'll double the wait time each time
				elapsed += wait
				retries++
				wait *= 2
			}
		} else {
			break
		}
	}

	if elapsed != 0 {
		logger.Info("Waited for %d seconds for SonarQube scan to be processed. %d retries.\n", elapsed, retries)
	}

	task := taskResponseData.Task

	project.Name = task.ComponentName
	project.Key = task.ComponentKey
	sonarResults.TaskID = task.TaskID
	analysisId := task.AnalysisID
	sonarResults.Status = task.Status

	// This should only happen if the task is pending - either because the project is large and the scan takes a long time
	// to process, or because SonarQube is experiencing delays for some reason.
	if analysisId == "" {
		return "", fmt.Errorf("analysis ID not found on %s. The scan results are not yet available, likely due to: \n1. Your project being particularly large and the scan taking time to process, or \n2. SonarQube experiencing delays in processing scans. \nTry rerunning the command with the --max-wait flag", sonarResults.ServerUrl)
	}

	if project.Url == "" {
		project.Url, err = sonarURL(sonarResults.ServerUrl, "dashboard", url.Values{"id": {project.Key}})
		if err != nil {
			return "", err
		}
	}

	if taskResponseData.Task.PullRequest != "" {
		sonarResults.PullRequest = taskResponseData.Task.PullRequest
		sonarResults.Branch = nil
	} else if taskResponseData.Task.Branch != "" {
		sonarResults.Branch = &Branch{}
		sonarResults.Branch.Name = taskResponseData.Task.Branch
		sonarResults.Branch.Type = taskResponseData.Task.BranchType
	} else {
		sonarResults.Branch = nil
	}

	return analysisId, nil
}

func GetProjectAnalysisFromRevision(httpClient *http.Client, sonarResults *SonarResults, project *Project, revision, tokenHeader string, logger *log.Logger) (string, error) {
	var analysisID string

	projectAnalysesURL, err := sonarURL(sonarResults.ServerUrl, "api/project_analyses/search", url.Values{"project": {project.Key}})
	if err != nil {
		return "", err
	}
	projectAnalysesRequest, err := http.NewRequest("GET", projectAnalysesURL, nil)
	if err != nil {
		return "", err
	}
	projectAnalysesRequest.Header.Add("Authorization", tokenHeader)

	projectAnalysesResponse, err := httpClient.Do(projectAnalysesRequest)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := projectAnalysesResponse.Body.Close(); err != nil {
			logger.Warn("failed to close project analyses response body: %v", err)
		}
	}()

	projectAnalysesData := &ProjectAnalyses{}
	err = json.NewDecoder(projectAnalysesResponse.Body).Decode(projectAnalysesData)
	if err != nil {
		return "", fmt.Errorf("please check your API token and SonarQube server URL are correct and you have the correct permissions in SonarQube")
	}

	if projectAnalysesData.Errors != nil {
		return "", fmt.Errorf("SonarQube error: %s", projectAnalysesData.Errors[0].Msg)
	}

	for analysis := range projectAnalysesData.Analyses {
		if projectAnalysesData.Analyses[analysis].Revision == revision {
			sonarResults.AnalysedAt = projectAnalysesData.Analyses[analysis].Date
			analysisID = projectAnalysesData.Analyses[analysis].Key
			break
		}
	}

	if sonarResults.AnalysedAt == "" {
		return "", fmt.Errorf("analysis for revision %s of project %s not found. Check the revision is correct. \nThe scan may still be being processed by SonarQube, try again later.\n Otherwise if you are attesting an older scan, the snapshot may also have been deleted by SonarQube", revision, project.Key)
	}

	return analysisID, nil
}

func GetProjectAnalysisFromAnalysisID(httpClient *http.Client, sonarResults *SonarResults, project *Project, analysisID, tokenHeader string) error {
	projectAnalysesURL, err := sonarURL(sonarResults.ServerUrl, "api/project_analyses/search", url.Values{"project": {project.Key}})
	if err != nil {
		return err
	}
	projectAnalysesRequest, err := http.NewRequest("GET", projectAnalysesURL, nil)
	if err != nil {
		return err
	}
	projectAnalysesRequest.Header.Add("Authorization", tokenHeader)

	projectAnalysesResponse, err := httpClient.Do(projectAnalysesRequest)
	if err != nil {
		return err
	}
	defer func() { _ = projectAnalysesResponse.Body.Close() }()

	projectAnalysesData := &ProjectAnalyses{}
	err = json.NewDecoder(projectAnalysesResponse.Body).Decode(projectAnalysesData)
	if err != nil {
		return fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
	}

	if projectAnalysesData.Errors != nil {
		return fmt.Errorf("SonarQube error: %s", projectAnalysesData.Errors[0].Msg)
	}

	for analysis := range projectAnalysesData.Analyses {
		if projectAnalysesData.Analyses[analysis].Key == analysisID {
			sonarResults.AnalysedAt = projectAnalysesData.Analyses[analysis].Date
			sonarResults.Revision = projectAnalysesData.Analyses[analysis].Revision
			break
		}
	}

	if sonarResults.AnalysedAt == "" {
		return fmt.Errorf("analysis with ID %s not found on %s. Snapshot may have been deleted by SonarQube", analysisID, sonarResults.ServerUrl)
	}

	return nil
}

// GetPRAnalysisData retrieves the revision and analysis date for a pull request scan
// from the project_pull_requests/list API. This is needed because the project_analyses/search
// API does not return PR analyses on SonarCloud.
func GetPRAnalysisData(httpClient *http.Client, sonarResults *SonarResults, project *Project, prKey, tokenHeader string) error {
	prURL, err := sonarURL(sonarResults.ServerUrl, "api/project_pull_requests/list", url.Values{"project": {project.Key}})
	if err != nil {
		return err
	}
	prRequest, err := http.NewRequest("GET", prURL, nil)
	if err != nil {
		return err
	}
	prRequest.Header.Add("Authorization", tokenHeader)

	prResponse, err := httpClient.Do(prRequest)
	if err != nil {
		return err
	}
	defer func() { _ = prResponse.Body.Close() }()

	prData := &PullRequestsResponse{}
	err = json.NewDecoder(prResponse.Body).Decode(prData)
	if err != nil {
		return fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
	}

	if prData.Errors != nil {
		return fmt.Errorf("SonarQube error: %s", prData.Errors[0].Msg)
	}

	for _, pr := range prData.PullRequests {
		if pr.Key == prKey {
			sonarResults.AnalysedAt = pr.AnalysisDate
			sonarResults.Revision = pr.Commit.SHA
			return nil
		}
	}

	return fmt.Errorf("pull request %s not found for project %s on %s", prKey, project.Key, sonarResults.ServerUrl)
}

func GetQualityGate(httpClient *http.Client, sonarResults *SonarResults, qualityGate *QualityGate, analysisID, projectKey, pullRequest, tokenHeader string) (*QualityGate, error) {
	var qualityGateURL string
	var err error
	if pullRequest == "" {
		qualityGateURL, err = sonarURL(sonarResults.ServerUrl, "api/qualitygates/project_status", url.Values{"analysisId": {analysisID}})
		if err != nil {
			return nil, err
		}
	} else {
		qualityGateURL, err = sonarURL(sonarResults.ServerUrl, "api/qualitygates/project_status", url.Values{"projectKey": {projectKey}, "pullRequest": {pullRequest}})
		if err != nil {
			return nil, err
		}
	}
	qualityGateRequest, err := http.NewRequest("GET", qualityGateURL, nil)
	if err != nil {
		return nil, err
	}
	qualityGateRequest.Header.Add("Authorization", tokenHeader)

	qualityGateResponse, err := httpClient.Do(qualityGateRequest)
	if err != nil {
		return nil, err
	}
	defer func() { _ = qualityGateResponse.Body.Close() }()

	qualityGateData := &QualityGateResponse{}
	err = json.NewDecoder(qualityGateResponse.Body).Decode(qualityGateData)
	if err != nil {
		return nil, err
	} else if qualityGateData.Errors != nil {
		return nil, fmt.Errorf("SonarQube error: %s", qualityGateData.Errors[0].Msg) //We should never reach this point, since incorrect/outdated task/analysis IDs etc. should already have raised errors
	} else {
		qualityGate.Status = qualityGateData.ProjectStatus.Status
		// The server expects an array of conditions if the Quality Gate exists, so if there are no conditions, we need to send an empty array
		if len(qualityGateData.ProjectStatus.Conditions) == 0 {
			qualityGate.Conditions = []Condition{}
		} else {
			for condition := range qualityGateData.ProjectStatus.Conditions {
				qualityGate.Conditions = append(qualityGate.Conditions, Condition{
					Metric:         qualityGateData.ProjectStatus.Conditions[condition].MetricKey,
					ErrorThreshold: qualityGateData.ProjectStatus.Conditions[condition].ErrorThreshold,
					Operator:       qualityGateData.ProjectStatus.Conditions[condition].Comparator,
					Value:          qualityGateData.ProjectStatus.Conditions[condition].ActualValue,
					Status:         qualityGateData.ProjectStatus.Conditions[condition].Status,
				})
			}
		}
	}

	return qualityGate, nil
}

func GetTaskID(httpClient *http.Client, sonarResults *SonarResults, project *Project, analysisID, tokenHeader string, logger *log.Logger) error {
	CEActivityURL, err := sonarURL(sonarResults.ServerUrl, "api/ce/activity", url.Values{"component": {project.Key}})
	if err != nil {
		return err
	}
	CEActivityRequest, err := http.NewRequest("GET", CEActivityURL, nil)
	if err != nil {
		return err
	}
	CEActivityRequest.Header.Add("Authorization", tokenHeader)

	CEActivityResponse, err := httpClient.Do(CEActivityRequest)
	if err != nil {
		return err
	}
	defer func() {
		if err := CEActivityResponse.Body.Close(); err != nil {
			logger.Warn("failed to close CE activity response body: %v", err)
		}
	}()

	CEActivityData := &ActivityResponse{}
	err = json.NewDecoder(CEActivityResponse.Body).Decode(CEActivityData)
	if err != nil {
		return fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
	}

	for t := range CEActivityData.Tasks {
		task := CEActivityData.Tasks[t]
		matched := (analysisID != "" && task.AnalysisID == analysisID) ||
			(analysisID == "" && sonarResults.PullRequest != "" && task.PullRequest == sonarResults.PullRequest)
		if matched {
			sonarResults.TaskID = task.TaskID
			sonarResults.Status = task.Status
			project.Name = task.ComponentName
			if task.PullRequest != "" {
				sonarResults.PullRequest = task.PullRequest
				sonarResults.Branch = nil
			} else if task.Branch != "" {
				sonarResults.Branch = &Branch{}
				sonarResults.Branch.Name = task.Branch
				sonarResults.Branch.Type = task.BranchType
			} else {
				sonarResults.Branch = nil
			}
			break
		}
	}

	return nil
}
