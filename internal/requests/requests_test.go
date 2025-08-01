package requests

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/version"
	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type RequestsTestSuite struct {
	suite.Suite
	fakeService *httpfake.HTTPFake
}

// create a fakeserver before the suite execution
func (suite *RequestsTestSuite) SetupSuite() {
	suite.fakeService = httpfake.New()
	suite.fakeService.NewHandler().
		Put("/artifacts/1").
		Reply(201).
		BodyString(`{"sha": "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f","name": "artifact"}`)
	suite.fakeService.NewHandler().
		Put("/html").
		Reply(200).
		BodyString(`<!DOCTYPE html>
		<html lang="en"><head>
		  <meta charset="utf-8">`)
	suite.fakeService.NewHandler().
		Get("/no-go/").
		Reply(404).
		BodyString(`{"message": "resource not found"}`)
	suite.fakeService.NewHandler().
		Get("/bad-request1/").
		Reply(400).
		BodyString(`{"message": "Input payload validation failed", "errors": [{"name":"'123foo' does not match '^[a-zA-Z][a-zA-Z0-9\\-]*$'"}]}`)
	suite.fakeService.NewHandler().
		Get("/bad-request2/").
		Reply(400).
		BodyString(`{"error": "random error"}`)
	suite.fakeService.NewHandler().
		Get("/denied/").
		Reply(403).
		BodyString(`"Denied"`)
	suite.fakeService.NewHandler().
		Get("/fail/").
		Reply(500).
		BodyString("server broken")
}

// shutdown the fake service after the suite execution
func (suite *RequestsTestSuite) TearDownSuite() {
	suite.fakeService.Close()
}

func (suite *RequestsTestSuite) TestNewKosliClient() {
	for _, t := range []struct {
		name       string
		httpProxy  string
		maxRetries int
		debug      bool
		wantError  bool
	}{
		{
			name:       "client is created with expected settings 1",
			maxRetries: 1,
			debug:      true,
		},
		{
			name:       "client is created with expected settings 2",
			maxRetries: 3,
			debug:      false,
		},
		{
			name:       "client is created with an http proxy",
			maxRetries: 3,
			debug:      false,
			httpProxy:  "http://192.0.0.1:8001",
		},
		{
			name:       "client creation fails when http proxy URL is invalid",
			maxRetries: 3,
			debug:      false,
			httpProxy:  "http://:foo.com",
			wantError:  true,
		},
	} {
		suite.Suite.Run(t.name, func() {
			client, err := NewKosliClient(t.httpProxy, t.maxRetries, t.debug, logger.NewStandardLogger())
			if !t.wantError {
				require.NoError(suite.Suite.T(), err)
				require.NotNil(suite.Suite.T(), client)
				require.Equal(suite.Suite.T(), t.maxRetries, client.MaxAPIRetries)
				require.Equal(suite.Suite.T(), t.debug, client.Debug)
			} else {
				require.Error(suite.Suite.T(), err)
			}

		})
	}
}

func (suite *RequestsTestSuite) TestNewHttpRequest() {
	for _, t := range []struct {
		name                      string
		params                    *RequestParams
		wantError                 bool
		expectedContentTypePrefix string
	}{
		{
			name: "request with token",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    "https://google.com",
				Token:  "secret",
			},
		},
		{
			name: "request with user/pass",
			params: &RequestParams{
				Method:   http.MethodGet,
				URL:      "https://google.com",
				Username: "user",
				Password: "password",
			},
		},
		{
			name: "request with password only (like Kosli requests)",
			params: &RequestParams{
				Method:   http.MethodGet,
				URL:      "https://google.com",
				Password: "password",
			},
		},
		{
			name: "request with additional headers",
			params: &RequestParams{
				Method:   http.MethodGet,
				URL:      "https://google.com",
				Username: "user",
				Password: "password",
				AdditionalHeaders: map[string]string{
					"HEADER1": "VALUE1",
					"HEADER2": "VALUE2",
				},
			},
		},
		{
			name: "request with valid payload",
			params: &RequestParams{
				Method:   http.MethodPost,
				URL:      "https://google.com",
				Username: "user",
				Password: "password",
				Payload:  "test payload",
			},
		},
		{
			name: "request with invalid URL (starts with space) causes an error",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    " https://google.com",
				Token:  "secret",
			},
			wantError: true,
		},
		{
			name: "request with invalid payload causes an error",
			params: &RequestParams{
				Method:  http.MethodPost,
				URL:     "https://google.com",
				Token:   "secret",
				Payload: make(chan string),
			},
			wantError: true,
		},
		{
			name: "request with form works",
			params: &RequestParams{
				Method: http.MethodPost,
				URL:    "https://google.com",
				Token:  "secret",
				Form: []FormItem{
					{
						Type:      "field",
						FieldName: "field1",
						Content:   "some content",
					},
					{
						Type:      "file",
						FieldName: "field2",
						Content:   "requests.go",
					},
				},
			},
			expectedContentTypePrefix: "multipart/form-data; boundary=",
		},
		{
			name: "request with form that has invalid content causes an error",
			params: &RequestParams{
				Method: http.MethodPost,
				URL:    "https://google.com",
				Token:  "secret",
				Form: []FormItem{
					{
						Type:      "field",
						FieldName: "field1",
						Content:   make(chan string),
					},
				},
			},
			wantError: true,
		},
	} {
		suite.Suite.Run(t.name, func() {
			req, _, err := t.params.newHTTPRequest()
			if t.wantError {
				require.Error(suite.Suite.T(), err)
			} else {
				require.NoError(suite.Suite.T(), err)
				require.Equal(suite.Suite.T(), t.params.Method, req.Method)
				require.Equal(suite.Suite.T(), "Kosli/"+version.GetVersion(), req.UserAgent())
				if t.expectedContentTypePrefix == "" {
					t.expectedContentTypePrefix = "application/json; charset=utf-8"
				}
				require.True(suite.Suite.T(), strings.HasPrefix(req.Header.Get("Content-Type"), t.expectedContentTypePrefix))
				if t.params.Username != "" || t.params.Password != "" {
					user, pass, ok := req.BasicAuth()
					require.True(suite.Suite.T(), ok)
					require.Equal(suite.Suite.T(), t.params.Username, user)
					require.Equal(suite.Suite.T(), t.params.Password, pass)
				}
				if t.params.Token != "" {
					require.Equal(suite.Suite.T(), fmt.Sprintf("Bearer %s", t.params.Token), req.Header.Get("Authorization"))
				}
				for k, v := range t.params.AdditionalHeaders {
					require.Equal(suite.Suite.T(), v, req.Header.Get(k))
				}

				if t.params.Method == http.MethodGet {
					require.Nil(suite.Suite.T(), req.Body)
				}
			}
		})
	}
}

