package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	cni "github.com/containernetworking/cni/pkg/types"
)

type NetConf struct {
	cni.NetConf

	VeryImportantParam string    `json:"very-important-param"`
	IrrelevantParam    string    `json:"irrelevant-param,omitempty"`
	IPAMConfig         *IPAMConf `json:"ipam"`
}

func (nc *NetConf) UnmarshalJSON(b []byte) error {
	type NetConfAlias NetConf
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()

	var confAlias *NetConfAlias
	if err := decoder.Decode(&confAlias); err != nil {
		return err
	}

	*nc = NetConf(*confAlias)
	return nil
}

type IPAMConf struct {
	cni.IPAM

	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
}

func (nc *IPAMConf) SqlConnection() string {
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		nc.Host, nc.Port, nc.User, nc.Password, nc.DbName)
}
