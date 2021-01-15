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
		knownNodes[payload.AddrFrom.IP.String()] = createNewAddress(payload.AddrFrom)
	}

	// If successfully received the Version message, confirm with the sender that it has been received, to update the this receiving node as successful handshake on
	// the sender node.
	sendVerack(payload.AddrFrom.String())

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
	if payload.Version > nodeVersion {
		fmt.Printf("switching version \"%d\" to match node %s\n", payload.Version, payload.AddrFrom.String())
		nodeVersion = payload.Version
	}
}

// handleVerack is responsible for setting a nodes status to successful handshake if a verack message is received.
func handleVerack(address NetAddress) {
	// make sure it is actually coming from the right place
	fmt.Println("verack from:", address.String())
	if knownNodes[address.IP.String()] != nil {
		fmt.Printf("Successfully sent version message, and received verack!\n")
		knownNodes[address.IP.String()].Handshake = true
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
		fmt.Printf("ERROR: block height \"%d\" on address %s has a different hash than this node!\n", payload.Height, address.String())
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
		blk, err := core.ReadBlockFromFile(int(i))
		if err != nil {
			fmt.Printf("error reading in block height \"%d\" for handleGetBlocks: %v\n", i, err)
			return
		}

		inv.Height = append(inv.Height, i)
		inv.Items = append(inv.Items, blk.Hash)
	}

	sendInv(address, inv)
}

func handleInventory(req []byte, bc *core.Blockchain) {
	var payload Inventory

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding handleInventory of length %d: %v\n", len(req), err)
		return
	}

	if payload.Kind == "blocks" {
		blocksNeeded := payload.Height
		fmt.Printf("So i need %v\n", blocksNeeded)
	}
}