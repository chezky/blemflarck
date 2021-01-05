package core

import (
	"bytes"
)

type Output struct {
	Value      int
	PubKeyHash []byte
}

func (out Output) CanBeUnlocked(address []byte) bool {
	return bytes.Compare(address, out.PubKeyHash) == 0
}

func CreateOutput(address []byte, amount int) Output {
	out := Output{
		Value:      amount,
		PubKeyHash: address,
	}

	return out
}
