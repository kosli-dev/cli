package sonar

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/kosli-dev/cli/internal/logger"
)

type SonarConfig struct {
	APIToken   string
	WorkingDir string
	CETaskUrl  string
	revision   string
	projectKey string
	serverURL  string
	allowWait  int
}

// Structs to build the JSON for our attestation payload
type SonarResults struct {
	ServerUrl   string       `json:"serverUrl"`
	TaskID      string       `json:"taskId"`
	Status      string       `json:"status"`
	AnalaysedAt string       `json:"analysedAt"`
	Revision    string       `json:"revision"`
	Project     Project      `json:"project"`
	Branch      *Branch      `json:"branch,omitempty"`
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
	Task Task `json:"task"`
}
type Task struct {
	TaskID        string `json:"id"`
	ComponentName string `json:"componentName"`
	ComponentKey  string `json:"componentKey"`
	AnalysisID    string `json:"analysisId"`
	Status        string `json:"status"`
	Branch        string `json:"branch"`
	BranchType    string `json:"branchType"`
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

// Struct for error messages from sonar APIs
type Error struct {
	Msg string `json:"msg"`
}

func NewSonarConfig(apiToken, workingDir, ceTaskUrl, projectKey, serverURL, revision string, allowWait int) *SonarConfig {
	return &SonarConfig{
		APIToken:   apiToken,
		WorkingDir: workingDir,
		CETaskUrl:  ceTaskUrl,
		revision:   revision,
		projectKey: projectKey,
		serverURL:  serverURL,
		allowWait:  allowWait,
	}
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

	// Read the report-task.txt file (if it exists) to get the project key, server URL, dashboard URL and ceTaskURL
	err = sc.readFile(project, sonarResults)
	if err != nil {
		if sc.projectKey == "" || sc.revision == "" {
			return nil, fmt.Errorf("%s. Alternatively provide the project key and revision for the scan to attest", err)
			// If the report-task.txt does not exist, but we've been given the project key and revision, we can still get the data
		} else {
			project.Key = sc.projectKey
			sonarResults.ServerUrl = sc.serverURL
			sonarResults.Revision = sc.revision
			project.Url = fmt.Sprintf("%s/dashboard?id=%s", sonarResults.ServerUrl, project.Key)
			analysisID, err = GetProjectAnalysisFromRevision(httpClient, sonarResults, project, sc.revision, tokenHeader)
			if err != nil {
				return nil, err
			}
			err = GetTaskID(httpClient, sonarResults, project, analysisID, tokenHeader)
			if err != nil {
				return nil, err
			}
		}
	}

	if analysisID == "" {
		//Get the analysis ID, status, project name and branch data from the ceTaskURL (ce API)
		analysisID, err = GetCETaskData(httpClient, project, sonarResults, sc.CETaskUrl, tokenHeader, sc.allowWait, logger)
		if err != nil {
			return nil, err
		}

		//Get project revision and scan date/time from the projectAnalyses API
		err = GetProjectAnalysisFromAnalysisID(httpClient, sonarResults, project, analysisID, tokenHeader)
		if err != nil {
			return nil, err
		}
	}

	//Get the quality gate status from the qualitygates/project_status API
	qualityGate, err = GetQualityGate(httpClient, sonarResults, qualityGate, analysisID, tokenHeader)
	if err != nil {
		return nil, err
	}

	sonarResults.Project = *project
	sonarResults.QualityGate = qualityGate

	return sonarResults, nil
}

func (sc *SonarConfig) readFile(project *Project, results *SonarResults) error {
	metadata, err := os.Open(filepath.Join(sc.WorkingDir, "report-task.txt"))
	if err != nil {
		return fmt.Errorf("%s. Check your working directory is set correctly", err)
	}
	defer metadata.Close()

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

func GetCETaskData(httpClient *http.Client, project *Project, sonarResults *SonarResults, ceTaskURL, tokenHeader string, allowWait int, logger *log.Logger) (string, error) {
	taskRequest, err := http.NewRequest("GET", ceTaskURL, nil)
	taskRequest.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return "", err
	}

	wait := 10 // seconds to sleep between checks
	if allowWait < wait {
		wait = allowWait
	}
	maxWait := allowWait // 20 minutes
	elapsed := 0         // seconds elapsed
	taskResponseData := &TaskResponse{}

	for elapsed < maxWait || maxWait == 0 {
		taskResponse, err := httpClient.Do(taskRequest)
		if err != nil {
			return "", err
		}

		err = json.NewDecoder(taskResponse.Body).Decode(taskResponseData)
		if err != nil {
			return "", fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
		}

		if maxWait == 0 {
			taskResponse.Body.Close()
			break
		}

		//taskResponseData.Task.Status = "PENDING"
		if taskResponseData.Task.Status == "PENDING" || taskResponseData.Task.Status == "IN_PROGRESS" {
			logger.Info("waiting %ds for SonarQube scan to be processed... \n%d seconds elapsed", wait, elapsed)
			time.Sleep(time.Duration(wait) * time.Second)
			if elapsed > 300 { // If we've waited 5 minutes, we'll wait 300 seconds (5 minutes) before checking again. This is so that we don't end up with extremely long waiting intervals with the doubling approach.
				elapsed += wait
				wait += 300
			} else { // Otherwise, we'll double the wait time each time
				elapsed += wait
				wait *= 2
			}
		} else {
			taskResponse.Body.Close()
			break
		}
	}

	if elapsed != 0 {
		logger.Info("Waited for %d seconds for SonarQube scan to be processed", elapsed)
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
		return "", fmt.Errorf("analysis ID not found on %s. The scan results are not yet available, likely due to: \n1. Your project being particularly large and the scan taking time to process, or \n2. SonarQube is experiencing delays in processing scans. \nTry rerunning the command with the --allow-wait flag.", sonarResults.ServerUrl)
	}

	if project.Url == "" {
		project.Url = fmt.Sprintf("%s/dashboard?id=%s", sonarResults.ServerUrl, project.Key)
	}

	if taskResponseData.Task.Branch != "" {
		sonarResults.Branch = &Branch{}
		sonarResults.Branch.Name = taskResponseData.Task.Branch
		sonarResults.Branch.Type = taskResponseData.Task.BranchType
	} else {
		sonarResults.Branch = nil
	}

	return analysisId, nil
}

func GetProjectAnalysisFromRevision(httpClient *http.Client, sonarResults *SonarResults, project *Project, revision, tokenHeader string) (string, error) {
	var analysisID string

	projectAnalysesURL := fmt.Sprintf("%s/api/project_analyses/search?project=%s", sonarResults.ServerUrl, project.Key)
	projectAnalysesRequest, err := http.NewRequest("GET", projectAnalysesURL, nil)
	projectAnalysesRequest.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return "", err
	}

	projectAnalysesResponse, err := httpClient.Do(projectAnalysesRequest)
	if err != nil {
		return "", err
	}

	projectAnalysesData := &ProjectAnalyses{}
	err = json.NewDecoder(projectAnalysesResponse.Body).Decode(projectAnalysesData)
	if err != nil {
		return "", fmt.Errorf("please check your API token and SonarQube server URL are correct and you have the correct permissions in SonarQube")
	}

	for analysis := range projectAnalysesData.Analyses {
		if projectAnalysesData.Analyses[analysis].Revision == revision {
			sonarResults.AnalaysedAt = projectAnalysesData.Analyses[analysis].Date
			analysisID = projectAnalysesData.Analyses[analysis].Key
			break
		}
	}

	if projectAnalysesData.Errors != nil {
		return "", fmt.Errorf("sonar error: %s", projectAnalysesData.Errors[0].Msg)
	}

	if sonarResults.AnalaysedAt == "" {
		return "", fmt.Errorf("analysis for revision %s of project %s not found. Check the revision is correct. Snapshot may also have been deleted by SonarQube", revision, project.Key)
	}
	projectAnalysesResponse.Body.Close()

	return analysisID, nil
}

