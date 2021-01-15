package p2p

import (
	"fmt"
	"github.com/chezky/blemflarck/core"
)


func sendVersion(address NetAddress, bc *core.Blockchain) {
	if knownNodes[address.IP.String()] == nil {
		knownNodes[address.IP.String()] = createNewAddress(address)
	}

	height, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting height for send version: %v\n", err)
		return
	}

	version := createVersion(address.IP, address.Port, height)

	enc, err := core.GobEncode(version)
	if err != nil {
		return
	}

	cmd := commandToBytes("version")
	payload := append(cmd, enc...)

	if err := SendCmd(address.String(), payload); err != nil {
		fmt.Printf("error sending version cmd: %v\n", err)
		return
	}
	fmt.Printf("sent version to address %s\n", address.String())
}

// sendVerack is sent to acknowledge a Version handshake was received. Once verack is sent back, we can verify that a version is
func sendVerack(address string) {
	cmd := commandToBytes("verack")
	SendCmd(address, cmd)
}

func sendGetBlocks(address NetAddress, bc *core.Blockchain) {
	height, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting chain height for sendGetBlocks: %v\n", err)
		return
	}

	hash, err := bc.GetTailHash()
	if err != nil {
		fmt.Printf("error getting tail hash for sendGetBlocks: %v\n", err)
		return
	}

	getBlocks := GetBlocks{
		Height: height,
		Hash:   hash,
	}

	enc, err := core.GobEncode(getBlocks)
	if err != nil {
		fmt.Printf("error encoding sendGetBlocks: %v\n", err)
		return
	}

	cmd := commandToBytes("getblocks")
	payload := append(cmd, enc...)

	err = SendCmd(address.String(), payload)
	if err != nil {
		fmt.Printf("error sending \"%s\" command: %v\n", "getblocks", err)
		return
	}
}

func sendInv(address NetAddress, inv *Inventory) {
	enc, err := core.GobEncode(&inv)
	if err != nil {
		fmt.Printf("error encoding payload for sendInv: %v\n", err)
		return
	}

	cmd := commandToBytes("inv")
	payload := append(cmd, enc...)

	if err := SendCmd(address.String(), payload); err != nil {
		fmt.Printf("error sending getInv command to %s: %v", address, err)
		return
	}
}

func sendGetData(kind string) {
	cmd := commandToBytes("getdata")
	address := getRandomAddress()

	if kind == "blocks" {
		for height, hash := range blocksNeeded {
			data := GetData{
				Height: height,
				Hash:   hash,
				Kind:	kind,
			}

			enc, err := core.GobEncode(data)
			if err != nil {
				fmt.Printf("error encoding block for sendGetData: %v\n", err)
				return
			}

			payload := append(cmd, enc...)
			SendCmd(address.String(), payload)
		}
	}
}

func sendBlock(block core.Block, address NetAddress) {
	enc, err := block.EncodeBlock()
	if err != nil {
		fmt.Printf("error encoding block for block height %d in sendBlock: %v\n", block.Height, err)
		return
	}

	cmd := commandToBytes("block")
	payload := append(cmd, enc...)

	SendCmd(address.String(), payload)
}