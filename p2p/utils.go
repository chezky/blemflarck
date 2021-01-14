package p2p

import (
	"fmt"
	"net"
)

func getIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%s:%d", localAddr.IP.String(), nodePort)
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

func nodeIsKnow(n string) bool {
	for addr, _ := range knownNodes {
		if addr == n {
			return true
		}
	}
	return false
}