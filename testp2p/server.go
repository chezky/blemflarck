package testp2p

import (
	"fmt"
	"net"
	"time"
)

func getIPString() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%s:%d", localAddr.IP.String(), 8069)
}

func StartServer() error {
	ln, err := net.Listen("tcp", getIPString())
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {

			time.Sleep(time.Second*5)

			_, err := conn.Write([]byte("dope"))
			if err != nil {
				fmt.Printf("error sending to address")
			}

		}(conn)
	}

}