package core

import (

)

// Input is a single Transaction input
// Inputs always reference outputs, unless they are part of a coinbase transaction.
type Input struct {
	TransactionID []byte // TransactionID is the ID of the transaction that houses the output that this input references.
	OutputIndex   int // OutputIndex is the index of the output on the transaction.
	Signature     []byte // Signature stores the signature of the transaction after it gets signed. This signature can then be verified.
	PubKey        []byte // PubKey is the full public key of the one who created this input by creating a transaction. I.e: the sender.
}

