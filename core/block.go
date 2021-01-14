// Blocks Bucket

// 'b' + 64-byte block hash : block height of that block hash
// 'l' : block height of latest block

package core

import (
	"bytes"
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"strconv"
	"time"
)

// Goal is to create some sort of Proof of Storage
// Reason being, PoW creates lots of computing power, yet the computing power has no effect on the actual blockchain. Meaning, it does not actually
// allow the bc to store more data, it just is required to figure out their hashes.

// Let's create a block that does both PoS for reward, along with a random address that is storing data. That way we have all the benefits of PoS along
// with an incentive for people to store data. As long as we have a hash record of every file, it's size, and where it is stored, we can pick a random
// file, and then check if that address is hosting over 1 GB. If it is, then it gets a reward alongside the validator. Although this might decrease the
// value of a single coin, it will provide incentive to both be a validator and incentive to store files

const (
	genesisData  = "Welcome to a world created by the people, for the people."
	dbFile       = "blemflarck.db"
	blocksBucket = "blocks"
)

//Block is an instance of a single block.
type Block struct {
	Timestamp    int64  // Timestamp is the time when the block was created.
	Hash         []byte // Hash is sha512 64-byte hash of the block. Its unique identifier.
	PrevHash     []byte // PrevHash is the previous blocks hash.
	Transactions []Transaction
	Height       int    // Height is the index of the block in the blockchain
	Validator    []byte // Validator is the winner of the Proof of Stake lottery
	Winner       []byte // Winner is the winner of the random file lottery
}

// NewBlock takes the previous block, some data, and then creates a new block
func NewBlock(PrevBlock Block, TXs []Transaction) (Block, error) {
	var err error

	block := Block{
		Timestamp:    time.Now().Unix(),
		PrevHash:     PrevBlock.Hash,
		Transactions: TXs,
		Height:       PrevBlock.Height + 1,
	}

	block.Hash, err = block.GenerateHash()

	return block, err
}

// AddBlock adds a block to the blockchain. It first gets the previous block, and then creates a new block. It saves the new block to
// a file, and updates the block bucket, storing the file number under the key 'l'.
func (bc *Blockchain) AddBlock(TXs []Transaction) error {
	var (
		prevBlock Block
		err       error
	)

	utxo := UTXO{Blockchain: bc}

	for _, tx := range TXs {
		if !tx.IsCoinbase() {
			verified, err := utxo.VerifyTransaction(tx)
			if err != nil {
				return err
			}
			if !verified {
				return errors.New("ERROR: TX is invalid")
			}
		}
	}

	if err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastIdxByte := b.Get([]byte("l"))
		lastIdx, err := strconv.Atoi(string(lastIdxByte))
		if err != nil {
			return err
		}
		prevBlock, err = ReadBlockFromFile(lastIdx)
		return err
	}); err != nil {
		fmt.Printf("error getting prev block for AddBLock: %v\n", err)
		return err
	}

	block, err := NewBlock(prevBlock, TXs)
	if err != nil {
		fmt.Printf("error creating new block with prev bloch hash %s: %v\n", hex.EncodeToString(prevBlock.Hash), err)
		return err
	}

	err = block.SaveToFile()
	if err != nil {
		fmt.Printf("error creating file for block #%d: %v\n", block.Height, err)
		return err
	}

	if err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if err := b.Put(FormatB(block.Hash), []byte(strconv.Itoa(block.Height))); err != nil {
			return err
		}
		if err := b.Put([]byte("l"), []byte(strconv.Itoa(block.Height))); err != nil {
			return err
		}
		return nil
	}); err != nil {
		fmt.Printf("error updating db with new block: %v\n", err)
		return err
	}

	if err := utxo.Update(block); err != nil {
		fmt.Printf("error updating UTXO during new block: %v\n", err)
		return err
	}

	return nil
}

// EncodeBlock encodes a block to a byte slice, this allows the block to be saved to file.
func (b Block) EncodeBlock() ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(b); err != nil {
		fmt.Printf("error encoding Block: %v\n", err)
		return nil, err
	}

	return buff.Bytes(), nil
}

// DecodeBlock takes in an encoded block, decodes it, and then returns the block.
func DecodeBlock(data []byte) (Block, error) {
	var block Block
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&block)
	if err != nil {
		fmt.Printf("error decoding block, data is of length %d: %v\n", len(data), err)
	}
	return block, err

}

// GenerateHash generates a new sha512 hash for a block
// Eventually replace this when implementing proof
func (b Block) GenerateHash() ([]byte, error) {
	enc, err := b.EncodeBlock()
	if err != nil {
		return nil, err
	}
	hash := sha512.Sum512(enc)
	return hash[:], nil
}

// SaveToFile saves a block to file. If the block height is 518 then the file will be 518.dat
func (b Block) SaveToFile() error {
	encoded, err := b.EncodeBlock()
	if err != nil {
		fmt.Printf("error trying to save block: %v\n", err)
		return err
	}

	err = ioutil.WriteFile(BlockFile(b.Height), encoded, 0666)
	return err
}

// ReadBlocksFromFile reads a block in from a file, with the file name being its height followed by .dat. For example, 400.dat
//TODO: perhaps we should make height an in32 instead of int
func ReadBlockFromFile(height int) (Block, error) {
	var block Block

	encBlock, err := ioutil.ReadFile(BlockFile(height))
	if err != nil {
		fmt.Printf("error reading file for block height %d: %v\n", height, err)
		return block, err
	}

	block, err = DecodeBlock(encBlock)
	return block, err
}

// Proof of Stake

// When a validator node wants to connect to the network, it broadcasts to other nodes its intention to connect, and if he is verified to have more than
// x amount of coins, he can join the network as a validator. Store in a bucket a list of blocked addresses, and have it be shared across all nodes.
// Every 30 minutes send a heartbeat, and if it doesn't get accepted, remove the validator from the list. Figure out how to make sure each validator node
// isn't completely bombarded by heartbeat signals every 30 minutes. Figure out some way to 'stake' validator node's coins.

// ‘Randomized Block Selection’ and ‘Coin Age Selection’.

// Things to remember
// In order to ensure that the Validator node doesn't just put down himself for the storage winner too, A, ensure that they must be different,
// and B, when the validator node is chosen, the storage winner should also be chosen.

// Questions?

// How do I ensure that each node doesn't chose a different validator node, and there is no data race for picking a validator node
// What exactly is stopping someone from making a node with a completely fake blockchain
