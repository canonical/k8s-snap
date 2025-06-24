package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
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
