package requests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/kube"
	"github.com/merkely-development/reporter/internal/server"
)

// HTTPResponse is a simplified version of http.Response
type HTTPResponse struct {
	Body       string
	StatusCode int
}

// K8sEnvRequest represents the PUT request body to be sent to merkely from k8s
type K8sEnvRequest struct {
	Artifacts []*kube.PodData `json:"artifacts"`
	Type      string          `json:"type"`
	Id        string          `json:"id"`
}

// EcsEnvRequest represents the PUT request body to be sent to merkely from ECS
type EcsEnvRequest struct {
	Artifacts []*aws.EcsTaskData `json:"artifacts"`
	Type      string             `json:"type"`
	Id        string             `json:"id"`
}

// ServerEnvRequest represents the PUT request body to be sent to merkely from a server
type ServerEnvRequest struct {
	Artifacts []*server.ServerData `json:"artifacts"`
	Type      string               `json:"type"`
	Id        string               `json:"id"`
}

func getRetryableHttpClient(maxAPIRetries int) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxAPIRetries
	// get a standard *http.Client from the retryable client
	client := retryClient.StandardClient()
	return client
}

// doRequest sends an HTTP request to a URL and returns the response body and status code
func doRequest(jsonBody []byte, url string, apiToken string, maxAPIRetries int, method string) (*HTTPResponse, error) {
	client := getRetryableHttpClient(maxAPIRetries)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to create post request to %s : %v", url, err)
	}
	req.SetBasicAuth(apiToken, "unset")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)

	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to send post request to %s : %v", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to read response from post request to %s : %v", url, err)
	}

	return &HTTPResponse{
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}, nil
}

// SendPayload sends a JSON payload to a URL
func SendPayload(payload []byte, url, token string, maxRetries int, dryRun bool, method string) error {
	if dryRun {
		fmt.Println("############### THIS IS A DRY-RUN  ###############")
		fmt.Println(string(payload))
	} else {
		fmt.Println("****** Sending the payload to the API ******")
		fmt.Println(string(payload))
		resp, err := doRequest(payload, url, token, maxRetries, method)
		if err != nil {
			return err
		}
		if resp.StatusCode != 201 && resp.StatusCode != 200 {
			return fmt.Errorf("failed to send payload. Got status %d: %v", resp.StatusCode, resp.Body)
		}
	}
	return nil
}