func (suite *RequestsTestSuite) TestDo() {
	for _, t := range []struct {
		name             string
		params           *RequestParams
		wantError        bool
		expectedLog      string
		expectedErrorMsg string
		expectedBody     string
	}{
		{
			name: "GET request to cyber-dojo with fake password",
			params: &RequestParams{
				Method:   http.MethodGet,
				URL:      "https://app.kosli.com/api/v2/environments/cyber-dojo",
				Password: "secret",
			},
		},
		{
			name: "PUT request to fake server",
			params: &RequestParams{
				Method: http.MethodPut,
				URL:    suite.fakeService.ResolveURL("/artifacts/1"),
			},
			expectedBody: `{"sha": "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f","name": "artifact"}`,
		},
		{
			name: "GET request to 404 endpoint",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/no-go/"),
			},
			wantError:        true,
			expectedErrorMsg: "resource not found",
		},
		{
			name: "GET request to 500 endpoint",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/fail/"),
			},
			wantError:        true,
			expectedErrorMsg: fmt.Sprintf("Get \"%s\": GET %s giving up after 2 attempt(s)", suite.fakeService.ResolveURL("/fail/"), suite.fakeService.ResolveURL("/fail/")),
		},
		{
			name: "GET request with invalid URL causes an error",
			params: &RequestParams{
				Method:   http.MethodGet,
				URL:      "  https://app.kosli.com/api/v2/environments/cyber-dojo/foo",
				Password: "secret",
			},
			wantError:        true,
			expectedErrorMsg: "failed to create a GET request to   https://app.kosli.com/api/v2/environments/cyber-dojo/foo : failed to create GET request to   https://app.kosli.com/api/v2/environments/cyber-dojo/foo : parse \"  https://app.kosli.com/api/v2/environments/cyber-dojo/foo\": first path segment in URL cannot contain colon",
		},
		{
			name: "PUT request to cyber-dojo with dry-run",
			params: &RequestParams{
				Method:   http.MethodPut,
				URL:      "https://app.kosli.com/api/v2/environments/cyber-dojo",
				Password: "secret",
				DryRun:   true,
				Payload:  "some payload",
			},
			expectedLog: "############### THIS IS A DRY-RUN  ###############\nthe request would have been sent to: https://app.kosli.com/api/v2/environments/cyber-dojo\nthis is the payload that would be sent in a real run: \n \"some payload\"\n",
		},
		{
			name: "GET request to cyber-dojo with dry-run",
			params: &RequestParams{
				Method:   http.MethodGet,
				URL:      "https://app.kosli.com/api/v2/environments/cyber-dojo",
				Password: "secret",
				DryRun:   true,
			},
			expectedLog: "############### THIS IS A DRY-RUN  ###############\nthe request would have been sent to: https://app.kosli.com/api/v2/environments/cyber-dojo\n",
		},
		{
			name: "GET request to 400 endpoint with message and errors in response",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/bad-request1/"),
			},
			wantError:        true,
			expectedErrorMsg: "Input payload validation failed: [map[name:'123foo' does not match '^[a-zA-Z][a-zA-Z0-9\\-]*$']]",
		},
		{
			name: "GET request to 400 endpoint with no message in response",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/bad-request2/"),
			},
			wantError:        true,
			expectedErrorMsg: "map[error:random error]",
		},
		{
			name: "GET request to 403 endpoint",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/denied/"),
			},
			wantError:        true,
			expectedErrorMsg: "Denied",
		},
		{
			name: "GET request to a PUT endpoint fails because of invalid response",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/html"),
			},
			wantError:        true,
			expectedErrorMsg: "unexpected end of JSON input",
		},
	} {
		suite.Suite.Run(t.name, func() {
			buf := new(bytes.Buffer)
			client, err := NewKosliClient("", 1, false, logger.NewLogger(buf, buf, false))
			require.NoError(suite.Suite.T(), err)
			resp, err := client.Do(t.params)
			if t.wantError {
				require.Error(suite.Suite.T(), err)
				require.Equal(suite.Suite.T(), t.expectedErrorMsg, err.Error())
			} else {
				require.NoError(suite.Suite.T(), err)
				output := buf.String()
				require.Equal(suite.Suite.T(), t.expectedLog, output)
				if t.expectedBody != "" {
					require.Equal(suite.Suite.T(), t.expectedBody, resp.Body)
				}

			}
		})
	}
}

