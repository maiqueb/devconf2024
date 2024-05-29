package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ipam"

	logging "github.com/k8snetworkplumbingwg/cni-log"

	"github.com/maiqueb/devconf2024/pkg/config"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

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
			GC:     cmdGC,
			Status: status,
		}, cniVersion.All,
		"Dummy CNI for learning purposes",
	)
}

func status(args *skel.CmdArgs) error {
	logging.Infof("INVOKED STATUS")
	netConf, err := loadNetConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed rendering plugin configuration: %w", err)
	}
	logging.Infof("read configuration: %v", netConf)

	if netConf.IPAMConfig != nil {
		if err := ipam.ExecStatus(netConf.IPAMConfig.Type, args.StdinData); err != nil {
			return err
		}
	}

	return nil
}

func cmdAdd(args *skel.CmdArgs) error {
	logging.Infof("INVOKED ADD")
	netConf, err := loadNetConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed rendering plugin configuration: %w", err)
	}
	logging.Infof("read configuration: %v", netConf)

	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{},
	}

	if netConf.IPAMConfig != nil {
		r, err := ipam.ExecAdd(netConf.IPAMConfig.Type, args.StdinData)
		if err != nil {
			return err
		}

		ipamResult, err := current.NewResultFromResult(r)
		if err != nil {
			return err
		}

		result.IPs = ipamResult.IPs
		result.Routes = ipamResult.Routes
		result.DNS = ipamResult.DNS
	}

	return types.PrintResult(result, netConf.CNIVersion)
}

func cmdGC(args *skel.CmdArgs) error {
	logging.Infof("INVOKED GC")
	netConf, err := loadNetConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed rendering plugin configuration: %w", err)
	}

	logging.Infof("read configuration: %v", netConf)
	logging.Infof("read IPAM CONFIG: %v", netConf.IPAMConfig)
	logging.Infof("read attachments to keep: %v", netConf.ValidAttachments)

	if err := ipam.ExecGC(netConf.IPAMConfig.Type, args.StdinData); err != nil {
		return err
	}

	return nil
}

func loadNetConf(bytes []byte) (*config.NetConf, error) {
	var conf *config.NetConf
	if err := json.Unmarshal(bytes, &conf); err != nil {
		return nil, err
	}
	return conf, nil
}
