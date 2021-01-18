package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/chezky/blemflarck/core"
)

//L -> R: Send version message with the local peer's version
//R -> L: Send version message back
//R -> L: Send verack message
//R:      Sets version to the minimum of the 2 versions
//L -> R: Send verack message after receiving version message from R
//L:      Sets version to the minimum of the 2 versions

var (
	blocksNeeded = make(map[int32][]byte)
)

func handleVersion(req []byte, bc *core.Blockchain) {
	var (
		payload Version
	)

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding handleVersion with payload of length %d: %v\n", len(req), err)
		return
	}

	fmt.Printf("version payload is %v\n", payload)

	if !nodeIsKnow(payload.AddrFrom.IP) {
		fmt.Printf("New node found with address: %s\n", payload.AddrFrom)
		// If it is a new node, respond with your own version message before you can confirm it is valid
		sendVersion(payload.AddrFrom, bc)
		knownNodes.Addresses[payload.AddrFrom.IP.String()] = createNewAddress(payload.AddrFrom)
	}

	// If successfully received the Version message, confirm with the sender that it has been received, to update the this receiving node as successful handshake on
	// the sender node.
	sendVerack(payload.AddrFrom)

	myBlockHeight, err := bc.GetChainHeight()
	if err != nil {
		return
	}

	if myBlockHeight > payload.BlockHeight {
		fmt.Printf("my block height is higher haha!\n")
	} else if myBlockHeight < payload.BlockHeight {
		sendGetBlocks(payload.AddrFrom, bc)
	} else {
		fmt.Println("Blockchain is up to date!")
	}

	// Update version to lower of the two nodes
	if payload.Version < nodeVersion {
		fmt.Printf("switching version \"%d\" to match node %s\n", payload.Version, payload.AddrFrom.String())
		nodeVersion = payload.Version
	}
}

// handleVerack is responsible for setting a nodes status to successful handshake if a verack message is received.
func handleVerack(address NetAddress) {
	// make sure it is actually coming from the right place
	if knownNodes.Addresses[address.IP.String()] != nil {
		fmt.Printf("Successfully sent version message, and received verack!\n")
		knownNodes.Addresses[address.IP.String()].Handshake = true
		if err := knownNodes.SaveToFile(); err != nil {
			fmt.Printf("error saving addresses to file: %\n", err)
		}
	}
	fmt.Println("known nodes", knownNodes)
}

// What I want to do, is when I get a "getblocks" command, verify that the
func handleGetBlocks(req []byte, address NetAddress, bc *core.Blockchain) {
	var payload GetBlocks

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding GetBlocks of length %d: %v\n", len(req), err)
		return
	}

	fmt.Printf("getting blocks starting with height %d for address %s\n", payload.Height+1, address.String())

	blk, err := core.ReadBlockFromFile(int(payload.Height))
	if err != nil {
		fmt.Printf("error reading block height %d from file: %v\n", payload.Height, err)
		return
	}

	if bytes.Compare(blk.Hash, payload.Hash) != 0 {
	//	TODO: handle this MUCH better
		fmt.Printf("ERROR: block height \"%d\" on address %s has a different hash than this node does!\n", payload.Height, address.String())
		return
	}

	myHeight, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting chain height for handleGetBlocks: %v\n", err)
		return
	}

	inv := &Inventory{
		Kind:   "blocks",
	}

	for i:=payload.Height; i < myHeight; i++ {
		// over here is i+1 since if i starts at their height we want to start with the next block up
		blk, err := core.ReadBlockFromFile(int(i+1))
		if err != nil {
			fmt.Printf("error reading in block height \"%d\" for handleGetBlocks: %v\n", i, err)
			return
		}
		// same here, we start with their height +1
		inv.Height = append(inv.Height, i+1)
		inv.Items = append(inv.Items, blk.Hash)
	}

	sendInv(address, inv)
}

// handleInventory handles calls to inventory. This can be triggered unsolicited, or in response to getblocks. When handling blocks, store the list of blocks you need to get,
// and then get the blocks from multiple places.
func handleInventory(req []byte, address NetAddress, bc *core.Blockchain) {
	var payload Inventory

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding handleInventory of length %d: %v\n", len(req), err)
		return
	}

	if payload.Kind == "blocks" {
		// populate the map blocksNeeded with all the new blocks to get
		// payload.Height stores the height of the blocks i need. A single height is an int of the block height.
		for idx, height := range payload.Height {
			// payload.Items[idx] is the block hash at that height
			blocksNeeded[height] = payload.Items[idx]
		}

		sendGetData("blocks")
	}
}

func handleGetData(req []byte, address NetAddress) {
	var payload GetData

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error deocding during handleGetData: %v\n", err)
		return
	}

	if payload.Kind == "blocks" {
		fmt.Printf("Address %s requested block \"%d\"\n", address.IP.String(), payload.Height)
		blk, err := core.ReadBlockFromFile(int(payload.Height))
		if err != nil {
			fmt.Printf("error reading in block height \"%d\" for handleGetData: %v\n", payload.Height, err)
			return
		}

		if bytes.Compare(blk.Hash, payload.Hash) != 0 {
			//TODO:// handle this. Maybe have some sort of way to send back errors
			fmt.Printf("ERROR: block #%d and requested block #%d don't have matching hashes\n", blk.Height, payload.Height)
			return
		}

		sendBlock(blk, address)
	}
}

func handleBlock(req []byte, bc *core.Blockchain) {
	// TODO: wow i need tons of block verification work here
	var block core.Block

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&block); err != nil {
		fmt.Printf("error decoding block for handleBlock, with request of length %d: %v", len(req), err)
		return
	}

	if err := bc.UpdateWithNewBlock(block); err != nil {
		fmt.Printf("error updating blockchain with new block #%d: %v\n", block.Height, err)
		return
	}

	fmt.Printf("successfully added block #%d\n", block.Height)
}