package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func handleVersion(req []byte) {
	var (
		buff bytes.Buffer
		payload Version
	)

	dec := gob.NewDecoder(&buff)
	if err := dec.Decode(&payload); err != nil {
		fmt.Printf("error decoding handleVersion with payload of length %d: %v", len(req), err)
		return
	}

	fmt.Printf("version payload is %v", payload)
}
