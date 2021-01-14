package p2p

import (
	"bytes"
	"fmt"
	"github.com/chezky/blemflarck/core"
	"io"
	"io/ioutil"
	"net"
)

const (
	cmdLength = 12
	nodeVersion = 1
	// Eventually remove this, as all will be port https/http
	nodePort = 8080
)

var (
	knownNode []string
)

type Version struct {
	AddrFrom string
	BlockHeight int
	Version int
}

func StartServer() error {
	addr := getIP()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
		return err
	}

	fmt.Printf("starting server on address: %s\n", addr)

	defer ln.Close()

	if !core.ChainExists() {
		fmt.Printf("Unable to find a chain. Either create one, or redownload the genesis block.")
		return err
	}

	bc, err := core.CreateBlockchain("")
	if err != nil {
		return err
	}

	defer bc.DB.Close()

	sendVersion("10.0.0.1:8080", bc)

	for {
		conn, err := ln.Accept(); if err != nil {
			fmt.Printf("error accepting connection: %v\n", err)
			// TOOD: perhaps remove this exit call
			return err
		}
		go HandleConnection(conn, bc)
	}
}

func HandleConnection(conn net.Conn, bc *core.Blockchain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Printf("error handling connection: %v\n", err)
	}

	cmd := bytesToCommand(req[:cmdLength])
	fmt.Printf("recieved \"%s\" command!\n", cmd)

	switch cmd {
	case "version":
		handleVersion(req[cmdLength:], bc)
	default:
		fmt.Printf("ERROR: %s is an unknown command\n", cmd)
	}
}

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

type Blocks struct {
	AddrFrom string
	Height int
	BlockHashes [][]byte
}

func sendGetBlocks(address string, bc *core.Blockchain) {
	height, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting chain height for sendGetBlocks: %v\n", err)
		return
	}

	tailHash, err := bc.GetTailHash()
	if err != nil {
		fmt.Printf("error getting tail hash for sendGetBlocks: %v\n", err)
		return
	}

	blocks := Blocks{AddrFrom: getIP(), Height: height}
	blocks.BlockHashes = append(blocks.BlockHashes, tailHash)

	enc, err := core.GobEncode(blocks)
	if err != nil {
		fmt.Printf("error encoding sendGetBlocks: %v\n", err)
		return
	}

	cmd := commandToBytes("getblocks")
	payload := append(cmd, enc...)

	err = SendCmd(address, payload)
	if err != nil {
		fmt.Printf("error sending \"%s\" command\n", err)
		return
	}
}

func SendCmd(address string, payload []byte) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("error dialing address: %s: %v\n", address, err)
		return err
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(payload))
	return err
}