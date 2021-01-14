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

	if !nodeIsKnow(payload.AddrFrom) {
		fmt.Printf("New node found with address: %s\n", payload.AddrFrom)
		// If it is a new node, respond with your own version message before you can confirm it is valid
		sendVersion(payload.AddrFrom, bc)
		knownNodes[payload.AddrFrom] = createNewAddress(payload.AddrFrom)
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
		sendGetBlocks(payload.AddrFrom)
	} else if myBlockHeight < payload.BlockHeight {
		// TODO: switch this to ask for a different node than the one we just got blocks from
		sendVersion(payload.AddrFrom, bc)
		// handle this
	} else {
		fmt.Println("Blockchain is up to date!")
	}

	// Update version to lower of the two nodes
	if payload.Version > nodeVersion {
		nodeVersion = payload.Version
	}
}

// handleVerack is responsible for setting a nodes status to successful handshake if a verack message is received.
func handleVerack(address string) {
	// make sure it is actually coming from the right place
	if knownNodes[address] != nil {
		fmt.Printf("Successfully sent version message, and received verack!\n")
		knownNodes[address].Handshake = true
	}
	fmt.Println("known nodes", knownNodes)
}

// What I want to do, is when I get a "getblocks" command, verify that the
func handleGetBlocks(req []byte, bc *core.Blockchain) {
	var payload Version

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding GetBlocks of length %d: %v\n", len(req), err)
		return
	}

	sendInv(payload.AddrFrom, bc, "blocks")
}

func handleInventory(req []byte, bc *core.Blockchain) {
	var payload Inventory

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding handleInventory of length %d: %v\n", len(req), err)
		return
	}

	if payload.Item == "blocks" {
		valid, err := bc.CompareBlocks(payload.Height, payload.Hashes[0])
		if err != nil {
			fmt.Printf("error comparing blocks for handleInv: %v\n", err)
			return
		}

		if !valid {
			// TODO: properly handle this issue. Maybe have an error command and handle it
			fmt.Printf("ERROR: block #%d on address %s does not match this nodes block\n", payload.Height, payload.AddrFrom)
			return
		}
	}


}