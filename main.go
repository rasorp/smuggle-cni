package main

import (
	"github.com/rasorp/smuggle-cni/internal/cni"
	"github.com/rasorp/smuggle-cni/internal/version"

	"github.com/containernetworking/cni/pkg/skel"
	cniv "github.com/containernetworking/cni/pkg/version"
)

func main() {
	skel.PluginMainFuncs(
		skel.CNIFuncs{
			Add:   cni.CommandAdd,
			Check: cni.CommandCheck,
			Del:   cni.CommandDel,
		},
		cniv.All,
		version.Get(),
	)
}
