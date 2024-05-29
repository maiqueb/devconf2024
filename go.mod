module github.com/maiqueb/devconf2024

go 1.21.9

require (
	github.com/containernetworking/cni v1.2.0
	github.com/containernetworking/plugins v1.5.0
	github.com/k8snetworkplumbingwg/cni-log v0.0.0-20230801160229-b6e062c9e0f2
	github.com/lib/pq v1.10.9
)

require (
	github.com/coreos/go-iptables v0.7.0 // indirect
	github.com/safchain/ethtool v0.3.0 // indirect
	github.com/vishvananda/netlink v1.2.1-beta.2 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/sys v0.20.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/containernetworking/plugins => github.com/maiqueb/plugins v1.0.0-rc1.0.20240529145227-36d58b2dc9db
