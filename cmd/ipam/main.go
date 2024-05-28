package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"net"
	"strconv"
	"strings"

	logging "github.com/k8snetworkplumbingwg/cni-log"

	_ "github.com/lib/pq"

	"github.com/maiqueb/devconf2024/pkg/config"
)

func main() {
	skel.PluginMainFuncs(
		skel.CNIFuncs{
			Add: cmdAdd,
			Check: func(_ *skel.CmdArgs) error {
				return nil
			},
			Del: func(_ *skel.CmdArgs) error {
				return nil
			},
			GC: func(_ *skel.CmdArgs) error {
				return nil
			},
			Status: status,
		}, cniVersion.All,
		"Dummy IPAM CNI for learning purposes",
	)
}

func status(args *skel.CmdArgs) error {
	logging.Infof("INVOKED IPAM STATUS")
	netConf, err := loadIPAMConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed rendering plugin configuration: %w", err)
	}

	ipamConf := netConf.IPAMConfig
	logging.Infof("read configuration: %v", ipamConf)

	db, err := sql.Open("postgres", ipamConf.SqlConnection())
	if err != nil {
		return logging.Errorf("read configuration: %v", ipamConf)
	}
	defer func() {
		_ = db.Close()
	}()

	return db.Ping()
}

func cmdAdd(args *skel.CmdArgs) error {
	netConf, err := loadIPAMConf(args.StdinData)
	if err != nil {
		return err
	}

	ipamConf := netConf.IPAMConfig
	db, err := sql.Open("postgres", ipamConf.SqlConnection())
	if err != nil {
		return logging.Errorf("read configuration: %v", ipamConf)
	}
	defer func() {
		_ = db.Close()
	}()

	logging.Infof("ARGS: %q", args.Args)

	podUID := args.ContainerID
	ip, err := extractCNIArgsIP(args.Args)
	if err != nil {
		return fmt.Errorf("error parsing the CNI args %q: %w", args.Args, err)
	}

	if _, err = db.ExecContext(context.Background(), persistIPQuery(), podUID, args.IfName, ip); err != nil {
		return fmt.Errorf("error persisting the IP address: %w", err)
	}

	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{},
		IPs:        []*current.IPConfig{buildIPConfig(ip)},
		Routes:     nil,
		DNS:        types.DNS{},
	}

	return types.PrintResult(result, netConf.CNIVersion)
}

func loadIPAMConf(bytes []byte) (*config.NetConf, error) {
	n := config.NetConf{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, err
	}

	if n.IPAMConfig == nil {
		return nil, fmt.Errorf("IPAM config missing 'ipam' key")
	}

	return &n, nil
}

func extractCNIArgsIP(envArgs string) (string, error) {
	var ip string

	splitEnvArgs := strings.Split(envArgs, ";")
	for _, splitEnvArg := range splitEnvArgs {
		kvs := strings.Split(splitEnvArg, "=")
		if len(kvs) != 2 {
			return "", fmt.Errorf("invalid env var: %q", splitEnvArg)
		}
		if kvs[0] == "IP" {
			ip = kvs[1]
		}
	}
	return ip, nil
}

func buildIPConfig(ipWithSubnet string) *current.IPConfig {
	ipAndSubnetMask := strings.Split(ipWithSubnet, "/")
	ip := ipAndSubnetMask[0]
	mask := ipAndSubnetMask[1]
	numberOfOnes, err := strconv.Atoi(mask)
	if err != nil {
		return nil
	}
	iface := 0
	return &current.IPConfig{
		Interface: &iface,
		Address: net.IPNet{
			IP:   net.ParseIP(ip),
			Mask: net.CIDRMask(numberOfOnes, 32), // hardcode
		},
		Gateway: nil,
	}
}

func persistIPQuery() string {
	return `insert into ips(pod_id, interface, ip) values($1, $2, $3)`
}
