package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
)

func CommandAdd(args *skel.CmdArgs) error {

	argsConfig, err := readCommandArgs(args.StdinData)
	if err != nil {
		return err
	}

	smugglePath := filepath.Join(smuggleAgentConfigDir, argsConfig.Name+".conf")

	smuggleConfig, err := readSmuggleSubnetConfig(smugglePath)
	if err != nil {
		return err
	}

	confListJSON, err := generateBridgeCNIConfList(smuggleConfig)
	if err != nil {
		return err
	}

	if err = writeCNIConf(args.ContainerID, smuggleCNIDataDir, confListJSON); err != nil {
		return err
	}

	result, err := invoke.DelegateAdd(context.TODO(), "bridge", confListJSON, nil)
	if err != nil {
		return fmt.Errorf("failed to delegate add: %w", err)
	}
	return result.Print()

}

func CommandDel(args *skel.CmdArgs) error {
	cleanup, netConfBytes, err := readCNIConf(args.ContainerID, smuggleCNIDataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	defer func() {
		cleanup(err)
	}()

	nc := &types.NetConf{}

	// We are deleting the network, so if we cannot parse the config, we cannot
	// proceed. Writing the error will give cluster operators more information
	// about what went wrong. In the end, a node reboot or manual intervention
	// may be required to clean up resources.
	if err = json.Unmarshal(netConfBytes, nc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse netconf: %v", err)
		return nil
	}

	return invoke.DelegateDel(context.TODO(), nc.Type, netConfBytes, nil)
}

func CommandCheck(args *skel.CmdArgs) error {
	argsConfig, err := readCommandArgs(args.StdinData)
	if err != nil {
		return err
	}

	smugglePath := filepath.Join(smuggleAgentConfigDir, argsConfig.Name+".conf")

	smuggleConfig, err := readSmuggleSubnetConfig(smugglePath)
	if err != nil {
		return err
	}

	confListJSON, err := generateBridgeCNIConfList(smuggleConfig)
	if err != nil {
		return err
	}

	// Delegate CHECK to the bridge plugin. The bridge plugin will verify that
	// the container's network interface exists and has the expected
	// configuration.
	return invoke.DelegateCheck(context.TODO(), "bridge", confListJSON, nil)
}

func writeCNIConf(containerID, dataDir string, conf []byte) error {

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(dataDir, containerID), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	n, err := f.Write(conf)
	if err == nil && n < len(conf) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func readCNIConf(containerID, dataDir string) (func(error), []byte, error) {
	path := filepath.Join(dataDir, containerID)

	// The cleanup function will trigger if no error occurs. As per the CNI spec,
	// ignore errors when cleaning up resources that may already be removed.
	cleanup := func(err error) {
		if err == nil {
			_ = os.Remove(path)
		}
	}
	netConfBytes, err := os.ReadFile(path)

	return cleanup, netConfBytes, err
}
