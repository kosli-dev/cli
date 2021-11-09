package main

import (
	"fmt"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd, err := newRootCmd(os.Stdout, os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
