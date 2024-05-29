# Devconf2024 - Teaching an old dog new tricks
Demo scripts / config for the Teaching an old dog new tricks talk at Devconf 2024.

## Requirements
- git
- a container runtime (docker / podman / ...)
- golang >= 1.21

## Setup demo environment
Execute the setup script provided in this repo:
```shell
./setup_dependencies.sh
```
This will get everything needed to run the demo.

The following executables are required and will be made available in the
following locations:
- cnitool: plugins/cnitool
- cni plugins: plugins/bin/...

To build these aforementioned resources:
```shell
make
```

## Demo context
This demo uses 2 types of plugins: a net plugin, and an IPAM plugin.

The net plugin is directly invoked by CNI tool, and our simple implementation
only parses the configuration and invoked the IPAM plugin. In a real
deployment - e.g. Kubernetes - the CNI plugin would be invoked by CRI, running
in Kubelet.

Our implementation of the IPAM plugin behaves like static IPAM (it "assigns"
the address requested by the user), and persists that allocation in a
relational DB deployed in the system.

## Checking if the network is ready
To check if the network is ready, CNI now provides the `STATUS` verb.

For that, let's just execute the helper script:
```shell
demo-scenarios/status.sh ipam-status
+ '[' 1 -ne 1 ']'
+ net_name=ipam-status
++ pwd
+ export NETCONFPATH=/home/mduarted/github/devconf2024/examples
+ NETCONFPATH=/home/mduarted/github/devconf2024/examples
++ pwd
+ export CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
+ CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
++ pwd
+ export CNI_DIR=/home/mduarted/github/devconf2024/plugins
+ CNI_DIR=/home/mduarted/github/devconf2024/plugins
+ export CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
+ CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
+ export CNITOOL_BINARY=/home/mduarted/github/devconf2024/plugins/cnitool
+ CNITOOL_BINARY=/home/mduarted/github/devconf2024/plugins/cnitool
+ trap 'sudo ip netns del server' EXIT
+ sudo ip netns add server
+ sudo plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples status ipam-status
Error: network ipam-status is not ready: dial tcp 192.168.122.1:5432: connect: connection refused
+ sudo ip netns del server
```

Seems the network is not ready to accept running containers ... It requires a
PG database listening in `192.168.122.1:5432`. Please note this IP address is
not magical - it is defined in the CNI configuration, located in
`examples/12-status-ipam.conflist`. You should replace that with an IP valid
for your environment.

... let's start it (we're also providing it...):
```shell
podman run --name ipam-db --net=host -d ghcr.io/maiqueb/devconf2024/new-cni-verbs:latest
<some UUID>
```

And let's repeat the command (after a few seconds ...):
```shell
demo-scenarios/status.sh ipam-status
+ '[' 1 -ne 1 ']'
+ net_name=ipam-status
++ pwd
+ export NETCONFPATH=/home/mduarted/github/devconf2024/examples
+ NETCONFPATH=/home/mduarted/github/devconf2024/examples
++ pwd
+ export CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
+ CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
++ pwd
+ export CNI_DIR=/home/mduarted/github/devconf2024/plugins
+ CNI_DIR=/home/mduarted/github/devconf2024/plugins
+ export CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
+ CNI_PATH=/home/mduarted/github/devconf2024/plugins/bin
+ export CNITOOL_BINARY=/home/mduarted/github/devconf2024/plugins/cnitool
+ CNITOOL_BINARY=/home/mduarted/github/devconf2024/plugins/cnitool
+ trap 'sudo ip netns del server' EXIT
+ sudo ip netns add server
+ sudo plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples status ipam-status
2024-05-29T19:22:19.046838026+02:00 [info] INVOKED STATUS
2024-05-29T19:22:19.047003389+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> []}   0xc0000262a0}
2024-05-29T19:22:19.048630054+02:00 [info] INVOKED IPAM STATUS
2024-05-29T19:22:19.048774829+02:00 [info] read configuration: &{{ipam-status-cni} 192.168.122.1 5432 cni cni ipam}
Network ipam-status is ready for ADD requests
+ sudo ip netns del server
```

As we can see, now that we have the DB available, the STATUS reports the plugin as ready.