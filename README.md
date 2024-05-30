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
podman run --name ipam-db --net=host --env PGPASSWORD=cni -d ghcr.io/maiqueb/devconf2024/new-cni-verbs:latest
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

## Garbage collecting state resources
CNI spec 1.1 also provides a `GC` verb, intended to clean up stale cluster
resources (IP allocations for IPAM plugins, netfilter rules for firewalls,
etc).

Let's say for some reason the CNI DEL fails and the resource is not properly
cleaned up by CNI. An example would be the node died before actually invoking
the CNI DEL. Now, the runtime would check which pods are expected to run in
the system, and execute GC with the list of attachments it wants to keep. CNI
would then delete **everything else**.

To showcase this new verb, we must first simulate having some workloads; invoke
the following commands (CNI requires sudo / root):
```shell
demo-scenarios/add-ns.sh ipam-status pod1 192.168.200.1      
+ '[' 3 -ne 3 ']'
+ net_name=ipam-status
+ podname=pod1
+ IP=192.168.200.1
+ ifaceName=eth0
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
+ trap 'sudo ip netns del pod1' EXIT
+ sudo ip netns add pod1
+ sudo 'CNI_ARGS=IgnoreUnknown=true;IP=192.168.200.1/24' CNI_IFNAME=eth0 plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples add ipam-status pod1
2024-05-30T10:39:04.57111512+02:00 [info] INVOKED ADD
2024-05-30T10:39:04.571256995+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> []}   0xc0000262a0}
2024-05-30T10:39:04.572766189+02:00 [info] INVOKED IPAM ADD
2024-05-30T10:39:04.572914379+02:00 [info] ARGS: "IgnoreUnknown=true;IP=192.168.200.1/24"
{
    "cniVersion": "1.1.0",
    "ips": [
        {
            "address": "192.168.200.1/24",
            "interface": 0
        }
    ]
}+ sudo ip netns del pod1

demo-scenarios/add-ns.sh ipam-status pod2 192.168.200.22
+ '[' 3 -ne 3 ']'
+ net_name=ipam-status
+ podname=pod2
+ IP=192.168.200.22
+ ifaceName=eth0
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
+ trap 'sudo ip netns del pod2' EXIT
+ sudo ip netns add pod2
+ sudo 'CNI_ARGS=IgnoreUnknown=true;IP=192.168.200.22/24' CNI_IFNAME=eth0 plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples add ipam-status pod2
2024-05-30T10:39:09.171094015+02:00 [info] INVOKED ADD
2024-05-30T10:39:09.171232646+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> []}   0xc0000262a0}
2024-05-30T10:39:09.172682947+02:00 [info] INVOKED IPAM ADD
2024-05-30T10:39:09.172830597+02:00 [info] ARGS: "IgnoreUnknown=true;IP=192.168.200.22/24"
{
    "cniVersion": "1.1.0",
    "ips": [
        {
            "address": "192.168.200.22/24",
            "interface": 0
        }
    ]
}+ sudo ip netns del pod2

demo-scenarios/add-ns.sh ipam-status pod3 192.168.200.123
+ '[' 3 -ne 3 ']'
+ net_name=ipam-status
+ podname=pod3
+ IP=192.168.200.123
+ ifaceName=eth0
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
+ trap 'sudo ip netns del pod3' EXIT
+ sudo ip netns add pod3
+ sudo 'CNI_ARGS=IgnoreUnknown=true;IP=192.168.200.123/24' CNI_IFNAME=eth0 plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples add ipam-status pod3
2024-05-30T10:39:13.462724623+02:00 [info] INVOKED ADD
2024-05-30T10:39:13.462881683+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> []}   0xc0000282a0}
2024-05-30T10:39:13.464429958+02:00 [info] INVOKED IPAM ADD
2024-05-30T10:39:13.464584151+02:00 [info] ARGS: "IgnoreUnknown=true;IP=192.168.200.123/24"
{
    "cniVersion": "1.1.0",
    "ips": [
        {
            "address": "192.168.200.123/24",
            "interface": 0
        }
    ]
}+ sudo ip netns del pod3
```
Let's check the state of the `ips` table in the DB:
```shell
podman exec ipam-db psql -U cni -h 192.168.122.1 ipam -c "select * from ips;"
 id |            pod_id            | interface |         ip         |         created_on         
----+------------------------------+-----------+--------------------+----------------------------
  7 | cnitool-294e5a00691dfbb9d727 | eth0      | 192.168.200.1/24   | 2024-05-30 08:39:04.578071
  8 | cnitool-e1a2493cc355134dc89d | eth0      | 192.168.200.22/24  | 2024-05-30 08:39:09.178036
  9 | cnitool-bf4114a94da8d19d2b93 | eth0      | 192.168.200.123/24 | 2024-05-30 08:39:13.470398
(3 rows)
```
We should see 3 allocations - one for each "pod" we have created.

