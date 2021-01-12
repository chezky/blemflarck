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
)

type Message struct {
	AddrFrom string
	Payload []byte
}

func StartServer() error {
	addr := fmt.Sprintf("%s:%d", getIP(), 8080)
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

	cmd := req[:cmdLength]
	fmt.Printf("recieved %s command!", cmd)

	switch cmd {
	default:
		fmt.Printf("ERROR: %s is an unknown command", cmd)
	}
}

func SendCmd(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("error dialing address: %s: %v\n", address, err)
		return err
	}

	defer conn.Close()

	cmd := commandToBytes("version")

	_, err = io.Copy(conn, bytes.NewReader(cmd))
	return err
}

func getIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func commandToBytes(cmd string) []byte {
	var b [cmdLength]byte

	for i, c := range cmd {
		b[i] = byte(c)
	}
	return b[:]
}

func bytesToCommand(data []byte) string {
	var cmd []byte

	for _, b := range data {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}
	return fmt.Sprintf("%s", cmd)
}