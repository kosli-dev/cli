package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CreateFlowCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateFlowCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateFlowCommandTestSuite) TestCreateFlowCmd() {
	deprecationWarning := "[warning] creating a flow without --template-file or --use-empty-template uses a deprecated API endpoint and will stop working in a future release; please provide a template\n"
	visibilityDeprecationNotice := "Flag --visibility has been deprecated, this flag is deprecated and will be removed in a future version.\n"
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       "create flow newFlow xxx" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError:   true,
			name:        "fails when name is considered invalid by the server",
			cmd:         "create flow 'foo bar'" + suite.defaultKosliArguments,
			goldenRegex: "Error: .*foo bar",
		},
		{
			name:   "can create a flow (by default legacy template is used)",
			cmd:    "create flow newFlow --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: deprecationWarning + "flow 'newFlow' was created\n",
		},
		{
			name:   "re-creating a flow updates its metadata",
			cmd:    "create flow newFlow --description \"changed description\" " + suite.defaultKosliArguments,
			golden: deprecationWarning + "flow 'newFlow' was updated\n",
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd:       "create flow newFlow --description \"my new flow\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --org is not set\nUsage: kosli create flow FLOW-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "create flow newFlow --description \"my new flow\" --org cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli create flow FLOW-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing name argument fails",
			cmd:       "create flow --description \"my new flow\" -H http://localhost:8001 --org cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "cannot use --template and --template-file together",
			cmd:       "create flow newFlow --description \"my new flow\" --template foo --template-file testdata/valid_template.yml" + suite.defaultKosliArguments,
			golden:    "Error: only one of --template, --template-file is allowed\n",
		},
		{
			name:   "deprecated --visibility flag is accepted (no longer illegal) and warns",
			cmd:    "create flow newFlowWithVisibility --visibility public --use-empty-template --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: visibilityDeprecationNotice + "flow 'newFlowWithVisibility' was created\n",
		},
		// flows v2
		{
			name:   "can create a flow with a valid template",
			cmd:    "create flow newFlowWithTemplate --template-file testdata/valid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithTemplate' was created\n",
		},
		{
			name:   "re-creating a flow (with template) updates its metadata",
			cmd:    "create flow newFlowWithTemplate --template-file testdata/valid_template.yml --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithTemplate' was updated\n",
		},
		{
			wantError:   true,
			name:        "creating a flow with an invalid template fails",
			cmd:         "create flow newFlowWithTemplate --template-file testdata/invalid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			goldenRegex: "Error: Input payload validation failed.*",
		},
		{
			wantError: true,
			name:      "fails when both --template-file and --use-empty-template are provided",
			cmd:       "create flow newFlowWithTemplate --use-empty-template --template-file testdata/valid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			golden:    "Error: only one of --template-file, --use-empty-template is allowed\n",
		},
		{
			name:   "creating a flow with --use-empty-template works",
			cmd:    "create flow newFlowWithEmptyTemplate --use-empty-template --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithEmptyTemplate' was created\n",
		},
		{
			wantError: true,
			name:      "fails when --template-file is passed as empty string",
			cmd:       "create flow newFlow --template-file \"\" --description \"my new flow\" " + suite.defaultKosliArguments,
			golden:    "Error: flag '--template-file' was given an empty value\n",
		},
		{
			name:   "a --description of [] is a real value, not an empty one",
			cmd:    "create flow newFlowWithBracketsDescription --use-empty-template --description \"[]\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithBracketsDescription' was created\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestCreateFlowRejectsEmptyTemplateElement pins that an empty --template
// element is refused rather than dropped. --template names the attestations a
// flow requires, so an element lost on the way in weakens the flow's template
// for every artifact that passes through it afterwards, long after the run that
// caused it. `-t "$COVERAGE" -t unit-test` with COVERAGE unset is the shape
// that does it.
//
// This is deliberately not part of CreateFlowCommandTestSuite: the rejection
// happens while flags are parsed, before any request, so it needs no server.
func TestCreateFlowRejectsEmptyTemplateElement(t *testing.T) {
	_, _, _, _, err := executeCommandC(
		`create flow myflow -t "" -t unit-test --org demo --api-token DRY_RUN --dry-run`)

	require.Error(t, err)
	require.ErrorContains(t, err, "template")
}

// TestCreateFlowRejectsEmptyTemplateElementFromEnv pins the refusal on the
// environment path as well as argv. An environment variable cannot be repeated,
// so a comma list is the only way to name several required attestations that
// way, and `KOSLI_TEMPLATE="$COVERAGE,unit-test"` with COVERAGE unset is the
// shape that drops one.
//
// --template and --attachments share the type, so this holds transitively from
// the attachments env test. It is here so that createFlow says so itself,
// rather than leaving a reader to find the guarantee in another command's test.
func TestCreateFlowRejectsEmptyTemplateElementFromEnv(t *testing.T) {
	t.Setenv("KOSLI_TEMPLATE", ",unit-test")

	_, _, _, _, err := executeCommandC(
		`create flow myflow --org demo --api-token DRY_RUN --dry-run`)

	require.Error(t, err)
	require.ErrorContains(t, err, "template")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateFlowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateFlowCommandTestSuite))
}
