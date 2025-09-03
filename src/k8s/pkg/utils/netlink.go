package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"regexp"
)

type VXLANInterface struct {
	net.Interface
	Port *int
}

var ipLinks []struct {
	IfName   string `json:"ifname"`
	LinkInfo struct {
		InfoData struct {
			Port *int `json:"port"`
		} `json:"info_data"`
	} `json:"linkinfo"`
}

func ListVXLANInterfaces() ([]VXLANInterface, error) {
	vxlanDevices := []VXLANInterface{}

	cmd := exec.Command("ip", "-d", "-j", "link", "list", "type", "vxlan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return vxlanDevices, fmt.Errorf("running ip command failed: %s", string(out))
	}

	out = fixInvalidIproute2JSON(out)

	if err := json.Unmarshal(out, &ipLinks); err != nil {
		return vxlanDevices, fmt.Errorf("unmarshaling ip command output failed: %w", err)
	}

	for _, link := range ipLinks {

		// running ip -d -j link show
		if link.IfName == "" {
			continue
		}

		iface, err := net.InterfaceByName(link.IfName)
		if err != nil {
			return vxlanDevices, fmt.Errorf("returning interface by name failed: %w", err)
		}
		vxlanDevices = append(vxlanDevices, VXLANInterface{
			Interface: *iface,
			Port:      link.LinkInfo.InfoData.Port,
		})
	}

	return vxlanDevices, nil
}

func RemoveLink(name string) error {
	cmd := exec.Command("ip", "link", "delete", name)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("running ip command failed: %s", string(out))
	}

	return nil
}

// fixInvalidIproute2JSON cleans up invalid JSON output produced by the
// iproute2 command in arm64. Currently, the Ubuntu package combines the VXLAN
// VNI value with the fan-map extension, resulting in invalid JSON.
func fixInvalidIproute2JSON(input []byte) []byte {
	output := string(input)

	// Target the specific case where numeric VXLAN VNI values are concatenated
	// with text (like "0fan-map") without proper JSON quoting
	re := regexp.MustCompile(`"id":\s*([0-9]+[a-zA-Z][a-zA-Z0-9_-]*)`)

	// Enclose the entire VNI value in double quotes
	output = re.ReplaceAllString(output, `"id":"$1"`)
	return []byte(output)
}
