#!/bin/bash -x

if [ "$#" -lt 1 ]; then
    echo "$0 <name of the network configuration to use> [attachment to keep ...]"
    exit 1
fi

net_name="$1"
shift

export NETCONFPATH="${NETCONFPATH:-$(pwd)/examples}"
export CNI_PATH="${CNI_PATH:-$(pwd)/plugins/bin}"
export CNI_DIR=$(pwd)/plugins
export CNI_PATH="${CNI_PATH:-$CNI_DIR/bin}"
export CNITOOL_BINARY=$CNI_DIR/cnitool

sudo plugins/cnitool \
    --cni-bin-dir plugins/bin \
    --cni-conf-dir examples \
        gc "$net_name" "$@"
