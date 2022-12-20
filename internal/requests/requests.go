package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/version"
)

// HTTPResponse is a wrapper of http.Response with ready-extracted string body
type HTTPResponse struct {
	Body string
	Resp *http.Response
}

type Client struct {
	MaxAPIRetries int
	Debug         bool
	Logger        *logger.Logger
	HttpClient    *http.Client
}

func NewKosliClient(maxAPIRetries int, debug bool, logger *logger.Logger) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxAPIRetries
	retryClient.Logger = nil // this silences logging each individual attempt
	return &Client{
		MaxAPIRetries: maxAPIRetries,
		Debug:         debug,
		Logger:        logger,
		HttpClient:    retryClient.StandardClient(), // return a standard *http.Client from the retryable client
	}
}

type RequestParams struct {
	Method            string
	URL               string
	Payload           interface{}
	AdditionalHeaders map[string]string
	Username          string
	Password          string
	Token             string
	DryRun            bool
}

// newHTTPRequest returns a customized http request based on RequestParams
func (p *RequestParams) newHTTPRequest() (*http.Request, error) {
	jsonBytes, err := json.MarshalIndent(p.Payload, "", "    ")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(p.Method, p.URL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request to %s : %v", p.Method, p.URL, err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Kosli/"+version.GetVersion())

	// token authorization has higher precedence over basic auth
	if p.Token != "" {
		p.AdditionalHeaders["Authorization"] = fmt.Sprintf("Bearer %s", p.Token)
	} else if p.Username != "" || p.Password != "" {
		if p.Username == "" {
			// when communicating with Kosli, apiToken is sent as username
			// (passed to doRequest() as password)
			p.Username = p.Password
			// when communicating with Kosli, password should be "unset"
			p.Password = "unset"
		}
		req.SetBasicAuth(p.Username, p.Password)
	}

	for k, v := range p.AdditionalHeaders {
		req.Header.Set(k, v)
	}

	return req, nil
}

func (c *Client) Do(p *RequestParams) (*HTTPResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("XXXXXX")
	}
	req, err := p.newHTTPRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create a %s request to %s : %v", req.Method, req.URL, err)
	}

	if p.DryRun {
		c.Logger.Info("############### THIS IS A DRY-RUN  ###############")
		reqBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body to %s : %v", req.URL, err)
		}
		c.Logger.Info("this is the payload that would be sent in real run: \n %+v", string(reqBody))
		return nil, nil
	} else {
		resp, err := c.HttpClient.Do(req)
		if err != nil {
			// err from retryable client is detailed enough
			return nil, fmt.Errorf("%v", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response from %s request to %s : %v", req.Method, req.URL, err)
		}

		c.Logger.Debug("request made to %s and got status %d", req.URL, resp.StatusCode)

		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			var respBody map[string]interface{}
			err := json.Unmarshal([]byte(body), &respBody)
			if err != nil {
				return &HTTPResponse{}, err
			}

			cleanedErrorMessage := strings.Split(respBody["message"].(string), "You have requested")[0]
			return nil, fmt.Errorf(cleanedErrorMessage)
		}
		return &HTTPResponse{string(body), resp}, nil
	}
}

// func getRetryableHttpClient(maxAPIRetries int) *http.Client {
// 	retryClient := retryablehttp.NewClient()
// 	retryClient.RetryMax = maxAPIRetries
// 	retryClient.Logger = nil
// 	// return a standard *http.Client from the retryable client
// 	return retryClient.StandardClient()
// }

// createRequest returns an http request with a payload
func createRequest(method, url string, jsonBytes []byte, additionalHeaders map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request to %s : %v", method, url, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Kosli/"+version.GetVersion())

	for k, v := range additionalHeaders {
		req.Header.Set(k, v)
	}

	return req, nil
}

// DoRequestWithToken sends an HTTP request with auth token to a URL and returns the response body and status code
// func DoRequestWithToken(jsonBytes []byte, url, token string,
// 	maxAPIRetries int, method string, additionalHeaders map[string]string) (*HTTPResponse, error) {
// 	client := getRetryableHttpClient(maxAPIRetries)

// 	additionalHeaders["Authorization"] = fmt.Sprintf("Bearer %s", token)
// 	req, err := createRequest(method, url, jsonBytes, additionalHeaders)
// 	if err != nil {
// 		return &HTTPResponse{}, err
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return &HTTPResponse{}, fmt.Errorf("failed to send %s request to %s : %v", method, url, err)
// 	}

// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return &HTTPResponse{}, fmt.Errorf("failed to read response from %s request to %s : %v", method, url, err)
// 	}

// 	if resp.StatusCode != 200 && resp.StatusCode != 201 {
// 		return &HTTPResponse{}, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
// 	}

// 	return &HTTPResponse{
// 		Body: string(body),
// 		Resp: resp,
// 	}, nil
// }
