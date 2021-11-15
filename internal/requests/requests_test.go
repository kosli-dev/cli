package requests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/maxcnunes/httpfake"
	"github.com/sirupsen/logrus"
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
		Get("/no-go/").
		Reply(404).
		BodyString("")
}

// shutdown the fake service after the suite execution
func (suite *RequestsTestSuite) TearDownSuite() {
	suite.fakeService.Close()
}

func (suite *RequestsTestSuite) TestSendPayload() {
	type args struct {
		payload    interface{}
		url        string
		token      string
		maxRetries int
		dryRun     bool
		method     string
	}

	for _, t := range []struct {
		name        string
		args        args
		expectError bool
		want        *HTTPResponse
	}{
		{
			name: "PUT request works",
			args: args{
				payload:    bytes.NewBuffer([]byte(`{"sha": "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f"}`)),
				url:        suite.fakeService.ResolveURL("/artifacts/1"),
				token:      "secret",
				maxRetries: 3,
				dryRun:     false,
				method:     http.MethodPut,
			},
			expectError: false,
			want: &HTTPResponse{
				Body:       `{"sha": "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f","name": "artifact"}`,
				StatusCode: 201,
			},
		},
		{
			name: "invalid JSON payload throws an error",
			args: args{
				payload:    func() string { return "string" },
				url:        suite.fakeService.ResolveURL("/artifacts/1"),
				token:      "secret",
				maxRetries: 3,
				dryRun:     false,
				method:     http.MethodPut,
			},
			expectError: true,
		},
		{
			name: "non-existing endpoint throws an error",
			args: args{
				payload:    bytes.NewBuffer([]byte(`{"sha": "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f"}`)),
				url:        suite.fakeService.ResolveURL("/no-go/"),
				token:      "secret",
				maxRetries: 3,
				dryRun:     false,
				method:     http.MethodGet,
			},
			expectError: true,
		},
		{
			name: "dry-run to non-existing endpoint does not throw an error",
			args: args{
				payload:    bytes.NewBuffer([]byte(`{"sha": "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f"}`)),
				url:        suite.fakeService.ResolveURL("/no-go/"),
				token:      "secret",
				maxRetries: 3,
				dryRun:     true,
				method:     http.MethodGet,
			},
			expectError: false,
			want:        nil,
		},
	} {
		suite.Run(t.name, func() {
			resp, err := SendPayload(t.args.payload, t.args.url, t.args.token, t.args.maxRetries, t.args.dryRun, t.args.method, logrus.New())
			if t.expectError {
				require.Errorf(suite.T(), err, "error was expected but got none")
			} else {
				require.NoErrorf(suite.T(), err, "error was not expected, but got: %v", err)
				require.Equal(suite.T(), t.want, resp, fmt.Sprintf("want: %v -- got: %v", t.want, resp))
				// require.Equal(suite.T(), t.want.StatusCode, resp.StatusCode, fmt.Sprintf("Status Code ** want: %v -- got: %v", t.want.StatusCode, resp.StatusCode))

			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRequestsTestSuite(t *testing.T) {
	suite.Run(t, new(RequestsTestSuite))
}
