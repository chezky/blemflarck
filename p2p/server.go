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
			fmt.Printf("error accepting connection")
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
		handleVersion(req[cmdLength:])
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

	if err := SendCmd(address, enc); err != nil {
		fmt.Printf("error sending version cmd: %v", err)
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

	cmd := commandToBytes("version")
	data := bytes.NewReader(append(cmd, payload...))

	_, err = io.Copy(conn, data)
	return err
}