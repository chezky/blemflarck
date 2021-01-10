package core

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"strings"
)

const (
	UTXOBucket = "chainstate"
)

// UTXO stands for unspent transaction outputs.
type UTXO struct {
	Blockchain *Blockchain
}

// We want to store a list of all UTXO's in a chainstate bucket, that way we can easily look up if an output was referenced
// bc.FindUTXOs
// FindSpendableOutputs

// FindUTXOs can be used for reindexing, and then findSpendableOutputs can then be used to find throughout the chainstate bucket
// Also make an update function that removes or adds UTXOs every time a block is creates
// run through the inputs of each tx, and find which output they are referencing. Then remove that output
// Append every output on the block to chainstate

func (u UTXO) Reindex() error {
	if err := u.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(UTXOBucket)); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				fmt.Printf("error deleting UTXO bucket: %v\n", err)
				return err
			}
		}

		_, err := tx.CreateBucket([]byte(UTXOBucket))
		if err != nil {
			fmt.Printf("error creating UTXO bucket: %v\n", err)
		}

		return nil
	}); err != nil {
		fmt.Printf("error opening db connection for Reindex: %v\n", err)
		return err
	}

	UTXOs, err := u.Blockchain.FindUTXOs()
	if err != nil {
		fmt.Printf("error finding UTXOs for Reindex: %v\n", err)
		return err
	}

	if err := u.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))
		for txID, outputs := range UTXOs {
			serialized, err := outputs.SerializeOutputs(); if err != nil {
				return err
			}

			byteID, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			if err := b.Put(FormatC(byteID), serialized); err != nil {
				fmt.Printf("error putting in outputs during reindex: %v\n", err)
				return err
			}
		}

		return nil
	}); err != nil {
		fmt.Printf("error updating db during reindex: %v\n", err)
		return err
	}
	return nil
}

func (u UTXO) FindUTXOs() (map[string]*UTXOutputs, error) {
	UTXOs := make(map[string]*UTXOutputs)

	err := u.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			output, err := DecodeOutputs(v)
			if err != nil{
				return err
			}

			txID:= hex.EncodeToString(ReformatKey(k))

			UTXOs[txID] = &output
		}
		return nil
	}); if  err != nil {
		fmt.Printf("error finding UTXOs: %v\n", err)
	}
	return UTXOs, err
}

func (u UTXO) FindSpendableOutputs(address []byte, amount int) (int, map[string][]int, error) {
	outputs := make(map[string][]int)
	accumulated := 0

	UTXOs, err := u.FindUTXOs()
	if err != nil {
		fmt.Printf("error getting UTXOs for findBalance: %v\n", err)
		return 0, outputs, err
	}

	for txID, outs := range UTXOs {
		for outIdx, out := range outs.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				outputs[txID] = append(outputs[txID], outIdx)
			} else if accumulated >= amount {
				break
			}
		}
	}

	if accumulated < amount {
		return accumulated, outputs, errors.New("ERROR: not enough funds")
	}

	return accumulated, outputs, nil
}

func (u UTXO) FindReferencedOutputs(tx Transaction) (map[string]Transaction, error) {
	referenced := make(map[string]Transaction)

	// get all the outputs that are unspent
	UTXOs, err := u.FindUTXOs()
	if err != nil {
		fmt.Printf("error finding UTXOs for FindReferencedOutputs: %v\n", err)
		return referenced, err
	}

	// run through each output, and check if the ID from any of the txInputs match the outputs txID.
	// for every transaction that has unspent outputs:
	for txID, utxo := range UTXOs {
		// for every input in the new transaction
		for _, in := range tx.Vin {
			// is the transaction with open outputs the one that the new transaction is referencing?
			if txID == hex.EncodeToString(in.TransactionID) {
				// find that transaction
				referencedTX, err := u.FindTransaction(in.TransactionID, utxo.BlockHeight)
				if err != nil {
					fmt.Printf("error finding referencedTX: %v", err)
					return referenced, err
				}
				// add that transaction to the list of transaction this new transaction references
				referenced[txID] = referencedTX
			}
		}
	}
	// return the referenced transactions list
	return  referenced, nil
}

func (u UTXO) FindTransaction(txID []byte, blockHeight int) (Transaction, error) {
	var tx Transaction

	block, err := ReadBlockFromFile(blockHeight)
	if err != nil {
		fmt.Println("error reading block from file for findTransaction")
		return tx, err
	}

	for _, tx := range block.Transactions {
		if bytes.Compare(txID, tx.ID) == 0 {
			return tx, nil
		}
	}
	return tx, errors.New("ERROR: cannot find transaction in that block")
}

func (u UTXO) Update(block Block) error {
	if err := u.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))

		// for every transaction in this new block
		for _, tx := range block.Transactions {
			// delete all UTXOs that are now referenced by inputs
			// skip coinbase since it has no valid inputs
			if !tx.IsCoinbase() {
				// first we check for referenced outputs by seeing what each input references
				for _, in := range tx.Vin {
					// get the current UTXOs of a referenced transaction
					encOuts := b.Get(FormatC(in.TransactionID))
					// if for some reason there is no actual transaction, freak out
					if len(encOuts) == 0 {
						return errors.New("ERROR: for some reason referenced output couldn't be found in chainstate")
					}

					outs, err := DecodeOutputs(encOuts)
					if err != nil {
						return err
					}

					// create a new UTXOutputs to store the newly updated outputs
					updatedOuts := UTXOutputs{BlockHeight: outs.BlockHeight}

					// for every output in the range of unspent outputs
					for outIdx, out := range outs.Outputs {
						// If the index of where on the transaction this output lives, is equal to the index that the input references, then it's
						// that output that the input is referencing. Since we want every output that isn't referenced we find those that are not
						// equal.
						if outs.Indexes[outIdx] != in.OutputIndex {
							// if it isn't the one being referenced, then keep it
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
							updatedOuts.Indexes = append(updatedOuts.Indexes, outs.Indexes[outIdx])
						}
					}

					// if there are no more outputs, delete the entire transaction
					if len(updatedOuts.Outputs) == 0 {
						if err := b.Delete(FormatC(in.TransactionID)); err != nil {
							fmt.Printf("error deleting tx with UTXOs from chainstate: %v\n", err)
							return err
						}
					// otherwise update the transaction with the new amount of UTXOs
					} else {
						enc, err := updatedOuts.SerializeOutputs(); if err != nil {
							return err
						}
						if err := b.Put(FormatC(in.TransactionID), enc); err != nil {
							return err
						}
					}
				}
			}

			// add all new outputs to chainstate
			var outputs = UTXOutputs{BlockHeight: block.Height, Outputs: tx.Vout}
			// create a slice of int ranging from 0 to amountOfOutputs-1, since every output is a free output
			for outIdx, _ := range tx.Vout {
				outputs.Indexes = append(outputs.Indexes, outIdx)
			}

			enc, err := outputs.SerializeOutputs()
			if err != nil {
				return err
			}

			if err := b.Put(FormatC(tx.ID), enc); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("error during UTXO Update: %v\n", err)
		return err
	}

	return nil
}