package testp2p

import (
	"fmt"
	"io/ioutil"
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

	sendConn, err := net.Dial("tcp", "10.0.0.4:8069")
	if err != nil {
		return err
	}

	sendConn.Write([]byte("sending"))

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		req, err := ioutil.ReadAll(conn)
		fmt.Println(string(req))

		go func(c net.Conn) {

			time.Sleep(time.Second*5)

			_, err := conn.Write([]byte("dope"))
			if err != nil {
				fmt.Printf("error sending to address")
			}

		}(conn)
	}

}