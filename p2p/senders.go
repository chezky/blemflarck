package p2p

import (
	"fmt"
	"github.com/chezky/blemflarck/core"
)


func sendVersion(address string, bc *core.Blockchain) {
	height, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting height for send version: %v\n", err)
		return
	}

	version := Version{AddrFrom: getIP(),BlockHeight: height, Version: nodeVersion}

	enc, err := core.GobEncode(version)
	if err != nil {
		return
	}

	cmd := commandToBytes("version")
	payload := append(cmd, enc...)

	if err := SendCmd(address, payload); err != nil {
		fmt.Printf("error sending version cmd: %v\n", err)
		return
	}
}

func sendGetBlocks(address string) {
	getBlocks := Version{AddrFrom: getIP()}

	enc, err := core.GobEncode(getBlocks)
	if err != nil {
		fmt.Printf("error encoding sendGetBlocks: %v\n", err)
		return
	}

	cmd := commandToBytes("getblocks")
	payload := append(cmd, enc...)

	err = SendCmd(address, payload)
	if err != nil {
		fmt.Printf("error sending \"%s\" command: %v\n", "getblocks", err)
		return
	}
}

func sendInv(address string, bc *core.Blockchain, item string) {
	height, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting height for sendInv: %v\n", err)
		return
	}

	tailHash, err := bc.GetTailHash()
	if err != nil {
		fmt.Printf("error getting tail hash for sendInv: %v\n", err)
		return
	}

	data := Inventory{
		AddrFrom:    getIP(),
		Height:      height,
		Hashes: [][]byte{tailHash},
		Item: item,
	}

	enc, err := core.GobEncode(data)
	if err != nil {
		fmt.Printf("error encoding payload for sendInv: %v\n", err)
		return
	}

	cmd := commandToBytes("inv")
	payload := append(cmd, enc...)

	if err := SendCmd(address, payload); err != nil {
		fmt.Printf("error sending getInv command to %s: %v", address, err)
		return
	}
}
