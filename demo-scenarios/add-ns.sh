#!/bin/bash -x

if [ "$#" -ne 3 || "$#" -ne 4 ]; then
    echo "$0 <name of the network configuration to use> <pod name> <IP> [interface name]"
    exit 1
fi

net_name="$1"
podname="$2"
IP="$3"
ifaceName="${4:-eth0}"

export NETCONFPATH="${NETCONFPATH:-$(pwd)/examples}"
export CNI_PATH="${CNI_PATH:-$(pwd)/plugins/bin}"
export CNI_DIR=$(pwd)/plugins
export CNI_PATH="${CNI_PATH:-$CNI_DIR/bin}"
export CNITOOL_BINARY=$CNI_DIR/cnitool

trap "sudo ip netns del $podname" EXIT

# start the server "pod"
sudo ip netns add $podname

sudo CNI_ARGS="IgnoreUnknown=true;IP=$IP/24" \
    CNI_IFNAME="$ifaceName" \
    plugins/cnitool --cni-bin-dir plugins/bin \
    --cni-conf-dir examples \
        add $net_name "$podname"
