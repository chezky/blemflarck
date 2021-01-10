package core

import (
	"fmt"
	"github.com/boltdb/bolt"
	"strconv"
	"time"
)

// Store the blocks in their own separate .dat file. For example genesis would be 0.dat, block 1 would be 1.dat, etc...
// store as an int // 4 bytes

// Blockchain is a single instance of the blockchain.
type Blockchain struct {
	//Tip []byte
	DB *bolt.DB // DB is a pointer to an open boltDB connection
}

// BCIterator is an instance of a blockchain iterator
type BCIterator struct {
	LastHash []byte   // LastHash is the hash of the block the iterator will iterate over next.
	DB       *bolt.DB // DB is a pointer to an open boltDB connection
}

// CreateGenesisBlock creates the first (genesis) block of a chain.
func (bc *Blockchain) CreateGenesisBlock(address string) Block {
	cbTX, err := NewCoinbaseTransaction(address)
	if err != nil {
		fmt.Printf("error creating cbTX in genesis: %v\n", err)
	}

	genesis := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{cbTX},
		Height:       0,
	}

	genesis.Hash, _ = genesis.GenerateHash()

	//bc.Tip = genesis.Hash

	return genesis
}

// CreateBlockchain is responsible for either creating and returning, or just returning a blockchain instance. If there are no blocks, then
// create a new genesis and blockchain. Otherwise just return a blockchain instance.
func CreateBlockchain(address string) (*Blockchain, error) {
	var (
		bc  Blockchain
		err error
	)

	// open a db connection
	bc.DB, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Printf("error opening boltDB for file %s: %v\n", dbFile, err)
		return nil, err
	}

	// check if there already is a saved chain
	if ChainExists() {
		// just create a Blockchain instance without creating an entirely new chain
		return &bc, nil
	}

	// create a genesis block
	genesis := bc.CreateGenesisBlock(address)

	// save the genesis block
	err = genesis.SaveToFile()
	if err != nil {
		fmt.Printf("error creating file for gensis block: %v\n", err)
		return nil, err
	}

	// update the db with the genesis block
	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		if err != nil {
			fmt.Printf("error opening bucket %s: %v\n", blocksBucket, err)
			return err
		}
		// b+64-byte block hash : height of the block
		if err := b.Put(FormatB(genesis.Hash), []byte(strconv.Itoa(genesis.Height))); err != nil {
			fmt.Printf("error inserting genesis block in db: %v\n", err)
			return err
		}
		// l : height of the block
		if err := b.Put([]byte("l"), []byte(strconv.Itoa(genesis.Height))); err != nil {
			fmt.Printf("error updating l with genesis hash: %v\n", err)
			return err
		}
		return nil
	})

	fmt.Printf("Blockchain successfully created!\n")

	return &bc, err
}

// NewIterator creates a new blockchain iterator
func (bc Blockchain) NewIterator() (*BCIterator, error) {
	var iter BCIterator

	iter.DB = bc.DB

	err := iter.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		// get the height of the latest block
		lastIdxByte := b.Get([]byte("l"))
		// convert the height in bytes to an int
		lastIdx, err := strconv.Atoi(string(lastIdxByte))
		if err != nil {
			fmt.Printf("lastIdx was not an integer: %v\n", err)
			return err
		}
		// read the block in from its file
		block, err := ReadBlockFromFile(lastIdx)
		if err != nil {
			return err
		}
		iter.LastHash = block.Hash
		return nil
	})

	return &iter, err
}

// Next iterates over a blockchain and gets each block in the chain. It starts with the top and goes top -> down
func (bci *BCIterator) Next() Block {
	var block Block

	if err := bci.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		idxByte := b.Get(FormatB(bci.LastHash))
		idx, err := strconv.Atoi(string(idxByte))
		if err != nil {
			fmt.Printf("idxByte was not an integer: %v\n", err)
			return err
		}
		block, err = ReadBlockFromFile(idx)
		bci.LastHash = block.PrevHash
		return err
	}); err != nil {
		fmt.Printf("error getting next block in iter for block hash %s: %v\n", bci.LastHash, err)
	}
	return block
}

//func (bc Blockchain) FindReferencedOutputs(tx Transaction) (map[string]Transaction, error) {
//	referenced := make(map[string]Transaction)
//
//	UTXOs, err := bc.FindUTXOs()
//	if err != nil {
//		fmt.Printf("error finding UTXOs for FindReferencedOutputs: %v\n", err)
//		return referenced, err
//	}
//
//	// run through each output, and check if the ID from any of the txInputs match the outputs txID.
//	for txID, _ := range UTXOs {
//		for _, in := range tx.Vin {
//			if txID == hex.EncodeToString(in.TransactionID) {
//				referencedTX, err := bc.FindTransaction(in.TransactionID)
//				if err != nil {
//					fmt.Printf("error finding referencedTX: %v", err)
//					return referenced, err
//				}
//				referenced[txID] = referencedTX
//			}
//		}
//	}
//
//	return  referenced, nil
//}