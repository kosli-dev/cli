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
	Errors    []Error   `json:"errors,omitempty"` //So we can give the user the detailed error message from SonarCloud/SonarQube
}

type Error struct {
	Msg string `json:"msg,omitempty"`
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
	Metric    string   `json:"metric"`
	Value     string   `json:"value"`
	BestValue bool     `json:"bestValue,omitempty"`
	Periods   []Period `json:"periods,omitempty"`
}

type Period struct {
	Index     int    `json:"index"`
	Value     string `json:"value"`
	BestValue bool   `json:"bestValue,omitempty"`
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
	metrics := GetMetrics()

	if sc.SonarQubeURL != "" {
		baseUrl = sc.SonarQubeURL
		metrics = fmt.Sprintf("%s,new_security_issues", metrics) //This metric is available via the sonarqube api but not sonarcloud
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
		//If a non-existent URL given, HTTP request returns error
		return nil, fmt.Errorf("Incorrect SonarQube URL")
	}

	sonarResult := &SonarResults{Component: Component{}}
	err = json.NewDecoder(response.Body).Decode(sonarResult)
	if err != nil {
		//If the API token or SonarQube URL is incorrect, SonarCloud/Qube returns nothing, thus we get a Decode error
		return nil, fmt.Errorf("Incorrect API token or SonarQube URL")
	}

	//If the project key/branch name/pull request id is incorrect or a metric key is invalid, SonarCloud/Qube returns an error
	if sonarResult.Errors != nil {
		message := ""
		for errorIndex := range sonarResult.Errors {
			message = fmt.Sprintf("%s%s", message, sonarResult.Errors[errorIndex].Msg)
			if errorIndex != len(sonarResult.Errors)-1 {
				message = fmt.Sprintf("%s\n", message)
			}
		}
		return nil, fmt.Errorf(message)
	}

	return sonarResult, nil
}
