package calico

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
)

func CleanupNetwork(ctx context.Context, snap snap.Snap) error {
	output, err := exec.CommandContext(ctx, "ip", "-j", "link", "show").Output()
	if err != nil {
		return fmt.Errorf("failed to list network interfaces: %w", err)
	}

	// Parse the json output of `ip -j link show` to find the interfaces that were created by Calico.
	type NetworkInterface struct {
		Name string `json:"ifname"`
	}
	var interfaces []NetworkInterface
	err = json.Unmarshal(output, &interfaces)
	if err != nil {
		return fmt.Errorf("failed to parse network interface JSON: %w", err)
	}

	// Find the interfaces created by Calico
	for _, iface := range interfaces {
		// Check if the interface name matches the regex pattern
		match, err := regexp.MatchString("^vxlan[-v6]*.calico|cali[a-f0-9]*$", iface.Name)
		if err != nil {
			return fmt.Errorf("failed to match regex pattern: %w", err)
		}
		if match {
			// Perform cleanup for Calico interface
			if _, err := exec.CommandContext(ctx, "ip", "link", "delete", iface.Name).CombinedOutput(); err != nil {
				return fmt.Errorf("failed to delete interface %s: %w", iface.Name, err)
			}
		}
	}

	// List network namespaces in JSON format
	nsOutput, err := exec.CommandContext(ctx, "ip", "-j", "netns", "list").Output()
	if err != nil {
		return fmt.Errorf("failed to list network namespaces: %w", err)
	}

	// Parse the JSON output of `ip -j netns list` to find the namespaces that start with "cali-"
	type Namespace struct {
		Name string `json:"name"`
	}
	var namespaces []Namespace
	err = json.Unmarshal(nsOutput, &namespaces)
	if err != nil {
		return fmt.Errorf("failed to parse network namespace JSON: %w", err)
	}

	// Delete the namespaces that start with "cali-"
	for _, ns := range namespaces {
		if strings.HasPrefix(ns.Name, "cali-") {
			// Delete the namespace
			if _, err := exec.CommandContext(ctx, "ip", "netns", "delete", ns.Name).CombinedOutput(); err != nil {
				return fmt.Errorf("failed to delete network namespace %s: %w", ns.Name, err)
			}
		}
	}

	if _, err := exec.Command("bash", "-c", "iptables-legacy-save | grep -iv cali | iptables-legacy-restore").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove calico iptables rules: %w", err)
	}

	if _, err := exec.Command("bash", "-c", "iptables-save | grep -iv cali | iptables-restore").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove calico iptables rules: %w", err)
	}

	if _, err := exec.Command("bash", "-c", "ip6tables-legacy-save | grep -iv cali | ip6tables-legacy-restore").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove calico iptables rules: %w", err)
	}

	if _, err := exec.Command("bash", "-c", "ip6tables-save | grep -iv cali | ip6tables-restore").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove calico iptables rules: %w", err)
	}

	return nil
}
