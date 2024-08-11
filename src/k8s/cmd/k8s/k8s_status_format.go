package k8s

import (
	"fmt"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

type ClusterStatus apiv1.ClusterStatus

// TICS -COV_GO_SUPPRESSED_ERROR
// we are just formatting the output for the k8s status command, it is ok to ignore failures from result.WriteString()

func (c ClusterStatus) isHA() bool {
	voters := 0
	for _, member := range c.Members {
		if member.DatastoreRole == apiv1.DatastoreRoleVoter {
			voters++
		}
	}
	return voters > 2
}

// TODO: Print k8s version. However, multiple nodes can run different version, so we would need to query all nodes.
func (c ClusterStatus) String() string {
	result := strings.Builder{}

	// Status
	if c.Ready {
		result.WriteString(fmt.Sprintf("%-25s %s", "cluster status:", "ready"))
	} else {
		result.WriteString(fmt.Sprintf("%-25s %s", "cluster status:", "not ready"))
	}
	result.WriteString("\n")

	// Control Plane Nodes
	result.WriteString(fmt.Sprintf("%-25s ", "control plane nodes:"))
	if len(c.Members) > 0 {
		members := make([]string, 0, len(c.Members))
		for _, m := range c.Members {
			members = append(members, fmt.Sprintf("%s (%s)", m.Address, m.DatastoreRole))
		}
		result.WriteString(strings.Join(members, ", "))
	} else {
		result.WriteString("none")
	}
	result.WriteString("\n")

	// High availability
	result.WriteString(fmt.Sprintf("%-25s ", "high availability:"))
	if c.isHA() {
		result.WriteString("yes")
	} else {
		result.WriteString("no")
	}
	result.WriteString("\n")

	// Datastore
	// TODO: how to understand if the ds is running or not?
	if c.Datastore.Type != "" {
		result.WriteString(fmt.Sprintf("%-25s %s\n", "datastore:", c.Datastore.Type))
	} else {
		result.WriteString(fmt.Sprintf("%-25s %s\n", "datastore:", "disabled"))
	}

	result.WriteString(fmt.Sprintf("%-25s %s\n", "network:", c.Network))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "dns:", c.DNS))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "ingress:", c.Ingress))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "load-balancer:", c.LoadBalancer))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "local-storage:", c.LocalStorage))
	result.WriteString(fmt.Sprintf("%-25s %s", "gateway", c.Gateway))

	return result.String()
}

// TICS +COV_GO_SUPPRESSED_ERROR
