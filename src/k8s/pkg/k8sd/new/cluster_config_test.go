package newtypes_test

import (
	"encoding/json"
	"fmt"
	"testing"

	newtypes "github.com/canonical/k8s/pkg/k8sd/new"
)

func TestClusterConfig(t *testing.T) {
	var c newtypes.ClusterConfig

	b, _ := json.Marshal(c)
	fmt.Println(string(b))
	// t.FailNow()
}
