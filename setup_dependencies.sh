#!/usr/bin/env bash

set -o xtrace

unset LANG
unset LANGUAGE
LC_ALL=en_us.utf8
export LC_ALL

umask 022

PATH=$PATH:/usr/local/bin:/usr/local/sbin:/usr/sbin:/sbin
TOP_DIR=$(cd $(dirname "$0") && pwd)

function install_go {
    if [[ -x $(which go) ]]; then
        echo "Go already installed, skipping"
    else
        local file_url="https://go.dev/dl/go1.21.9.linux-amd64.tar.gz"
        local tempdir
        tempdir=$(mktemp -d)
        file_name=$(basename "$file_url")
        wget --progress=dot:giga -t 2 -c $file_url -O $tempdir/$file_name
        if [[ $? -ne 0 ]]; then
            die "$file_url could not be downloaded"
        fi
        echo $file_name
        sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf $tempdir/$file_name
        export PATH=$PATH:/usr/local/go/bin
        go version
        return
    fi
}

function git_clone {
    local git_remote=$1
    local git_dest=$2
    local orig_dir
    orig_dir=$(pwd)

    git clone $git_remote $git_dest
    cd $git_dest
    git show --oneline | head -1
    cd $orig_dir
}

function install_cni_plugins {
    local cni_plugins_repo="https://github.com/containernetworking/plugins.git"
    local cni_plugins_dir
    if [[ -d plugins ]]; then
        echo "CNI plugins already installed, skipping"
    else
        cni_plugins_dir=$(mktemp -d)
        git_clone $cni_plugins_repo $cni_plugins_dir
        mkdir -p $TOP_DIR/plugins
        cd $cni_plugins_dir
        git fetch
        git checkout v1.4.1
        ./build_linux.sh
        mv $cni_plugins_dir/bin $TOP_DIR/plugins
        cd $TOP_DIR
        rm -rf $cni_plugins_dir
    fi
}

function install_cnitool {
    local cnitool_repo="https://github.com/containernetworking/cni"
    local cnitool_dir
    if [[ -x $TOP_DIR/plugins/cnitool ]]; then
        echo "cnitool already installed, skipping"
    else
        cnitool_dir=$(mktemp -d)
        git_clone $cnitool_repo $cnitool_dir
        cd $cnitool_dir

        # MUST BUILD CNI TOOL FROM A SPECIFIC PR (implements GC API)
        git fetch origin "pull/1076/head:gc-verb-cni-tool"
        git checkout gc-verb-cni-tool

        go build -o $TOP_DIR/plugins/cnitool cnitool/main.go
        cd $TOP_DIR
        rm -rf $cnitool_dir
    fi
}

install_go
install_cni_plugins
install_cnitool
