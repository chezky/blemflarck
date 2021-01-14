package p2p

import (
	"time"
)

type Version struct {
	Version int32 // 4 bytes // version of the node
	Timestamp int64 // 8 bytes // timestamp of when this version message is being sent
	AddrRecv string // eventually make this 26 bytes // address of where this is being sent
	AddrFrom string // address to whom this came from
	BlockHeight int32 // current height of the blockchain on the node
}

type Inventory struct {
	AddrFrom string
	Height int32
	Hashes [][]byte
	Item string // "blocks" for blocks, "txs" for transactions
}


func createVersion(address string, height int32) Version {
	return Version{
		Version:     nodeVersion,
		Timestamp:   time.Now().Unix(),
		AddrRecv:    address,
		AddrFrom:    getIP(),
		BlockHeight: height,
	}
}