package p2p

import (
	"bytes"
	"fmt"
	"github.com/chezky/blemflarck/core"
	"io"
	"io/ioutil"
	"net"
	"time"
)

func StartServer() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", getIP(), 8080))
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
		return err
	}

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

	fmt.Printf("received command from %s\n", bytesToCommand(req))

	time.Sleep(4 * time.Second)

	err = SendCmd(bytesToCommand(req))
	if err != nil {
		fmt.Println("error sending command", err)
	}
}

func SendCmd(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("error dialing address: %s: %v", address, err)
		return err
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader([]byte(getIP())))
	return err
}

func getIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
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