Let's now see GC in action. Let's first simulate everything is OK with our
workloads; for that, we indicate we want to preserve the allocations for the
three pods:
```shell
demo-scenarios/gc.sh ipam-status pod1 pod2 pod3        
+ '[' 4 -lt 1 ']'
+ net_name=ipam-status
+ shift
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
+ sudo plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples gc ipam-status pod1 pod2 pod3
2024-05-30T10:46:12.306765364+02:00 [info] INVOKED GC
2024-05-30T10:46:12.306948415+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> [{cnitool-294e5a00691dfbb9d727 eth0} {cnitool-e1a2493cc355134dc89d eth0} {cnitool-bf4114a94da8d19d2b93 eth0}]}   0xc0000a81e0}
2024-05-30T10:46:12.306968334+02:00 [info] read IPAM CONFIG: &{{ipam-status-cni} 192.168.122.1 5432 cni cni ipam}
2024-05-30T10:46:12.306972531+02:00 [info] read attachments to keep: [{cnitool-294e5a00691dfbb9d727 eth0} {cnitool-e1a2493cc355134dc89d eth0} {cnitool-bf4114a94da8d19d2b93 eth0}]
2024-05-30T10:46:12.308518681+02:00 [info] INVOKED IPAM GC
2024-05-30T10:46:12.308725673+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> [{cnitool-294e5a00691dfbb9d727 eth0} {cnitool-e1a2493cc355134dc89d eth0} {cnitool-bf4114a94da8d19d2b93 eth0}]}   0xc0000262a0}
2024-05-30T10:46:12.308755986+02:00 [info] valid attachment: {cnitool-294e5a00691dfbb9d727 eth0}
2024-05-30T10:46:12.308759927+02:00 [info] valid attachment: {cnitool-e1a2493cc355134dc89d eth0}
2024-05-30T10:46:12.308763016+02:00 [info] valid attachment: {cnitool-bf4114a94da8d19d2b93 eth0}
2024-05-30T10:46:12.315062979+02:00 [info] cachedEntry: {7 cnitool-294e5a00691dfbb9d727 eth0 192.168.200.1/24 2024-05-30T08:39:04.578071Z}
2024-05-30T10:46:12.31508065+02:00 [info] cachedEntry: {8 cnitool-e1a2493cc355134dc89d eth0 192.168.200.22/24 2024-05-30T08:39:09.178036Z}
2024-05-30T10:46:12.31508465+02:00 [info] cachedEntry: {9 cnitool-bf4114a94da8d19d2b93 eth0 192.168.200.123/24 2024-05-30T08:39:13.470398Z}
2024-05-30T10:46:12.315096209+02:00 [info] desired attachments: map[cnitool-294e5a00691dfbb9d727-eth0:{} cnitool-bf4114a94da8d19d2b93-eth0:{} cnitool-e1a2493cc355134dc89d-eth0:{}]
2024-05-30T10:46:12.315105039+02:00 [info] looking at attachment "cnitool-294e5a00691dfbb9d727-eth0"
2024-05-30T10:46:12.315108091+02:00 [info] looking at attachment "cnitool-e1a2493cc355134dc89d-eth0"
2024-05-30T10:46:12.31511034+02:00 [info] looking at attachment "cnitool-bf4114a94da8d19d2b93-eth0"
```

And let's check the state of the DB:
```shell
podman exec ipam-db psql -U cni -h 192.168.122.1 ipam -c "select * from ips;"
 id |            pod_id            | interface |         ip         |         created_on         
----+------------------------------+-----------+--------------------+----------------------------
  7 | cnitool-294e5a00691dfbb9d727 | eth0      | 192.168.200.1/24   | 2024-05-30 08:39:04.578071
  8 | cnitool-e1a2493cc355134dc89d | eth0      | 192.168.200.22/24  | 2024-05-30 08:39:09.178036
  9 | cnitool-bf4114a94da8d19d2b93 | eth0      | 192.168.200.123/24 | 2024-05-30 08:39:13.470398
(3 rows)
```

