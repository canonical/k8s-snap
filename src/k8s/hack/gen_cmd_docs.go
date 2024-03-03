package main

import (
	"log"

	"github.com/canonical/k8s/cmd/k8s"
	"github.com/spf13/cobra/doc"
)

func main() {
	k8sCmd := k8s.NewRootCmd()

	err := doc.GenMarkdownTree(k8sCmd, "../../../docs/src/_parts/commands")
	if err != nil {
		log.Fatal(err)
	}
}
