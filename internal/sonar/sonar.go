package sonar

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SonarConfig struct {
	ProjectKey    string
	APIToken      string
	BranchName    string
	PullRequestID string
}

type SonarResults struct {
	Component Component `json:"component"`
}

type Component struct {
	Id          string     `json:"id,omitempty"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Qualifier   string     `json:"qualifier"`
	Measures    []Measures `json:"measures"`
	Branch      string     `json:"branch,omitempty"`
	PullRequest string     `json:"pullRequest,omitempty"`
}

type Measures struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}

func NewSonarConfig(projectKey, apiToken, branchName, pullRequestID string) *SonarConfig {
	return &SonarConfig{
		ProjectKey:    projectKey,
		APIToken:      apiToken,
		BranchName:    branchName,
		PullRequestID: pullRequestID,
	}
}

func (sc *SonarConfig) GetSonarResults() (*SonarResults, error) {
	httpClient := &http.Client{}
	var url string
	var token string

	if sc.ProjectKey != "" && sc.APIToken != "" {
		url = fmt.Sprintf("https://sonarcloud.io/api/measures/component?metricKeys=alert_status%%2Cquality_gate_details%%2Cbugs%%2Csecurity_issues%%2Ccode_smells%%2Ccomplexity%%2Cmaintainability_issues%%2Creliability_issues%%2Ccoverage&component=%s", sc.ProjectKey)
		token = fmt.Sprintf("Bearer %s", sc.APIToken)
	} else {
		return nil, fmt.Errorf("Project Key and API token must be given to retrieve data from SonarCloud/SonarQube")
	}

	if sc.BranchName != "" && sc.PullRequestID != "" {
		return nil, fmt.Errorf("Branch Name and Pull Request ID cannot both be given")
	}

	if sc.BranchName != "" {
		url = fmt.Sprintf("%s&branch=%s", url, sc.BranchName)
	} else if sc.PullRequestID != "" {
		url = fmt.Sprintf("%s&pullRequest=%s", url, sc.PullRequestID)
	}

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", token)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	sonarResult := &SonarResults{Component: Component{}}
	err = json.NewDecoder(response.Body).Decode(sonarResult)
	if err != nil {
		return nil, err
	}

	//With incorrect project key or API token we receive no data
	if sonarResult.Component.Key == "" {
		return nil, fmt.Errorf("No data retrieved from Sonarcloud - check your project key and API token are correct")
	}

	return sonarResult, nil
}
