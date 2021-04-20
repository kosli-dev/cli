package requests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/merkely-development/watcher/internal/kube"
)

// HTTPResponse is a simplified version of http.Response
type HTTPResponse struct {
	Body       string
	StatusCode int
}

// EnvRequest represents the POST request body to be sent to merkely harvest endpoint
type EnvRequest struct {
	PodsData    []*kube.PodData
	Owner       string
	Environment string
}

func getRetryableHttpClient(maxAPIRetries int) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxAPIRetries
	client := retryClient.StandardClient() // *http.Client
	return client
}

// DoPost sends an HTTP Post request to a URL and returns the response body and status code
func DoPost(jsonBody []byte, url string, apiToken string, maxAPIRetries int) (*HTTPResponse, error) {
	requestBody := bytes.NewBuffer(jsonBody)

	client := getRetryableHttpClient(maxAPIRetries)
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to create post request to %s : %v", url, err)
	}
	req.SetBasicAuth(apiToken, "unset")
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
