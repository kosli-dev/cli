package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type TagTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	envName               string
	envType               string
}

func (suite *TagTestSuite) SetupTest() {
	suite.flowName = "tag-flow"
	suite.envName = "tag-env"
	suite.envType = "K8S"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.Suite.T())
	CreateEnv(global.Org, suite.envName, suite.envType, suite.Suite.T())
}

func (suite *TagTestSuite) TestTagCmd() {
	tests := []cmdTestCase{
		{
			name:   "can tag an environment with one tag",
			cmd:    fmt.Sprintf("tag environment %s --set foo=bar %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] added for environment 'tag-env'\n",
		},
		{
			name:   "can tag an environment with multiple tags",
			cmd:    fmt.Sprintf("tag environment %s --set foo=bar --set key=value %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo, key] added for environment 'tag-env'\n",
		},
		{
			name:   "can remove tags from an environment",
			cmd:    fmt.Sprintf("tag environment %s --unset foo %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] removed for environment 'tag-env'\n",
		},
		{
			name:   "can add and remove tags from an environment at the same time",
			cmd:    fmt.Sprintf("tag environment %s --set key=value --unset foo %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [key] added, and Tag(s) [foo] removed for environment 'tag-env'\n",
		},
		{
			name:   "removing a non-existing tag is okay",
			cmd:    fmt.Sprintf("tag environment %s --unset non-existing %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [non-existing] removed for environment 'tag-env'\n",
		},
		{
			name:   "can tag a flow",
			cmd:    fmt.Sprintf("tag flow %s --set foo=bar %s", suite.flowName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] added for flow 'tag-flow'\n",
		},
	}
	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTagTestSuite(t *testing.T) {
	suite.Run(t, new(TagTestSuite))
}