func GetProjectAnalysisFromAnalysisID(httpClient *http.Client, sonarResults *SonarResults, project *Project, analysisID, tokenHeader string) error {
	projectAnalysesURL := fmt.Sprintf("%s/api/project_analyses/search?project=%s", sonarResults.ServerUrl, project.Key)
	projectAnalysesRequest, err := http.NewRequest("GET", projectAnalysesURL, nil)
	projectAnalysesRequest.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return err
	}

	projectAnalysesResponse, err := httpClient.Do(projectAnalysesRequest)
	if err != nil {
		return err
	}

	projectAnalysesData := &ProjectAnalyses{}
	err = json.NewDecoder(projectAnalysesResponse.Body).Decode(projectAnalysesData)
	if err != nil {
		return fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
	}

	for analysis := range projectAnalysesData.Analyses {
		if projectAnalysesData.Analyses[analysis].Key == analysisID {
			sonarResults.AnalaysedAt = projectAnalysesData.Analyses[analysis].Date
			sonarResults.Revision = projectAnalysesData.Analyses[analysis].Revision
			break
		}
	}

	if sonarResults.AnalaysedAt == "" {
		return fmt.Errorf("analysis with ID %s not found. Snapshot may have been deleted by SonarQube", analysisID)
	}

	return nil
}

