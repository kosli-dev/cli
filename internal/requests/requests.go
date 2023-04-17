package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/version"
)

type FormItem struct {
	Type      string
	FieldName string
	Content   interface{}
}

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

func NewStandardKosliClient() *Client {
	return NewKosliClient(3, false, logger.NewStandardLogger())
}

func (c *Client) SetDebug(debug bool) {
	c.Debug = debug
}

func (c *Client) SetLogger(logger *logger.Logger) {
	c.Logger = logger
}

func (c *Client) SetMaxAPIRetries(maxAPIRetries int) {
	c.MaxAPIRetries = maxAPIRetries
}

type RequestParams struct {
	Method            string
	URL               string
	Payload           interface{}
	Form              []FormItem
	AdditionalHeaders map[string]string
	Username          string
	Password          string
	Token             string
	DryRun            bool
}

// newHTTPRequest returns a customized http request based on RequestParams
func (p *RequestParams) newHTTPRequest() (*http.Request, error) {
	if len(p.AdditionalHeaders) == 0 {
		p.AdditionalHeaders = make(map[string]string)
	}

	var body *bytes.Buffer

	if len(p.Form) > 0 {
		var contentType string
		var err error
		contentType, body, err = createMultipartRequestBody(p.Form)
		if err != nil {
			return nil, err
		}
		p.AdditionalHeaders["Content-Type"] = contentType
	} else {
		jsonBytes, err := json.MarshalIndent(p.Payload, "", "    ")
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(p.Method, p.URL, body)
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

// createMultipartRequestBody process a list of FormItem and returns
// - the multipart form content type
// - request body for the multipart form in the form of bytes.Buffer
// - error, if any occurred
func createMultipartRequestBody(items []FormItem) (string, *bytes.Buffer, error) {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	defer writer.Close()

	for _, item := range items {
		if item.Type == "field" {
			part, err := writer.CreateFormField(item.FieldName)
			if err != nil {
				return "", body, err
			}

			jsonBytes, err := json.MarshalIndent(item.Content, "", "    ")
			if err != nil {
				return "", body, err
			}
			_, err = part.Write(jsonBytes)
			if err != nil {
				return "", body, err
			}
		} else if item.Type == "file" {
			filename := item.Content.(string)
			file, err := os.Open(filename)
			if err != nil {
				return "", body, err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(item.FieldName, filepath.Base(filename))
			if err != nil {
				return "", body, err
			}

			_, err = io.Copy(part, file)
			if err != nil {
				return "", body, err
			}
		}
	}
	contentType := writer.FormDataContentType()
	return contentType, body, nil
}

func (c *Client) Do(p *RequestParams) (*HTTPResponse, error) {
	req, err := p.newHTTPRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create a %s request to %s : %v", p.Method, p.URL, err)
	}

	if p.DryRun {
		c.Logger.Info("############### THIS IS A DRY-RUN  ###############")
		reqBody, err := io.ReadAll(req.Body)
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response from %s request to %s : %v", req.Method, req.URL, err)
		}

		c.Logger.Debug("request made to %s and got status %d", req.URL, resp.StatusCode)

		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			var respBody interface{}
			err := json.Unmarshal([]byte(body), &respBody)
			if err != nil {
				return &HTTPResponse{}, err
			}
			cleanedErrorMessage := ""
			if reflect.ValueOf(respBody).Kind() == reflect.String {
				cleanedErrorMessage = respBody.(string)
			} else if reflect.ValueOf(respBody).Kind() == reflect.Map {
				// Error response from kosli application SW contains a "message"
				// Error response from the API schema validation contains a "message" and a list of "errors"
				respBodyMap := respBody.(map[string]interface{})
				message, ok := respBodyMap["message"]
				if ok {
					errors, ok := respBodyMap["errors"]
					if ok {
						cleanedErrorMessage = strings.Split(message.(string), "You have requested")[0] +
							": " + fmt.Sprintf("%v", errors)
					} else {
						cleanedErrorMessage = strings.Split(message.(string), "You have requested")[0]
					}
				} else {
					cleanedErrorMessage = fmt.Sprintf("%s", respBodyMap)
				}
			}
			return nil, fmt.Errorf(cleanedErrorMessage)
		}
		return &HTTPResponse{string(body), resp}, nil
	}
}
