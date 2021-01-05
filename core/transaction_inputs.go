package core

import (

)

type Input struct {
	TransactionID []byte
	OutputIndex   int
	Signature     []byte
	PubKey        []byte
}

