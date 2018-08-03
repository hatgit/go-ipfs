package iptbutil

import (
	"gx/ipfs/QmYyFh6g1C9uieTpH8CR8PpWBUQjvMDJTsRhJWx5qkXy39/go-ipfs-config"
)

// IpfsNode defines the interface iptb requires to work
// with an IPFS node
type IpfsNode interface {
	Init() error
	Kill() error
	Start(args []string) error
	APIAddr() (string, error)
	GetPeerID() string
	RunCmd(args ...string) (string, error)
	Shell() error
	String() string

	GetAttr(string) (string, error)
	SetAttr(string, string) error

	GetConfig() (*config.Config, error)
	WriteConfig(*config.Config) error
}
