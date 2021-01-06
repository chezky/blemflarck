package core

import (
	"bytes"
)

// Output is a single output instance. Outputs exist withing transactions. They are where 'coins' are stored, and are locked with a public key hash.
type Output struct {
	Value      int // The amount of 'coins' stored in this output.
	PubKeyHash []byte // The public key hash of the owner of the coins. This hash is a double sha512 hash of the owners public key.
}

// CreateOutput creates an output for an address, with an amount, and then locks the output to that address
func CreateOutput(address string, amount int) Output {
	out := Output{
		Value:      amount,
		PubKeyHash: nil,
	}

	out.Lock([]byte(address))
	return out
}

// Lock is responsible for locking an output to an address. It gets the public key hash by decoding the address, and then removing the
// version and checksum from the hash.
func (out *Output) Lock(address []byte) {
	dec := Base58Decode(address)
	out.PubKeyHash = dec[1:len(dec)-checksumLen]
}

// CanBeUnlocked checks if an address is the one who locked the output.
func (out Output) CanBeUnlocked(address []byte) bool {
	dec := Base58Decode(address)
	pubKeyHash := dec[1:len(dec)-checksumLen]
	return bytes.Compare(pubKeyHash, out.PubKeyHash) == 0
}

