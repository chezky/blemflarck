package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const (
	coinbaseReward = 10
)

// Transactions

// The initial transaction created is called a coinbase transaction. A coinbase transaction has inputs that don't reference any outputs,
// and are locked to the individual being rewarded. For example, if someone mines a block, they get a coinbase transaction. Also in this
// implementation, a random address that is hosting a file, would also be getting a transaction.

type Transaction struct {
	ID   []byte
	Vout []Output
	Vin  []Input
}

// TOOD: update this as transaction gets more complicated
func (tx Transaction) IsCoinbase() bool {
	return tx.Vin[0].OutputIndex == -1
}

func NewCoinbaseTransaction(address []byte) (Transaction, error) {
	var err error

	out := Output{
		Value:      coinbaseReward,
		PubKeyHash: address,
	}

	in := Input{
		OutputIndex: -1,
	}

	tx := Transaction{
		Vout: []Output{out},
		Vin:  []Input{in},
	}

	tx.ID, err = tx.Hash()
	if err != nil {
		fmt.Printf("error creating coinbase tx hash: %v\n", err)
		return tx, err
	}

	return tx, nil
}

func (tx Transaction) Hash() ([]byte, error) {
	enc, err := tx.Serialize()
	if err != nil {
		return nil, err
	}
	hash := sha512.Sum512(enc)
	return hash[:], nil
}

func (tx Transaction) Serialize() ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(tx); err != nil {
		fmt.Printf("error serializing transaction: %v\n", err)
		return nil, err
	}

	return buff.Bytes(), nil
}

func DeserializeTransaction(data []byte) (Transaction, error) {
	var tx Transaction

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&tx); err != nil {
		fmt.Printf("error decoding transaction with data of len %d: %v", len(data), err)
		return tx, err
	}

	return tx, nil
}

// eventually create this where chainstate db stores state of UTXO's
func (bc Blockchain) FindUTXOs() (map[string][]Output, error) {
	var (
		// needs to be a slice of int, since one transaction can have multiple used outputs
		// this is a map of transactionID's mapped to the output idx that is referenced by an input
		references = make(map[string][]int)
		UTXOs      = make(map[string][]Output)
	)

	iter, err := bc.NewIterator()
	if err != nil {
		return UTXOs, err
	}

	for {
		// first get a block
		blk := iter.Next()
		// then begin looping over every transaction in the block
		for _, tx := range blk.Transactions {
			id := hex.EncodeToString(tx.ID)
			// next, loop over every output, and check if that output is referenced by an input
		Outputs:
			for outIdx, out := range tx.Vout {
				for _, usedIdx := range references[id] {
					if usedIdx == outIdx {
						continue Outputs
					}
				}
				UTXOs[id] = append(UTXOs[id], out)
			}

			// for every input, store which output it references
			for _, in := range tx.Vin {
				// coinbase inputs never reference an output
				if !tx.IsCoinbase() {
					referencedID := hex.EncodeToString(in.TransactionID)
					references[referencedID] = append(references[referencedID], in.OutputIndex)
				}
			}
		}
		if len(blk.PrevHash) == 0 {
			break
		}

	}
	return UTXOs, nil
}

func (bc Blockchain) FindSpendableOutputs(address []byte, amount int) (int, map[string][]int, error) {
	outputs := make(map[string][]int)
	accumulated := 0

	UTXOs, err := bc.FindUTXOs()
	if err != nil {
		fmt.Printf("error getting UTXOs for findBalance: %v\n", err)
		return 0, outputs, err
	}

	for txID, outs := range UTXOs {
		for outIdx, out := range outs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				outputs[txID] = append(outputs[txID], outIdx)
			}
			if accumulated > amount {
				break
			}
		}
	}

	if accumulated < amount {
		return accumulated, outputs, errors.New("ERROR: not enough funds")
	}

	return accumulated, outputs, nil
}

