package main

import (
	"bytes"
	"testing"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
)

// cmdTestCase describes a cmd test case.
type cmdTestCase struct {
	name      string
	cmd       string
	golden    string
	wantError bool
}

// executeCommandStdinC executes a command as a user would and return the output
func executeCommandC(cmd string) (*cobra.Command, string, error) {
	args, err := shellwords.Parse(cmd)
	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)

	root, err := newRootCmd(buf, args)
	if err != nil {
		return nil, "", err
	}

	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err := root.ExecuteC()
	output := buf.String()

	return c, output, err
}

// runTestCmd runs a table of cmd test cases
func runTestCmd(t *testing.T, tests []cmdTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("running cmd: %s", tt.cmd)
			_, out, err := executeCommandC(tt.cmd)
			if (err != nil) != tt.wantError {
				t.Errorf("error expectation not matched\n\n WANT error is: %t\n\n but GOT: '%v'", tt.wantError, err)
			}
			if tt.golden != "" {
				if !bytes.Equal([]byte(tt.golden), []byte(out)) {
					t.Errorf("does not match golden\n\nWANT:\n'%s'\n\nGOT:\n'%s'\n", tt.golden, out)
				}
			}
		})
	}
}
