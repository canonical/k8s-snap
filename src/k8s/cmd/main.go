package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/cmd/k8s"
	k8s_dqlite "github.com/canonical/k8s/cmd/k8s-dqlite"
	"github.com/canonical/k8s/cmd/k8sd"
	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("k8s", k8s.Main)
	reexec.Register("k8s-dqlite", k8s_dqlite.Main)
	reexec.Register("k8sd", k8sd.Main)
}

func main() {
	os.Args[0] = filepath.Base(os.Args[0])
	if reexec.Init() {
		return
	}
	panic(fmt.Errorf("invalid entrypoint name %q", os.Args[0]))
}
