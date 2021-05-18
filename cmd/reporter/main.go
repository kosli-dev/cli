package main

import (
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd, err := newRootCmd(os.Stdout, os.Args[1:])
	handleError(err)
	_ = cmd.Execute()
}
