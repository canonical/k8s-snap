package main

import (
	"fmt"
	"os"
	"path/filepath"

	k8s_dqlite "github.com/canonical/k8s/cmd/k8s-dqlite"
	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("k8s-dqlite", k8s_dqlite.Main)
}

func main() {
	os.Args[0] = filepath.Base(os.Args[0])
	if reexec.Init() {
		return
	}
	panic(fmt.Errorf("invalid entrypoint name %q", os.Args[0]))
}
