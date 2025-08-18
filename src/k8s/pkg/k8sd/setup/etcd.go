package setup

import (
	"fmt"
	"maps"
	"net"
	"path/filepath"
	"slices"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

func Etcd(snap snap.Snap, name string, nodeIP net.IP, clientPort, peerPort int, initialClusterMembers map[string]string, extraArgs map[string]*string) error {
	listenUrls := []string{
		fmt.Sprintf("https://%s", utils.JoinHostPort(nodeIP.String(), clientPort)),
	}

	localhostAddress, err := utils.GetLocalhostAddress()
	if err != nil {
		return fmt.Errorf("failed to get localhost address: %w", err)
	}

	if nodeIP != nil && !nodeIP.IsLoopback() {
		listenUrls = append(listenUrls, fmt.Sprintf("https://%s", utils.JoinHostPort(localhostAddress.String(), clientPort)))
	}

	listenClientURLs := strings.Join(listenUrls, ",")

	peerURL := fmt.Sprintf("https://%s", utils.JoinHostPort(nodeIP.String(), peerPort))

	advertiseClientURLs := fmt.Sprintf("https://%s", utils.JoinHostPort(nodeIP.String(), clientPort))

	clusterState := "new"
	if len(initialClusterMembers) > 0 {
		clusterState = "existing"
	}

	if initialClusterMembers == nil {
		initialClusterMembers = make(map[string]string)
	}

	initialClusterMembers[name] = peerURL

	var initialCluster []string

	for _, memberName := range slices.Sorted(maps.Keys(initialClusterMembers)) {
		initialCluster = append(initialCluster, fmt.Sprintf("%s=%s", memberName, initialClusterMembers[memberName]))
	}

	args := map[string]string{
		"--data-dir":                    filepath.Join(snap.EtcdDir(), "data"),
		"--name":                        name,
		"--initial-advertise-peer-urls": peerURL,
		"--listen-peer-urls":            peerURL,
		"--listen-client-urls":          listenClientURLs,
		"--advertise-client-urls":       advertiseClientURLs,
		"--initial-cluster-state":       string(clusterState),
		"--initial-cluster":             strings.Join(initialCluster, ","),
		"--client-cert-auth":            "true",
		"--trusted-ca-file":             filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
		"--cert-file":                   filepath.Join(snap.EtcdPKIDir(), "server.crt"),
		"--key-file":                    filepath.Join(snap.EtcdPKIDir(), "server.key"),
		"--peer-client-cert-auth":       "true",
		"--peer-trusted-ca-file":        filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
		"--peer-cert-file":              filepath.Join(snap.EtcdPKIDir(), "peer.crt"),
		"--peer-key-file":               filepath.Join(snap.EtcdPKIDir(), "peer.key"),
		"--auto-tls":                    "false",
		"--peer-auto-tls":               "false",
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "etcd", args, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "etcd", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write extra arguments: %w", err)
	}
	return nil
}
