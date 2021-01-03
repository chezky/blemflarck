package core

import (
	"fmt"
	"github.com/boltdb/bolt"
	"strconv"
	"time"
)

// Store the blocks in their own separate .dat file. For example genesis would be 0.dat, block 1 would be 1.dat, etc...
// store as an int // 4 bytes

type Blockchain struct {
	//Tip []byte
	DB *bolt.DB
}

type BCIterator struct {
	LastHash []byte
	DB *bolt.DB
}

func (bc *Blockchain) CreateGenesisBlock() Block {
	genesis := Block{
		Timestamp: time.Now().Unix(),
		Data:      []byte(genesisData),
		Height:    0,
	}

	genesis.Hash, _  = genesis.GenerateHash()

	//bc.Tip = genesis.Hash

	return genesis
}

func CreateBlockchain() (*Blockchain, error) {
	var (
		bc Blockchain
		err error
	)

	bc.DB, err = bolt.Open(dbFile, 0600, nil); if err != nil {
		fmt.Printf("error opening boltDB for file %s: %v\n", dbFile, err)
		return nil, err
	}

	if ChainExists() {
		// just create a Blockchain instance without creating an entirely new chain
		fmt.Printf("Blockchain already exists, using existing chain!\n")
		return &bc, nil
	}

	genesis := bc.CreateGenesisBlock()

	err = genesis.SaveToFile()
	if err != nil {
		fmt.Printf("error creating file for gensis block: %v\n", err)
		return nil, err
	}

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		if err != nil {
			fmt.Printf("error opening bucket %s: %v\n", blocksBucket, err)
			return err
		}
		// b+64-byte block hash : file name of the block
		if err := b.Put(FormatB(genesis.Hash), []byte(strconv.Itoa(genesis.Height))); err != nil {
			fmt.Printf("error inserting genesis block in db: %v\n", err)
			return err
		}
		// l : file name of the block
		if err := b.Put([]byte("l"), []byte(strconv.Itoa(genesis.Height))); err != nil {
			fmt.Printf("error updating l with genesis hash: %v\n", err)
			return err
		}
		return nil
	})

	return &bc, err
}

func (bc Blockchain) NewIterator() (*BCIterator, error) {
	var iter BCIterator

	iter.DB = bc.DB

	err := iter.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		lastIdxByte := b.Get([]byte("l"))
		lastIdx, err :=  strconv.Atoi(string(lastIdxByte))
		if err != nil {
			fmt.Printf("lastIdx was not an integer: %v\n", err)
			return err
		}
		block, err := ReadFromFile(lastIdx)
		if err != nil {
			return err
		}
		iter.LastHash = block.Hash
		return nil
	})

	return &iter, err
}

func (bci *BCIterator) Next() Block {
	var block Block

	if err := bci.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		idxByte := b.Get(FormatB(bci.LastHash))
		idx, err :=  strconv.Atoi(string(idxByte))
		if err != nil {
			fmt.Printf("idxByte was not an integer: %v\n", err)
			return err
		}
		block, err = ReadFromFile(idx)
		bci.LastHash = block.PrevHash
		return err
	}); err != nil {
		fmt.Printf("error getting next block in iter for block hash %s: %v\n", bci.LastHash, err)
	}
	return block
}
