package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func handleVersion(req []byte) {
	var (
		payload Version
	)

	dec := gob.NewDecoder(bytes.NewReader(req))
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding handleVersion with payload of length %d: %v\n", len(req), err)
		return
	}

	fmt.Printf("version payload is %v", payload)
}
