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
	addr := getIPV6String()
	ln, err := net.Listen("tcp6", addr)
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
		return err
	}

	fmt.Printf("starting server on address: %s\n", addr)
	fmt.Println("-------------")
	fmt.Println()
	fmt.Println(ln.Addr().String())

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

	//hardcoded now for testing locally
	if getIPString() != fmt.Sprintf("%s:%d", "[2a02:ed0:4266:1b00:b987:dc9d:e27c:c71a]", nodePort){
		addr := NetAddress{
			IP:   net.ParseIP("2a02:ed0:4266:1b00:b987:dc9d:e27c:c71a"),
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
	fmt.Printf("recieved \"%s\" command! from %s\n", cmd, fullAddr.IP.String())

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
		handleGetData(req[cmdLength:], addr)
	case "block":
		handleBlock(req[cmdLength:], bc)
	default:
		fmt.Printf("ERROR: %s is an unknown command\n", cmd)
	}
}

func SendCmd(address NetAddress, payload []byte) error {
	fmt.Printf("dialing %s\n", address.String())
	conn, err := net.Dial("tcp6", address.String())
	if err != nil {
		fmt.Printf("error dialing address: %s: %v\n", address, err)
		return err
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(payload))
	return err
}