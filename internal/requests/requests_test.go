package requests

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
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
		Put("/accepted/").
		Reply(202).
		BodyString(`{"status": "pending"}`)
	suite.fakeService.NewHandler().
		Delete("/no-content/").
		Reply(204)
	suite.fakeService.NewHandler().
		Get("/partial/").
		Reply(206).
		BodyString(`{"message": "partial`)
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
		Get("/locked/").
		Reply(409).
		BodyString("resource temporarily locked")
	suite.fakeService.NewHandler().
		Get("/conflict/").
		Reply(409).
		BodyString(`{"message": "already exists"}`)
	suite.fakeService.NewHandler().
		Get("/fail/").
		Reply(500).
		BodyString("server broken")
	suite.fakeService.NewHandler().
		Get("/proxy-block/").
		Reply(403).
		BodyString(`<!DOCTYPE html><html><body><h1>403 Forbidden</h1><p>Blocked by proxy</p></body></html>`)
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
		suite.Run(t.name, func() {
			client, err := NewKosliClient(t.httpProxy, t.maxRetries, t.debug, logger.NewStandardLogger())
			if !t.wantError {
				require.NoError(suite.T(), err)
				require.NotNil(suite.T(), client)
				require.Equal(suite.T(), t.maxRetries, client.MaxAPIRetries)
				require.Equal(suite.T(), t.debug, client.Debug)
			} else {
				require.Error(suite.T(), err)
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
		suite.Run(t.name, func() {
			req, _, err := t.params.newHTTPRequest()
			if t.wantError {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				require.Equal(suite.T(), t.params.Method, req.Method)
				require.Equal(suite.T(), "Kosli/"+version.GetVersion(), req.UserAgent())
				if t.expectedContentTypePrefix == "" {
					t.expectedContentTypePrefix = "application/json; charset=utf-8"
				}
				require.True(suite.T(), strings.HasPrefix(req.Header.Get("Content-Type"), t.expectedContentTypePrefix))
				if t.params.Username != "" || t.params.Password != "" {
					user, pass, ok := req.BasicAuth()
					require.True(suite.T(), ok)
					require.Equal(suite.T(), t.params.Username, user)
					require.Equal(suite.T(), t.params.Password, pass)
				}
				if t.params.Token != "" {
					require.Equal(suite.T(), fmt.Sprintf("Bearer %s", t.params.Token), req.Header.Get("Authorization"))
				}
				for k, v := range t.params.AdditionalHeaders {
					require.Equal(suite.T(), v, req.Header.Get(k))
				}

				if t.params.Method == http.MethodGet {
					require.Nil(suite.T(), req.Body)
				}
			}
		})
	}
}

func (suite *RequestsTestSuite) TestDo() {
	for _, t := range []struct {
		name             string
		params           *RequestParams
		debug            bool
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
			name: "202 is a success and returns its body",
			params: &RequestParams{
				Method: http.MethodPut,
				URL:    suite.fakeService.ResolveURL("/accepted/"),
			},
			expectedBody: `{"status": "pending"}`,
		},
		{
			name: "204 is not a success: callers read the body as the answer",
			params: &RequestParams{
				Method: http.MethodDelete,
				URL:    suite.fakeService.ResolveURL("/no-content/"),
			},
			wantError:        true,
			expectedErrorMsg: "HTTP 204 No Content",
		},
		{
			name: "206 is not a success: a partial body is not the answer",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/partial/"),
			},
			wantError:        true,
			expectedErrorMsg: "HTTP 206 Partial Content: the response is not JSON, rerun with --debug to see it",
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
			name: "GET request to 409 endpoint",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/locked/"),
			},
			wantError:        true,
			expectedErrorMsg: fmt.Sprintf("Get \"%s\": giving up after 2 attempt(s): HTTP 409 Conflict", suite.fakeService.ResolveURL("/locked/")),
		},
		{
			name: "409 is not retried and is surfaced when DisableConflictRetry is set",
			params: &RequestParams{
				Method:               http.MethodGet,
				URL:                  suite.fakeService.ResolveURL("/conflict/"),
				DisableConflictRetry: true,
			},
			wantError:        true,
			expectedErrorMsg: "already exists",
		},
		{
			name: "GET request to 500 endpoint",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/fail/"),
			},
			wantError:        true,
			expectedErrorMsg: fmt.Sprintf("Get \"%s\": giving up after 2 attempt(s): HTTP 500 Internal Server Error", suite.fakeService.ResolveURL("/fail/")),
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
			expectedErrorMsg: "HTTP 404 Not Found",
		},
		{
			name: "non-JSON error response body is logged with debug enabled",
			params: &RequestParams{
				Method: http.MethodGet,
				URL:    suite.fakeService.ResolveURL("/proxy-block/"),
			},
			debug:            true,
			wantError:        true,
			expectedErrorMsg: "HTTP 403 Forbidden: the response is not JSON",
			expectedLog:      `<!DOCTYPE html><html><body><h1>403 Forbidden</h1><p>Blocked by proxy</p></body></html>`,
		},
	} {
		suite.Run(t.name, func() {
			buf := new(bytes.Buffer)
			client, err := NewKosliClient("", 1, t.debug, logger.NewLogger(buf, buf, t.debug))
			require.NoError(suite.T(), err)
			resp, err := client.Do(t.params)
			if t.wantError {
				require.Error(suite.T(), err)
				require.Equal(suite.T(), t.expectedErrorMsg, err.Error())
				if t.expectedLog != "" {
					require.Contains(suite.T(), buf.String(), t.expectedLog)
				}
			} else {
				require.NoError(suite.T(), err)
				output := buf.String()
				require.Equal(suite.T(), t.expectedLog, output)
				if t.expectedBody != "" {
					require.Equal(suite.T(), t.expectedBody, resp.Body)
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
			// The method is logged beside the URL because the two together are
			// what identifies the request. Reading the log to find out what a
			// command sends - which is how the empty-flag audit learns which
			// payload field a flag controls - a URL alone does not say whether
			// it was written to with a PUT or a POST.
			expectedLog: "############### PAYLOAD ###############\npayload sent to: PUT http://localhost:8001/api/v2/environments/docs-cmd-test-user-shared\nthis is the payload being sent: \n {\n    \"name\": \"test-environment\",\n    \"type\": \"K8S\",\n    \"description\": \"test-environment\",\n    \"include_scaling\": true,\n    \"require_provenance\": true\n}\n",
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
		suite.Run(t.name, func() {
			buf := new(bytes.Buffer)
			client, err := NewKosliClient("", 1, true, logger.NewLogger(buf, buf, false))
			require.NoError(suite.T(), err)
			resp, err := client.Do(t.params)
			require.NoError(suite.T(), err)

			output := buf.String()
			require.Equal(suite.T(), t.expectedLog, output)
			if t.expectedBody != "" {
				require.Equal(suite.T(), t.expectedBody, resp.Body)
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
		suite.Run(t.name, func() {
			contentType, _, _, err := createMultipartRequestBody(t.formItems)
			require.True(suite.T(), t.wantError == (err != nil))
			require.True(suite.T(), strings.HasPrefix(contentType, t.expectedContentTypePrefix))
		})
	}
}

func (suite *RequestsTestSuite) TestMultipartDryRunOutput_IsPrettyPrinted() {
	buf := new(bytes.Buffer)
	client, err := NewKosliClient("", 1, false, logger.NewLogger(buf, buf, false))
	require.NoError(suite.T(), err)

	params := &RequestParams{
		Method: http.MethodPut,
		URL:    "http://localhost:8001/api/v2/attestations/test",
		DryRun: true,
		Token:  "test-token",
		Form: []FormItem{
			{
				Type:      "field",
				FieldName: "data_json",
				Content:   map[string]any{"key": "value", "nested": map[string]any{"a": 1}},
			},
		},
	}

	_, err = client.Do(params)
	require.NoError(suite.T(), err)

	output := buf.String()
	// Dry-run output should contain pretty-printed JSON with indentation
	require.Contains(suite.T(), output, "    \"key\": \"value\"", "dry-run output should pretty-print JSON fields")
	require.Contains(suite.T(), output, "    \"nested\": {", "dry-run output should pretty-print nested JSON")
}

func (suite *RequestsTestSuite) TestMultipartFieldJSON_IsCompact() {
	formItems := []FormItem{
		{
			Type:      "field",
			FieldName: "data_json",
			Content:   map[string]any{"key": "value", "nested": map[string]any{"a": 1}},
		},
	}
	_, body, jsonFields, err := createMultipartRequestBody(formItems)
	require.NoError(suite.T(), err)

	bodyStr := body.String()

	// The JSON in the multipart body must not contain indentation
	require.NotContains(suite.T(), bodyStr, "    ", "multipart JSON field should not contain indentation")

	// The raw JSON bytes stored for logging must be compact (single line, no newlines)
	rawJSON := string(jsonFields["data_json"].([]byte))
	require.NotContains(suite.T(), rawJSON, "\n", "multipart JSON field should be on a single line")
}

func (suite *RequestsTestSuite) TestNonMultipartJSON_IsCompact() {
	params := &RequestParams{
		Method: http.MethodPut,
		URL:    "https://example.com/api/v2/test",
		Token:  "test-token",
		Payload: map[string]any{
			"key":    "value",
			"nested": map[string]any{"a": 1},
		},
	}
	req, _, err := params.newHTTPRequest()
	require.NoError(suite.T(), err)

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(suite.T(), err)
	bodyStr := string(bodyBytes)

	// The JSON body must not contain indentation or newlines
	require.NotContains(suite.T(), bodyStr, "    ", "non-multipart JSON body should not contain indentation")
	require.NotContains(suite.T(), bodyStr, "\n", "non-multipart JSON body should be on a single line")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRequestsTestSuite(t *testing.T) {
	suite.Run(t, new(RequestsTestSuite))
}

func TestHTTPStatus(t *testing.T) {
	require.Equal(t, "HTTP 429 Too Many Requests", httpStatus(http.StatusTooManyRequests))
	require.Equal(t, "HTTP 599", httpStatus(599))
}

func TestJitteredBackoff(t *testing.T) {
	const minWait, maxWait = time.Second, 30 * time.Second

	t.Run("never waits less than the default backoff", func(t *testing.T) {
		for attempt := range 4 {
			wait := retryablehttp.DefaultBackoff(minWait, maxWait, attempt, nil)
			for range 1000 {
				got := jitteredBackoff(minWait, maxWait, attempt, nil)
				require.GreaterOrEqual(t, got, wait)
				require.LessOrEqual(t, got, wait+wait/2)
			}
		}
	})

	t.Run("spreads the waits of clients that failed together", func(t *testing.T) {
		seen := map[time.Duration]bool{}
		for range 100 {
			seen[jitteredBackoff(minWait, maxWait, 2, nil)] = true
		}
		require.Greater(t, len(seen), 1)
	})

	t.Run("never waits longer than the maximum", func(t *testing.T) {
		for range 1000 {
			require.Equal(t, maxWait, jitteredBackoff(minWait, maxWait, 10, nil))
		}
	})

	for _, resp := range []*http.Response{
		{StatusCode: http.StatusConflict, Header: http.Header{"Retry-After": []string{"7"}}},
		{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"soon"}}},
	} {
		t.Run(fmt.Sprintf("jitters a %d with a Retry-After the backoff ignores", resp.StatusCode), func(t *testing.T) {
			wait := retryablehttp.DefaultBackoff(minWait, maxWait, 1, nil)
			seen := map[time.Duration]bool{}
			for range 100 {
				got := jitteredBackoff(minWait, maxWait, 1, resp)
				require.GreaterOrEqual(t, got, wait)
				require.LessOrEqual(t, got, wait+wait/2)
				seen[got] = true
			}
			require.Greater(t, len(seen), 1)
		})
	}

	t.Run("keeps the wait the server asked for", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{"7"}},
		}
		require.Equal(t, 7*time.Second, jitteredBackoff(minWait, maxWait, 0, resp))
	})
}

func TestDoKeepsTheTransportErrorWhenRetriesRunOut(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	url := "http://" + listener.Addr().String() + "/closed/"
	require.NoError(t, listener.Close())

	client, err := NewKosliClient("", 1, false, logger.NewStandardLogger())
	require.NoError(t, err)
	_, err = client.Do(&RequestParams{Method: http.MethodGet, URL: url})

	require.Error(t, err)
	require.Contains(t, err.Error(), "giving up after 2 attempt(s): ")
	require.Contains(t, err.Error(), "connection refused")
	require.NotContains(t, err.Error(), "HTTP ")
}

// A misspelled form item type must fail rather than silently drop the part.
func TestCreateMultipartRequestBodyRejectsAnUnknownItemType(t *testing.T) {
	_, _, _, err := createMultipartRequestBody([]FormItem{
		{Type: "file_bytes", FieldName: "attachment_file", Content: "a/path"},
	})
	require.Error(t, err)
}

// The retry client logs every attempt with its full URL. For a pre-signed
// upload that URL carries the task's session token, which is blanked before
// the line is written; the rest of the query stays for debugging.
func TestCustomLoggerRedactsTheSecurityToken(t *testing.T) {
	var out bytes.Buffer
	cl := &CustomLogger{Logger: log.New(&out, "", 0)}

	cl.Printf("[DEBUG] %s %s", "PUT", "https://h/uploads/k?X-Amz-Credential=ASIA%2Fs3&X-Amz-Security-Token=IQoJsecret%2F%2F&X-Amz-Signature=abc")

	require.Equal(t, " PUT https://h/uploads/k?X-Amz-Credential=ASIA%2Fs3&X-Amz-Security-Token=<redacted>&X-Amz-Signature=abc\n", out.String())
}
