package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

// HTTPResponse is a wrapper of http.Response with ready-extracted string body
type HTTPResponse struct {
	Body string
	Resp *http.Response
}

func getRetryableHttpClient(maxAPIRetries int, logger *logrus.Logger) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxAPIRetries
	retryClient.Logger = nil
	// return a standard *http.Client from the retryable client
	return retryClient.StandardClient()
}

// createRequest returns an http request with a payload
func createRequest(method, url string, jsonBytes []byte, additionalHeaders map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request to %s : %v", method, url, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	for k, v := range additionalHeaders {
		req.Header.Set(k, v)
	}

	return req, nil
}

// DoBasicAuthRequest sends an HTTP request with basic auth to a URL and returns the response body and status code
func DoBasicAuthRequest(jsonBytes []byte, url, username, password string,
	maxAPIRetries int, method string, additionalHeaders map[string]string, logger *logrus.Logger) (*HTTPResponse, error) {
	client := getRetryableHttpClient(maxAPIRetries, logger)

	req, err := createRequest(method, url, jsonBytes, additionalHeaders)

	if err != nil {
		return &HTTPResponse{}, err
	}

	if username == "" {
		// when communicating with Kosli, apiToken is sent as username
		// (passed to doRequest() as password)
		username = password
		// when communicating with Kosli, password should be "unset"
		password = "unset"
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)

	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to send %s request to %s : %v", method, url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to read response from %s request to %s : %v", method, url, err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return &HTTPResponse{}, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return &HTTPResponse{
		Body: string(body),
		Resp: resp,
	}, nil
}

// DoRequestWithToken sends an HTTP request with auth token to a URL and returns the response body and status code
func DoRequestWithToken(jsonBytes []byte, url, token string,
	maxAPIRetries int, method string, additionalHeaders map[string]string, logger *logrus.Logger) (*HTTPResponse, error) {
	client := getRetryableHttpClient(maxAPIRetries, logger)

	additionalHeaders["Authorization"] = fmt.Sprintf("Bearer %s", token)
	req, err := createRequest(method, url, jsonBytes, additionalHeaders)
	if err != nil {
		return &HTTPResponse{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to send %s request to %s : %v", method, url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to read response from %s request to %s : %v", method, url, err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return &HTTPResponse{}, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return &HTTPResponse{
		Body: string(body),
		Resp: resp,
	}, nil
}

// SendPayload sends a JSON payload to a URL
func SendPayload(payload interface{}, url, username, token string, maxRetries int, dryRun bool, method string, logger *logrus.Logger) (*HTTPResponse, error) {
	var resp *HTTPResponse
	jsonBytes, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return resp, err
	}

	if dryRun {
		logger.Info("############### THIS IS A DRY-RUN  ###############")
		logger.Info(string(jsonBytes))
	} else {
		resp, err = DoBasicAuthRequest(jsonBytes, url, username, token, maxRetries, method, map[string]string{}, logger)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}
