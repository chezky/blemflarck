package p2p

import (

)

type Version struct {
	AddrFrom string
	BlockHeight int
	Version int
}

type Inventory struct {
	AddrFrom string
	Height int
	Hashes [][]byte
	Item string // "blocks" for blocks, "txs" for transactions
}
