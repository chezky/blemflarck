package core

import (
	"encoding/hex"
	"errors"
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

	utxo := UTXO{Blockchain: &bc}

	if err := utxo.Reindex(); err != nil {
		fmt.Printf("error reindexing UTXO for genesis block")
		return &bc, err
	}

	fmt.Printf("Blockchain successfully created!\n")

	return &bc, err
}

func (bc Blockchain) GetChainHeight() (int, error) {
	var (
		last int
		err error
	)

	if err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		enc := b.Get([]byte("l"))
		if len(enc) == 0 {
			return errors.New("ERROR: no blockchain last found")
		}

		last, err = strconv.Atoi(string(enc))
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		fmt.Printf("error getting chain height: %v\n", err)
		return 0, err
	}
	return last, nil
}

func (bc Blockchain) GetTailHash() ([]byte, error) {
	lastHeight, err := bc.GetChainHeight()
	if err != nil {
		fmt.Printf("error getting chain height for GetTailHash: %v\n", err)
		return nil, err
	}

	blk, err := ReadBlockFromFile(lastHeight)
	if err != nil {
		fmt.Printf("error reading block height %d for GetTailHash: %v\n", lastHeight, err)
		return nil, err
	}

	return blk.Hash, nil
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

// eventually create this where chainstate db stores state of UTXO's
func (bc Blockchain) FindUTXOs() (map[string]*UTXOutputs, error) {
	var (
		// needs to be a slice of int, since one transaction can have multiple used outputs
		// this is a map of transactionID's mapped to the output idx that is referenced by an input
		references = make(map[string][]int)
		UTXOs      = make(map[string]*UTXOutputs)
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
			var outputs UTXOutputs

			id := hex.EncodeToString(tx.ID)
			// next, loop over every output, and check if that output is referenced by an input
		Outputs:
			for outIdx, out := range tx.Vout {
				for _, usedIdx := range references[id] {
					// error where coinbase tx are false positives
					if usedIdx == outIdx {
						continue Outputs
					}
				}
				outputs.Outputs = append(outputs.Outputs, out)
				outputs.Indexes = append(outputs.Indexes, outIdx)
				outputs.BlockHeight = blk.Height
			}

			if len(outputs.Outputs) > 0 {
				UTXOs[id] = &outputs
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