func (bc Blockchain) NewTransaction(from, to string, amount int) (Transaction, error) {
	var (
		tx Transaction
	)

	acc, UTXOs, err := bc.FindSpendableOutputs([]byte(from), amount)
	if err != nil {
		return tx, err
	}

	for txID, outs := range UTXOs {
		id, err := hex.DecodeString(txID)
		if err != nil {
			return tx, err
		}
		for _, outIdx := range outs {
			inp := Input{
				TransactionID: id,
				OutputIndex:   outIdx,
				PubKey:        []byte(from),
				Signature:     nil,
			}
			tx.Vin = append(tx.Vin, inp)
		}
	}

	out := Output{
		Value:      amount,
		PubKeyHash: []byte(to),
	}

	tx.Vout = append(tx.Vout, out)

	if acc-amount > 0 {
		remaining := Output{
			Value:      acc - amount,
			PubKeyHash: []byte(from),
		}
		tx.Vout = append(tx.Vout, remaining)
	}

	tx.ID, err = tx.Hash()
	if err != nil {
		fmt.Printf("error hashing tx for newTransaction: %v", err)
		return tx, err
	}

	return tx, nil
}

func (bc Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iter, err := bc.NewIterator()
	if err != nil {
		return Transaction{}, err
	}

	for {
		blk := iter.Next()

		for _, tx := range blk.Transactions {
			if bytes.Compare(ID, tx.ID) == 0 {
				return tx, nil
			}
		}

		if len(blk.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction does not exist")
}

// TrimmedTransaction takes a transaction and removes the pubKey + signature from the inputs. This is in preparation for signing,
// as we don't need to sign the entire tx.
func (tx Transaction) TrimmedTransaction() Transaction {
	var trimmedTX Transaction

	for _, in := range tx.Vin {
		inp := Input{
			TransactionID: in.TransactionID,
			OutputIndex:   in.OutputIndex,
			Signature:     nil,
			PubKey:        nil,
		}
		trimmedTX.Vin = append(trimmedTX.Vin, inp)
	}

	trimmedTX.Vout = tx.Vout
	trimmedTX.ID = tx.ID
	return trimmedTX
}

// Sign is responsible for the logic behind signing a tx. Signing validates that when a transaction is made, the owner of the output is the one
// making the transaction. It does so by creating a trimmed transaction, setting the publicKey of each input to that of the output it is referencing,
// and then hashing that trimmed transaction. Then the private key and trimmed id get signed together to form a two piece signature. Those are appended
// and that is the signature. That signature can now be verified with the public key, and the end result of the same hashedTX process.
func (tx *Transaction) Sign(private ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	var err error

	if tx.IsCoinbase() {
		return nil
	}

	trimmed := tx.TrimmedTransaction()

	for inIdx, in := range trimmed.Vin {
		prevTX := prevTXs[hex.EncodeToString(in.TransactionID)]
		trimmed.Vin[inIdx].PubKey = prevTX.Vout[in.OutputIndex].PubKeyHash
		// remove signature for safety
		trimmed.Vin[inIdx].Signature = nil
		trimmed.ID, err = trimmed.Hash(); if err != nil {
			fmt.Println("error hashing trimmed transaction during signing")
			return err
		}
		// set pubKey back to nil for safety
		trimmed.Vin[inIdx].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &private, trimmed.ID)
		if err != nil {
			fmt.Printf("error siging transaction: %v\n", err)
			return err
		}
		tx.Vin[inIdx].Signature = append(r.Bytes(), s.Bytes()...)
	}
	return nil
}

func (tx Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
	var err error

	trimmed := tx.TrimmedTransaction()
	curve := elliptic.P256()

	for inIdx, in := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(tx.ID)]
		trimmed.Vin[inIdx].PubKey = prevTX.Vout[in.OutputIndex].PubKeyHash
		trimmed.Vin[inIdx].Signature = nil
		trimmed.ID, err = trimmed.Hash()
		if err != nil {
			fmt.Println("error hashing trimmed during verification")
			return false, err
		}
		trimmed.Vin[inIdx].PubKey = nil

		x, y := big.Int{}, big.Int{}
		pubKeyLen := len(tx.Vin[inIdx].PubKey)
		x.SetBytes(in.PubKey[:pubKeyLen/2])
		y.SetBytes(in.PubKey[pubKeyLen/2:])

		r,s := big.Int{}, big.Int{}
		signatureLen := len(in.Signature)
		r.SetBytes(in.Signature[:signatureLen/2])
		s.SetBytes(in.Signature[signatureLen/2:])

		pubKey := ecdsa.PublicKey{Curve: curve, X: &x ,Y: &y}

		if !ecdsa.Verify(&pubKey, trimmed.ID, &r, &s) {
			return false, nil
		}
	}

	return true, nil
}