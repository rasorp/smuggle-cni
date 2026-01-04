package cni

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/types"
)

const (
	smuggleAgentConfigDir = "/opt/smuggle/config/"
	smuggleCNIDataDir     = "/var/lib/cni/smuggle"
)

type SmuggleCNIConfig struct {
	Name   string                `json:"name"`
	Bridge string                `json:"bridge"`
	MTU    int                   `json:"mtu"`
	IPMasq bool                  `json:"ipmasq"`
	IPv4   *SmuggleCNIIPv4Config `json:"ipv4"`
}

type SmuggleCNIIPv4Config struct {
	Network string `json:"network"`
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
}

type CNIConflist struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Bridge    string `json:"bridge,omitempty"`
	IPMasq    bool   `json:"ipMasq"`
	IsGateway bool   `json:"isGateway"`
	MTU       int    `json:"mtu"`

	IPAM *CNIConflistIPAM `json:"ipam,omitempty"`
}

type CNIConflistIPAM struct {
	Type    string    `json:"type"`
	Ranges  [][]Range `json:"ranges"`
	Routes  []Route   `json:"routes"`
	DataDir string    `json:"dataDir,omitempty"`
}

type Range struct {
	Subnet string `json:"subnet"`
	GW     string `json:"gw,omitempty"`
}

type Route struct {
	Dst string `json:"dst"`
	GW  string `json:"gw,omitempty"`
}

func generateBridgeCNIConfList(cfg *SmuggleCNIConfig) ([]byte, error) {

	clist := CNIConflist{
		Type:      "bridge",
		Name:      cfg.Name,
		Bridge:    cfg.Bridge,
		IPMasq:    cfg.IPMasq,
		IsGateway: true,
		MTU:       cfg.MTU,
		IPAM: &CNIConflistIPAM{
			Type: "host-local",
			Ranges: [][]Range{
				{
					{
						Subnet: normalizeSubnet(cfg.IPv4.Subnet),
						GW:     cfg.IPv4.Gateway,
					},
				},
			},
			Routes: []Route{
				{
					Dst: cfg.IPv4.Network,
				},
				{
					Dst: "0.0.0.0/0",
					GW:  cfg.IPv4.Gateway,
				},
			},
			DataDir: smuggleCNIDataDir,
		},
	}

	// Marshal to JSON with indentation
	confListBytes, err := json.Marshal(&clist)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CNI conflist: %w", err)
	}

	return confListBytes, nil
}

// normalizeSubnet takes a CIDR string and returns it with host bits cleared. If
// the CIDR is invalid, it returns the original string.
func normalizeSubnet(cidr string) string {
	if _, ipNet, err := net.ParseCIDR(cidr); err != nil {
		return cidr
	} else {
		return ipNet.String()
	}
}

type NetConf struct {
	types.NetConf
}

func readCommandArgs(b []byte) (*NetConf, error) {
	n := NetConf{}

	if err := json.Unmarshal(b, &n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}

	return &n, nil
}

func readSmuggleSubnetConfig(p string) (*SmuggleCNIConfig, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config SmuggleCNIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}
