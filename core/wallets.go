package core

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	walletFile = "wallets.dat"
)

type Wallets struct {
	Wallets map[string]Wallet
}

func (ws Wallets) SaveToFile() error {
	enc, err := ws.EncodeWallets()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(walletFile, enc, 0666); err != nil {
		fmt.Printf("error writing wallets to file: %v\n", err)
	}
	return nil
}

func ReadWalletsFromFile() (Wallets, error) {
	var wallets Wallets

	// without this line, if wallets.dat doesn't exist, errors will be thrown
	wallets.Wallets = make(map[string]Wallet)

	encWallets, err :=  ioutil.ReadFile(walletFile)
	if err != nil {
		if strings.Contains(err.Error(), " file") {
			_ = ioutil.WriteFile(walletFile, nil, 0666)
			return wallets, nil
		}
		fmt.Printf("error reading in wallets from file %s: %v\n", walletFile, err)
		return wallets, err
	}

	if len(encWallets) > 0 {
		wallets, err = DecodeWallets(encWallets)
		if err != nil {
			fmt.Printf("error decoding wallets, during ReadWalletsFromFile\n")
		}
	}

	return wallets, err
}

func (ws Wallets) EncodeWallets() ([]byte, error) {
	var buff bytes.Buffer

	gob.Register(elliptic.P256())

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(ws); if err != nil {
		fmt.Printf("error encoding a wallet: %v\n", err)
	}
	return buff.Bytes(), err
}

func DecodeWallets(data []byte) (Wallets, error) {
	var wallets Wallets

	gob.Register(elliptic.P256())

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&wallets); if err != nil {
		fmt.Printf("error decoding wallets with data of lenght %d: %v\n", len(data), err)
	}

	return wallets, err
}
