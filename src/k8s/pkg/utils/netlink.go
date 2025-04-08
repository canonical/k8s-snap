package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
)

type Vxlan struct {
	net.Interface
	Port int
}

func VxlanDevices() ([]Vxlan, error) {
	vxlanDevices := []Vxlan{}

	cmd := exec.Command("ip", "-d", "-j", "link", "show", "type", "vxlan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return vxlanDevices, fmt.Errorf("running ip command failed: %s", string(out))
	}

	var ipCmdLinks []struct {
		IfName   string `json:"ifname"`
		LinkInfo struct {
			InfoData struct {
				Port int `json:"port"`
			} `json:"info_data"`
		} `json:"linkinfo"`
	}

	if err := json.Unmarshal(out, &ipCmdLinks); err != nil {
		return vxlanDevices, fmt.Errorf("unmarshaling ip command output failed: %w", err)
	}

	for _, link := range ipCmdLinks {
		ifi, err := net.InterfaceByName(link.IfName)
		if err != nil {
			return vxlanDevices, fmt.Errorf("returning interface by name failed: %w", err)
		}
		vxlanDevices = append(vxlanDevices, Vxlan{
			Interface: *ifi,
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
