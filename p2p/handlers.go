package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/chezky/blemflarck/core"
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

	myBlockHeight, err := bc.GetChainHeight()
	if err != nil {
		return
	}

	if myBlockHeight > payload.BlockHeight {
		fmt.Printf("my block height is higher haha!")
	} else if myBlockHeight < payload.BlockHeight {
		// TODO: switch this to ask for a different node than the one we just got blocks from
		sendVersion(payload.AddrFrom, bc)
		// handle this
	}

	if !nodeIsKnow(payload.AddrFrom) {
		fmt.Printf("New node with address: %s\n", payload.AddrFrom)
		knownNode = append(knownNode, payload.AddrFrom)
	}
}
