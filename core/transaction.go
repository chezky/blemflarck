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
	"time"
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
	// implemented since two cb tx's were ending up with duplicate hashes
	Timestamp int64
}

// TOOD: update this as transaction gets more complicated
func (tx Transaction) IsCoinbase() bool {
	return tx.Vin[0].OutputIndex == -1
}

func NewCoinbaseTransaction(address string) (Transaction, error) {
	var err error

	out := CreateOutput(address, coinbaseReward)

	in := Input{
		OutputIndex: -1,
	}

	tx := Transaction{
		Vout: []Output{out},
		Vin:  []Input{in},
	}

	tx.Timestamp = time.Now().Unix()
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

func (bc *Blockchain) NewTransaction(from, to string, amount int) (Transaction, error) {
	var (
		tx Transaction
	)

	wallets, err := ReadWalletsFromFile()
	if err != nil {
		fmt.Printf("error reading wallets from file for new TX: %v\n", err)
		return tx, err
	}

	wallet := wallets.Wallets[from]
	if wallet.PublicKey == nil {
		return tx, errors.New("ERROR: this address was not found")
	}

	utxo := UTXO{ Blockchain: bc}

	acc, UTXOs, err := utxo.FindSpendableOutputs([]byte(from), amount)
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
				PubKey:        wallet.PublicKey,
				Signature:     nil,
			}
			tx.Vin = append(tx.Vin, inp)
		}
	}

	out := CreateOutput(to, amount)

	tx.Vout = append(tx.Vout, out)

	if acc-amount > 0 {
		remainingOut := CreateOutput(from, acc-amount)
		tx.Vout = append(tx.Vout, remainingOut)
	}

	tx.Timestamp = time.Now().Unix()
	tx.ID, err = tx.Hash()
	if err != nil {
		fmt.Printf("error hashing tx for newTransaction: %v", err)
		return tx, err
	}

	if err := utxo.SignTransaction(tx, wallet.PrivateKey); err != nil {
		fmt.Printf("error signing tx: %v\n", err)
		return tx, err
	}

	return tx, nil
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

// When one makes a transaction, in reality he should be providing his public+private keys, and the address of the recipient. In this current
// implementation, since the wallets are stored locally, we have the sender input his address. We then lookup a senders private+public keys
// in relation to that address.

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
		prevTX := prevTXs[hex.EncodeToString(in.TransactionID)]
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

func (u UTXO) SignTransaction(tx Transaction, private ecdsa.PrivateKey) error {
	prevTXs, err := u.FindReferencedOutputs(tx)
	if err != nil {
		fmt.Printf("error finding refrenced outputs for Signing: %v\n", err)
		return err
	}

	if err := tx.Sign(private, prevTXs); err != nil {
		fmt.Printf("error signing tx: %v\n", err)
		return err
	}
	return nil
}

func (u UTXO) VerifyTransaction(tx Transaction) (bool, error) {
	prevTXs, err := u.FindReferencedOutputs(tx)
	if err != nil {
		fmt.Printf("error finding referenced outputs during verification: %v\n", err)
		return false, err
	}

	verified, err := tx.Verify(prevTXs)
	if err != nil {
		fmt.Printf("error verifiying transaction: %v\n", err)
		return false, err
	}

	return verified, err
}