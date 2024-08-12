package sonar

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SonarConfig struct {
	APIToken   string
	WorkingDir string
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
	Name string `json:"name",omitempty`
	Url  string `json:"url"`
}

type Branch struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type QualityGate struct {
	Status     string      `json:"status"`
	Conditions []Condition `json:"conditions"`
	//Name       string      `json:"name"` I cannot find a way to find out which quality gate was used for a specific scan
}

type Condition struct {
	Metric         string `json:"metric"`
	ErrorThreshold string `json:"errorThreshold"`
	Operator       string `json:"operator"`
	Value          string `json:"value",omitempty`
	Status         string `json:"status"`
}

// These are the structs for the response from the qualitygates/project_status API
type QualityGateResponse struct {
	ProjectStatus ProjectStatus `json:"projectStatus"`
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
	ComponentName string `json:"componentName"`
	AnalysisID    string `json:"analysisId"`
	Status        string `json:"status"`
	Branch        string `json:"branch"`
	BranchType    string `json:"branchType"`
}

// These are the structs for the response from the project_analyses/search API
type ProjectAnalyses struct {
	Analyses []Analysis `json:"analyses"`
}
type Analysis struct {
	Key      string `json:"key"`
	Date     string `json:"date"`
	Revision string `json:"revision"`
}

func NewSonarConfig(apiToken, workingDir string) *SonarConfig {
	return &SonarConfig{
		APIToken:   apiToken,
		WorkingDir: workingDir,
	}
}

func (sc *SonarConfig) GetSonarResults() (*SonarResults, error) {
	httpClient := &http.Client{}
	var ceTaskURL string
	var analysisID, tokenHeader string
	project := &Project{}
	qualityGate := &QualityGate{}
	sonarResults := &SonarResults{}

	//Check if the API token is given
	if sc.APIToken != "" {
		tokenHeader = fmt.Sprintf("Bearer %s", sc.APIToken)
	} else {
		return nil, fmt.Errorf("API token must be given to retrieve data from SonarCloud/SonarQube")
	}

	//Read the report-task.txt file to get the project key, server URL, dashboard URL and ceTaskURL
	ceTaskURL, err := sc.readFile(project, sonarResults)
	if err != nil {
		return nil, err
	}

	//Get the analysis ID, status, project name and branch data from the ceTaskURL (ce API)
	analysisID, err = GetCETaskData(httpClient, project, sonarResults, ceTaskURL, tokenHeader)
	if err != nil {
		return nil, err
	}

	//Get project revision and scan date/time from the projectAnalyses API
	err = GetProjectAnalysis(httpClient, sonarResults, project, analysisID, tokenHeader)
	if err != nil {
		return nil, err
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

func (sc *SonarConfig) readFile(project *Project, results *SonarResults) (string, error) {
	var ceTaskURL string

	metadata, err := os.Open(filepath.Join(sc.WorkingDir, "report-task.txt"))
	if err != nil {
		return "", fmt.Errorf("report-task.txt not found. Check your working directory is set correctly.")
	}
	defer metadata.Close()

	scanner := bufio.NewScanner(metadata)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "projectKey=") {
			project.Key = strings.TrimPrefix(line, "projectKey=")
			continue
		}
		if strings.HasPrefix(line, "serverUrl=") {
			results.ServerUrl = strings.TrimPrefix(line, "serverUrl=")
			continue
		}
		if strings.HasPrefix(line, "dashboardUrl=") {
			project.Url = strings.TrimPrefix(line, "dashboardUrl=")
			continue
		}
		if strings.HasPrefix(line, "ceTaskUrl=") {
			ceTaskURL = strings.TrimPrefix(line, "ceTaskUrl=")
			continue
		}
	}
	return ceTaskURL, nil
}

func GetCETaskData(httpClient *http.Client, project *Project, sonarResults *SonarResults, ceTaskURL, tokenHeader string) (string, error) {
	taskRequest, err := http.NewRequest("GET", ceTaskURL, nil)
	taskRequest.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return "", err
	}

	taskResponse, err := httpClient.Do(taskRequest)
	if err != nil {
		return "", err
	}

	taskResponseData := &TaskResponse{}
	err = json.NewDecoder(taskResponse.Body).Decode(taskResponseData)
	if err != nil {
		return "", fmt.Errorf("Please check your API token is correct and you have the correct permissions in SonarCloud/SonarQube.")
	}

	project.Name = taskResponseData.Task.ComponentName
	sonarResults.TaskID = taskResponseData.Task.AnalysisID
	sonarResults.Status = taskResponseData.Task.Status

	if taskResponseData.Task.Branch != "" {
		sonarResults.Branch = &Branch{}
		sonarResults.Branch.Name = taskResponseData.Task.Branch
		sonarResults.Branch.Type = taskResponseData.Task.BranchType
	} else {
		sonarResults.Branch = nil
	}

	return sonarResults.TaskID, nil
}

func GetProjectAnalysis(httpClient *http.Client, sonarResults *SonarResults, project *Project, analysisID, tokenHeader string) error {
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
		return fmt.Errorf("Please check your API token is correct and you have the correct permissions in SonarCloud/SonarQube.")
	}

	for analysis := range projectAnalysesData.Analyses {
		if projectAnalysesData.Analyses[analysis].Key == analysisID {
			sonarResults.AnalaysedAt = projectAnalysesData.Analyses[analysis].Date
			sonarResults.Revision = projectAnalysesData.Analyses[analysis].Revision
			break
		}
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
		qualityGate = nil
	} else {
		qualityGate.Status = qualityGateData.ProjectStatus.Status
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

	return qualityGate, nil
}
