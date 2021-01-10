package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// FormatB formats a block hash, and joins it with the letter 'b'. Used when inserting a new block in the db
func FormatB(hash []byte) []byte {
	last := bytes.Join(
		[][]byte{[]byte("b"), hash},
		[]byte{},
	)
	return last
}

func FormatC(txID []byte) []byte {
	last := bytes.Join(
		[][]byte{[]byte("C"), txID},
		[]byte{},
	)
	return last
}

func ReformatKey(key []byte) []byte {
	return key[1:]
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

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	ReverseBytes(result)

	return result
}

// Base58Decode decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}
	return decoded
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}