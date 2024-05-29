#!/bin/bash -x

if [ "$#" -ne 1 ]; then
    echo "$0 <name of the network configuration to use>"
    exit 1
fi

net_name="$1"

export NETCONFPATH="${NETCONFPATH:-$(pwd)/examples}"
export CNI_PATH="${CNI_PATH:-$(pwd)/plugins/bin}"
export CNI_DIR=$(pwd)/plugins
export CNI_PATH="${CNI_PATH:-$CNI_DIR/bin}"
export CNITOOL_BINARY=$CNI_DIR/cnitool

trap "sudo ip netns del server" EXIT

# start the server "pod"
sudo ip netns add server

sudo plugins/cnitool \
    --cni-bin-dir plugins/bin \
    --cni-conf-dir examples \
        status $net_name
