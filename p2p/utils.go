package p2p

import (
	"fmt"
	"net"
	"time"
)

func getIPString() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%s:%d", localAddr.IP.String(), nodePort)
}

func getIPV6String() string {
	conn, _ := net.Dial("udp6", "[2a00:1450:4001:817::200e]:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("[%s]:%d", localAddr.IP, nodePort)
}

func nodeIP() net.IP {
	conn, err := net.Dial("udp", "[2a00:1450:4001:817::200e]:80")
	if err != nil {
		fmt.Printf("error finding nodeIP: %v\n", err)
		return net.IP{}
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
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

func nodeIsKnow(n net.IP) bool {
	for addr, _ := range knownNodes {
		if addr == n.String() {
			return true
		}
	}
	return false
}

func getRandomAddress() NetAddress {
	var node *Address
	for _, node = range knownNodes {
		// if it's a valid node, and node has responded within the last 30m
		if node.Handshake && (node.Timestamp + 1800) > time.Now().Unix() {
			return node.Address
		}
	}
	fmt.Printf("ERROR: can't find a node that is accepted and has a heartbeat\n")
	return node.Address
}