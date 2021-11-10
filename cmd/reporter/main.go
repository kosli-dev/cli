package main

import (
	"os"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var log = logrus.New()

func main() {
	out := os.Stdout
	log.Out = out
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	cmd, err := newRootCmd(out, os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