We should still see the same 3 allocations; let's now imagine pod2 died, and
CRI failed to invoke the corresponding CNI DEL:
```shell
demo-scenarios/gc.sh ipam-status pod1 pod3
+ '[' 3 -lt 1 ']'
+ net_name=ipam-status
+ shift
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
+ sudo plugins/cnitool --cni-bin-dir plugins/bin --cni-conf-dir examples gc ipam-status pod1 pod3
2024-05-30T10:52:53.697772352+02:00 [info] INVOKED GC
2024-05-30T10:52:53.697912831+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> [{cnitool-294e5a00691dfbb9d727 eth0} {cnitool-bf4114a94da8d19d2b93 eth0}]}   0xc0000262a0}
2024-05-30T10:52:53.697931431+02:00 [info] read IPAM CONFIG: &{{ipam-status-cni} 192.168.122.1 5432 cni cni ipam}
2024-05-30T10:52:53.697935396+02:00 [info] read attachments to keep: [{cnitool-294e5a00691dfbb9d727 eth0} {cnitool-bf4114a94da8d19d2b93 eth0}]
2024-05-30T10:52:53.699570429+02:00 [info] INVOKED IPAM GC
2024-05-30T10:52:53.699711694+02:00 [info] read configuration: &{{1.1.0 ipam-status status-cni map[] {} {[]  [] []} map[] <nil> [{cnitool-294e5a00691dfbb9d727 eth0} {cnitool-bf4114a94da8d19d2b93 eth0}]}   0xc0000ae1e0}
2024-05-30T10:52:53.699732914+02:00 [info] valid attachment: {cnitool-294e5a00691dfbb9d727 eth0}
2024-05-30T10:52:53.699735465+02:00 [info] valid attachment: {cnitool-bf4114a94da8d19d2b93 eth0}
2024-05-30T10:52:53.705327206+02:00 [info] cachedEntry: {1 cnitool-294e5a00691dfbb9d727 eth0 192.168.200.1/24 2024-05-30T08:52:38.284902Z}
2024-05-30T10:52:53.705341137+02:00 [info] cachedEntry: {2 cnitool-e1a2493cc355134dc89d eth0 192.168.200.22/24 2024-05-30T08:52:41.535088Z}
2024-05-30T10:52:53.70534523+02:00 [info] cachedEntry: {3 cnitool-bf4114a94da8d19d2b93 eth0 192.168.200.123/24 2024-05-30T08:52:45.429173Z}
2024-05-30T10:52:53.705361475+02:00 [info] desired attachments: map[cnitool-294e5a00691dfbb9d727-eth0:{} cnitool-bf4114a94da8d19d2b93-eth0:{}]
2024-05-30T10:52:53.705372017+02:00 [info] looking at attachment "cnitool-294e5a00691dfbb9d727-eth0"
2024-05-30T10:52:53.705376697+02:00 [info] looking at attachment "cnitool-e1a2493cc355134dc89d-eth0"
2024-05-30T10:52:53.708987344+02:00 [info] successfully deleted the existing allocation for "cnitool-e1a2493cc355134dc89d-eth0"
2024-05-30T10:52:53.70899233+02:00 [info] looking at attachment "cnitool-bf4114a94da8d19d2b93-eth0"
```

As we can see in the logs, one of the entries was deleted.

We can confirm that by checking the contents of the DB:
```shell
podman exec -t ipam-db psql -U cni -h 192.168.122.1 ipam -c "select * from ips;"
 id |            pod_id            | interface |         ip         |         created_on         
----+------------------------------+-----------+--------------------+----------------------------
  7 | cnitool-294e5a00691dfbb9d727 | eth0      | 192.168.200.1/24   | 2024-05-30 08:39:04.578071
  9 | cnitool-bf4114a94da8d19d2b93 | eth0      | 192.168.200.123/24 | 2024-05-30 08:39:13.470398
(2 rows)
```

## Conclusions
This repo contains demo scripts and code for two CNI plugins - a net config,
and an IPAM CNI plugins - for learning purposes.

Following the instructions in this README, the reader can learn about the two
new verbs introduced in
[CNI spec 1.1](https://github.com/containernetworking/cni/blob/v1.2.0/SPEC.md):
`STATUS` and `GC`.

In a real cluster, `STATUS` would be used by the runtime to know if the network
is ready to be used, whereas the `GC` would be used by the runtime to garbage
collect resources which have leaked in the system - e.g. resources which the
CNI plugin failed to cleanup on missed CNI DELs, or failed CNI ADDs.
