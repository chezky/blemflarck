package p2p

import (
	"fmt"
	"net"
	"time"
)

type NetAddress struct {
	IP net.IP
	Port int
}

type Version struct {
	Version int32 // 4 bytes // version of the node
	Timestamp int64 // 8 bytes // timestamp of when this version message is being sent
	AddrRecv NetAddress // eventually make this 26 bytes // address of where this is being sent
	AddrFrom NetAddress // address to whom this came from
	BlockHeight int32 // current height of the blockchain on the node
}

type Inventory struct {
	AddrFrom NetAddress
	Height int32
	Hashes [][]byte
	Item string // "blocks" for blocks, "txs" for transactions
}

type GetBlocks struct {
	Height int32 // Height of the latest block you have
	Hash []byte // Hash of the last block you have
}

type Address struct {
	Address NetAddress
	Handshake bool
	Timestamp int64
}

func createNewAddress(addr NetAddress) *Address {
	return &Address{
		Address: addr,
		Handshake: false,
		Timestamp: time.Now().Unix(),
	}
}

func (addr NetAddress) String() string {
	return fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
}

func createVersion(addr net.IP, port int, height int32) Version {
	return Version{
		Version:     nodeVersion,
		Timestamp:   time.Now().Unix(),
		AddrRecv:    NetAddress{
			IP:   addr,
			Port: port,
		},
		AddrFrom: NetAddress{
			IP:   nodeIP(),
			Port: nodePort,
		},
		BlockHeight: height,
	}
}