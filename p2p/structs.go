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
	Height []int32 // Height is the height of the block that the item is at
	Items [][]byte // Items are a list of hashes. Transaction or block hashes
	Kind string // Kind tells the node whether this is for transactions or for blocks
}

type GetBlocks struct {
	Height int32 // Height of the latest block you have
	Hash []byte // Hash of the last block you have
}

type GetData struct {
	Height int32
	Hash []byte
	Kind string
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

// String converts a full netAddress to string
func (addr NetAddress) String() string {
	return fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
}

// SetPort sets the port of an address. Default is nodePort. If the address is known tho, make the port the actual port of the address. Usually all ports are the same.
func (addr *NetAddress) SetPort() {
	if !nodeIsKnow(addr.IP) {
		addr.Port = nodePort
		return
	}

	addr.Port = knownNodes[addr.IP.String()].Address.Port
}

// createVersion creates a new Version struct with an address, port, and height
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