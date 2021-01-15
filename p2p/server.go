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
	nodePort int = 8069
)

var (
	knownNodes = make(map[string]*Address)
	nodeVersion int32 = 1
)

func StartServer() error {
	addr := getIPString()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
		return err
	}

	fmt.Printf("starting server on address: %s\n", addr)
	fmt.Println("-------------")
	fmt.Println()

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

	// hardcoded now for testing locally
	if getIPString() != fmt.Sprintf("%s:%d", "10.0.0.1", nodePort){
		addr := NetAddress{
			IP:   net.IPv4(10,0,0,1),
			Port: nodePort,
		}
		sendVersion(addr, bc)
	}

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

	fullAddr := conn.RemoteAddr().(*net.TCPAddr)
	addr := NetAddress{
		IP:   fullAddr.IP,
	}
	addr.SetPort()

	cmd := bytesToCommand(req[:cmdLength])
	fmt.Printf("recieved \"%s\" command!\n", cmd)

	switch cmd {
	case "version":
		handleVersion(req[cmdLength:], bc)
	case "verack":
		handleVerack(addr)
	case "getblocks":
		handleGetBlocks(req[cmdLength:], addr, bc)
	case "inv":
		handleInventory(req[cmdLength:], addr, bc)
	case "getdata":
		handleGetData(req[cmdLength:], addr, bc)
	default:
		fmt.Printf("ERROR: %s is an unknown command\n", cmd)
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