//Ideally this should not be harcoded, and instead have seeds, but here we are

package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

const (
	addressesFile = "addresses.dat"
)

// A slice of addresses, used for reading in addresses from file
type Addresses struct {
	Addresses map[string]*Address
}

func createNewAddress(addr NetAddress) *Address {
	return &Address{
		Address: addr,
		Handshake: false,
		Timestamp: time.Now().Unix(),
	}
}

// String converts a full netAddress to string
func (addr NetAddress) String() string {
	if addr.IP.To16() != nil {
		return fmt.Sprintf("[%s]:%d", addr.IP.String(), addr.Port)
	}
	return fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
}

// SetPort sets the port of an address. Default is nodePort. If the address is known tho, make the port the actual port of the address. Usually all ports are the same.
func (addr *NetAddress) SetPort() {
	if !nodeIsKnow(addr.IP) {
		addr.Port = nodePort
		return
	}

	addr.Port = knownNodes.Addresses[addr.IP.String()].Address.Port
}

func (addrs Addresses) SaveToFile() error {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(addrs); err != nil {
		fmt.Printf("error encoding addresses to save to file: %v\n", err)
		return err
	}

	// TODO: like cmon what eeven is this
	if err := ioutil.WriteFile("./"+addressesFile, []byte{}, 0666); err != nil {
		return err
	}

	if err := ioutil.WriteFile("./"+addressesFile, buff.Bytes(), 0666); err != nil {
		fmt.Printf("error writing adddresses to file: %v\n", err)
		return err
	}

	return nil
}

func ReadAddressesFromFile() (Addresses, error) {
	var addr Addresses

	encAddresses, err :=ioutil.ReadFile("./" + addressesFile)
	if err != nil {
		if strings.Contains(err.Error(), "file") {
			ioutil.WriteFile("./" + addressesFile, []byte{}, 0666)
			return ReadAddressesFromFile()
		}
		fmt.Printf("error reading addresses in from file of length %d: %v\n", len(encAddresses), err)
		return addr, err
	}

	enc := gob.NewDecoder(bytes.NewReader(encAddresses))
	if err := enc.Decode(&addr); err != nil {
		if strings.Contains(err.Error(), "EOF") {
			addr.Addresses = make(map[string]*Address)
			return addr, nil
		}
		fmt.Printf("error decoding addresses of length %d: %v\n", len(encAddresses), err)
		return addr, err
	}

	return addr, nil
}