package sonar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type SonarConfig struct {
	ProjectKey    string
	APIToken      string
	SonarQubeURL  string
	BranchName    string
	PullRequestID string
}

type SonarResults struct {
	Component Component `json:"component"`
}

type Component struct {
	Id          string     `json:"id,omitempty"`
	Description string     `json:"description,omitempty"`
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

func NewSonarConfig(projectKey, apiToken, sonarQubeUrl, branchName, pullRequestID string) *SonarConfig {
	return &SonarConfig{
		ProjectKey:    projectKey,
		APIToken:      apiToken,
		SonarQubeURL:  sonarQubeUrl,
		BranchName:    branchName,
		PullRequestID: pullRequestID,
	}
}

func (sc *SonarConfig) GetSonarResults() (*SonarResults, error) {
	httpClient := &http.Client{}
	var baseUrl, fullUrl, tokenHeader string
	//metrics := []string{"alert_status", "quality_gate_details", "bugs", "security_issues", "code_smells", "complexity", "maintainability_issues", "reliability_issues", "coverage"}
	metrics := "alert_status,quality_gate_details,bugs,security_issues,code_smells,complexity,maintainability_issues,reliability_issues,coverage"

	if sc.SonarQubeURL != "" {
		baseUrl = sc.SonarQubeURL
	} else {
		baseUrl = "https://sonarcloud.io"
	}

	if sc.ProjectKey != "" && sc.APIToken != "" {
		metricsPath := url.PathEscape(metrics)
		fullUrl = fmt.Sprintf("%s/api/measures/component?metricKeys=%s&component=%s", baseUrl, metricsPath, sc.ProjectKey)
		tokenHeader = fmt.Sprintf("Bearer %s", sc.APIToken)
	} else {
		return nil, fmt.Errorf("Project Key and API token must be given to retrieve data from SonarCloud/SonarQube")
	}

	if sc.BranchName != "" && sc.PullRequestID != "" {
		return nil, fmt.Errorf("Branch Name and Pull Request ID cannot both be given")
	}

	if sc.BranchName != "" {
		fullUrl = fmt.Sprintf("%s&branch=%s", fullUrl, sc.BranchName)
	} else if sc.PullRequestID != "" {
		fullUrl = fmt.Sprintf("%s&pullRequest=%s", fullUrl, sc.PullRequestID)
	}

	request, err := http.NewRequest("GET", fullUrl, nil)
	request.Header.Add("Authorization", tokenHeader)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		//If incorrect URL given, HTTP request returns error
		return nil, fmt.Errorf("Incorrect SonarQube URL")
	}

	sonarResult := &SonarResults{Component: Component{}}
	err = json.NewDecoder(response.Body).Decode(sonarResult)
	if err != nil {
		//If the API token is incorrect, SonarCloud returns nothing, thus we get a Decode error
		return nil, fmt.Errorf("Incorrect API token or SonarQube URL")
	}

	//If the project key/branch name/pull request id is incorrect, SonarCloud returns an error
	//and therefore the component key will be empty
	if sonarResult.Component.Key == "" {
		return nil, fmt.Errorf("No data retrieved - check your project key and branch or pull request id are correct")
	}

	return sonarResult, nil
}
