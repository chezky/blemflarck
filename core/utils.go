package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

// FormatB formats a block hash, and joins it with the letter 'b'. Used when inserting a new block in the db
func FormatB(hash []byte) []byte {
	last := bytes.Join(
		[][]byte{[]byte("b"), hash},
		[]byte{},
	)
	return last
}

// BlockFile takes in a block height and returns a string with that heights filename
func BlockFile(h int) string {
	return fmt.Sprintf("./blocks_gen/%d.dat", h)
}

// ChainExists checks if there is already a chain
func ChainExists() bool {
	files, err := ioutil.ReadDir("./blocks_gen")
	if err != nil {
		os.Mkdir("./blocks_gen", 0777)
	}
	return len(files) > 0
}
