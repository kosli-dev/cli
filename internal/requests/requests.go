package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
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

// CustomLogger wraps log.Logger and implements the Printf method
// It is used as a custom logger for retryableClient
type CustomLogger struct {
	*log.Logger
}

// Printf intercepts the log message and removes the hardcoded [DEBUG] part
func (cl *CustomLogger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// Remove the hardcoded [DEBUG] prefix if it exists
	msg = strings.TrimPrefix(msg, "[DEBUG]")

	// Call the underlying log.Logger's Printf method with the cleaned message
	cl.Logger.Printf(msg)
}

func NewKosliClient(httpProxyURL string, maxAPIRetries int, debug bool, logger *logger.Logger) (*Client, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxAPIRetries
	if debug {
		retryClient.Logger = &CustomLogger{
			Logger: log.New(os.Stderr, "[debug]", log.Lmsgprefix),
		}
	} else {
		retryClient.Logger = nil // this silences logging each individual attempt
	}

	client := retryClient.StandardClient() // return a standard *http.Client from the retryable client
	if httpProxyURL != "" {
		proxyURL, err := url.Parse(httpProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL when creating a Kosli http client: %s", err)
		}
		// client.Transport is already set by retryClient.StandardClient() and we add
		// the proxy to it
		client.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
	}

	return &Client{
		MaxAPIRetries: maxAPIRetries,
		Debug:         debug,
		Logger:        logger,
		HttpClient:    client,
	}, nil
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

func (p *RequestParams) newHTTPRequest() (*http.Request, map[string]interface{}, error) {
	if len(p.AdditionalHeaders) == 0 {
		p.AdditionalHeaders = make(map[string]string)
	}

	var body io.Reader
	var jsonFields map[string]interface{}

	if len(p.Form) > 0 {
		// Multipart form handling (with possible file attachments)
		var contentType string
		var err error
		contentType, body, jsonFields, err = createMultipartRequestBody(p.Form)
		if err != nil {
			return nil, nil, err
		}
		p.AdditionalHeaders["Content-Type"] = contentType
	} else {
		// JSON payload handling
		if p.Method != http.MethodGet {
			jsonBytes, err := json.MarshalIndent(p.Payload, "", "    ")
			if err != nil {
				return nil, nil, err
			}
			body = bytes.NewBuffer(jsonBytes)
		}
	}

	req, err := http.NewRequest(p.Method, p.URL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create %s request to %s : %v", p.Method, p.URL, err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Kosli/"+version.GetVersion())

	// Token-based or Basic authentication handling
	if p.Token != "" {
		p.AdditionalHeaders["Authorization"] = fmt.Sprintf("Bearer %s", p.Token)
	} else if p.Username != "" || p.Password != "" {
		req.SetBasicAuth(p.Username, p.Password)
	}

	for k, v := range p.AdditionalHeaders {
		req.Header.Set(k, v)
	}

	return req, jsonFields, nil
}

// createMultipartRequestBody processes a list of FormItem and returns:
// - the multipart form content type
// - request body for the multipart form in the form of bytes.Buffer
// - a map of the JSON fields to log during dry-run
// - error, if any occurred
func createMultipartRequestBody(items []FormItem) (string, *bytes.Buffer, map[string]interface{}, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	// Map to store the JSON fields for logging during dry-run
	jsonFields := make(map[string]interface{})

	for _, item := range items {
		if item.Type == "field" {
			part, err := writer.CreateFormField(item.FieldName)
			if err != nil {
				return "", body, nil, err
			}

			// Marshal the JSON field and add it to the multipart writer
			jsonBytes, err := json.MarshalIndent(item.Content, "", "    ")
			if err != nil {
				return "", body, nil, err
			}
			_, err = part.Write(jsonBytes)
			if err != nil {
				return "", body, nil, err
			}

			// Add the JSON field to jsonFields map for dry-run logging
			jsonFields[item.FieldName] = jsonBytes

		} else if item.Type == "file" {
			// Handle file upload separately
			filename := item.Content.(string)
			file, err := os.Open(filename)
			if err != nil {
				return "", body, nil, err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(item.FieldName, filepath.Base(filename))
			if err != nil {
				return "", body, nil, err
			}

			_, err = io.Copy(part, file)
			if err != nil {
				return "", body, nil, err
			}
		}
	}
	contentType := writer.FormDataContentType()

	// Return the content type, the body, and the JSON fields
	return contentType, body, jsonFields, nil
}

func (c *Client) Do(p *RequestParams) (*HTTPResponse, error) {
	req, jsonFields, err := p.newHTTPRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create a %s request to %s : %v", p.Method, p.URL, err)
	}

	if p.DryRun {
		c.Logger.Info("############### THIS IS A DRY-RUN  ###############")
		c.Logger.Info("the request would have been sent to: %s", req.URL)

		// Check the content type to determine what to log
		contentType := req.Header.Get("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			// Log only the JSON fields for multipart/form-data
			c.Logger.Info("this is the payload data that would be sent in a real run:")
			for key, value := range jsonFields {
				c.Logger.Info("Field: %s, Value: %+v", key, string(value.([]byte)))
			}
		} else if req.Body != nil {
			// For non-multipart requests, log the full JSON body
			reqBody, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body to %s : %v", req.URL, err)
			}
			c.Logger.Info("this is the payload that would be sent in a real run: \n %+v", string(reqBody))
		}
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
			return nil, fmt.Errorf("%s", cleanedErrorMessage)
		}
		return &HTTPResponse{string(body), resp}, nil
	}
}