func (suite *RequestsTestSuite) TestDebugPayloadOutput() {
	for _, t := range []struct {
		name             string
		params           *RequestParams
		wantError        bool
		expectedLog      string
		expectedErrorMsg string
		expectedBody     string
	}{
		{
			name: "PUT request with debug logs payload",
			params: &RequestParams{
				Method: http.MethodPut,
				URL:    "http://localhost:8001/api/v2/environments/docs-cmd-test-user-shared",
				DryRun: false,
				Token:  "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
				Payload: struct {
					Name              string `json:"name"`
					Type              string `json:"type"`
					Description       string `json:"description"`
					IncludeScaling    bool   `json:"include_scaling"`
					RequireProvenance bool   `json:"require_provenance"`
				}{
					Name:              "test-environment",
					Type:              "K8S",
					Description:       "test-environment",
					IncludeScaling:    true,
					RequireProvenance: true,
				},
			},
			expectedLog: "############### PAYLOAD ###############\npayload sent to: http://localhost:8001/api/v2/environments/docs-cmd-test-user-shared\nthis is the payload being sent: \n {\n    \"name\": \"test-environment\",\n    \"type\": \"K8S\",\n    \"description\": \"test-environment\",\n    \"include_scaling\": true,\n    \"require_provenance\": true\n}\n",
		},
		{
			name: "GET request with debug does not log non-existent payload",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    "https://app.kosli.com/api/v2/environments/cyber-dojo",
				DryRun: false,
			},
		},
	} {
		suite.Suite.Run(t.name, func() {
			buf := new(bytes.Buffer)
			client, err := NewKosliClient("", 1, true, logger.NewLogger(buf, buf, false))
			require.NoError(suite.Suite.T(), err)
			resp, err := client.Do(t.params)
			require.NoError(suite.Suite.T(), err)

			output := buf.String()
			require.Equal(suite.Suite.T(), t.expectedLog, output)
			if t.expectedBody != "" {
				require.Equal(suite.Suite.T(), t.expectedBody, resp.Body)
			}
		})
	}
}

func (suite *RequestsTestSuite) TestCreateMultipartRequestBody() {
	for _, t := range []struct {
		name                      string
		formItems                 []FormItem
		wantError                 bool
		expectedErrorMsg          string
		expectedContentTypePrefix string
	}{
		{
			name: "a form can be created from one item",
			formItems: []FormItem{
				{
					Type:      "field",
					FieldName: "data",
					Content:   "some text",
				},
			},
			expectedContentTypePrefix: "multipart/form-data; boundary=",
		},
		{
			name: "a form can be created from multiple items",
			formItems: []FormItem{
				{
					Type:      "field",
					FieldName: "data",
					Content:   "some text",
				},
				{
					Type:      "file",
					FieldName: "upload",
					Content:   "requests.go",
				},
			},
			expectedContentTypePrefix: "multipart/form-data; boundary=",
		},
		{
			name: "a form with a non-existing file item fails",
			formItems: []FormItem{
				{
					Type:      "file",
					FieldName: "upload",
					Content:   "non-existing",
				},
			},
			wantError: true,
		},
		{
			name: "a form with an invalid field content item fails",
			formItems: []FormItem{
				{
					Type:      "field",
					FieldName: "data",
					Content:   make(chan string),
				},
			},
			wantError: true,
		},
	} {
		suite.Suite.Run(t.name, func() {
			contentType, _, _, err := createMultipartRequestBody(t.formItems)
			require.True(suite.Suite.T(), t.wantError == (err != nil))
			require.True(suite.Suite.T(), strings.HasPrefix(contentType, t.expectedContentTypePrefix))
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRequestsTestSuite(t *testing.T) {
	suite.Run(t, new(RequestsTestSuite))
}