func GetQualityGate(httpClient *http.Client, sonarResults *SonarResults, qualityGate *QualityGate, analysisID, tokenHeader string) (*QualityGate, error) {
	qualityGateURL := fmt.Sprintf("%s/api/qualitygates/project_status?analysisId=%s", sonarResults.ServerUrl, analysisID)
	qualityGateRequest, err := http.NewRequest("GET", qualityGateURL, nil)
	qualityGateRequest.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return nil, err
	}

	qualityGateResponse, err := httpClient.Do(qualityGateRequest)
	if err != nil {
		return nil, err
	}

	qualityGateData := &QualityGateResponse{}
	err = json.NewDecoder(qualityGateResponse.Body).Decode(qualityGateData)
	if err != nil {
		return nil, err
	} else if qualityGateData.Errors != nil {
		return nil, fmt.Errorf("sonar error: %s", qualityGateData.Errors[0].Msg) //We should never reach this point, since incorrect/outdated task/analysis IDs etc. should already have raised errors
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

func GetTaskID(httpClient *http.Client, sonarResults *SonarResults, project *Project, analysisID, tokenHeader string) error {
	CEActivityURL := fmt.Sprintf("%s/api/ce/activity?component=%s", sonarResults.ServerUrl, project.Key)
	CEActivityRequest, err := http.NewRequest("GET", CEActivityURL, nil)
	CEActivityRequest.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return err
	}

	CEActivityResponse, err := httpClient.Do(CEActivityRequest)
	if err != nil {
		return err
	}

	CEActivityData := &ActivityResponse{}
	err = json.NewDecoder(CEActivityResponse.Body).Decode(CEActivityData)
	if err != nil {
		return fmt.Errorf("please check your API token is correct and you have the correct permissions in SonarQube")
	}

	for task := range CEActivityData.Tasks {
		if CEActivityData.Tasks[task].AnalysisID == analysisID {
			sonarResults.TaskID = CEActivityData.Tasks[task].TaskID
			sonarResults.Status = CEActivityData.Tasks[task].Status
			project.Name = CEActivityData.Tasks[task].ComponentName
			if CEActivityData.Tasks[task].Branch != "" {
				sonarResults.Branch = &Branch{}
				sonarResults.Branch.Name = CEActivityData.Tasks[task].Branch
				sonarResults.Branch.Type = CEActivityData.Tasks[task].BranchType
			} else {
				sonarResults.Branch = nil
			}
			break
		}
	}
	CEActivityResponse.Body.Close()

	return nil
}
