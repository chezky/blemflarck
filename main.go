package main

import (
	"encoding/hex"
	"fmt"
	"github.com/chezky/blemflarck/core"
	"log"
)

func main() {
	//os.RemoveAll("./blocks_gen/")
	//os.Mkdir("blocks_gen", 0777)

	bc, err := core.CreateBlockchain("chezky")
	if err != nil {
		log.Fatal(err)
	}

	tx, err := bc.NewTransaction("chezky", "shmuel", 10)
	if err != nil {
		log.Fatal(err)
	}

	bcTX, err := core.NewCoinbaseTransaction([]byte("chezky"))
	if err != nil {
		log.Fatal(err)
	}

	if err := bc.AddBlock([]core.Transaction{tx, bcTX}); err != nil {
		log.Fatal(err)
	}

	iter, err := bc.NewIterator()
	if err != nil {
		log.Fatal(err)
	}

	for {
		blk := iter.Next()

		fmt.Printf("############ Block height: %d ############\n", blk.Height)
		fmt.Printf("Hash: %s\n", hex.EncodeToString(blk.Hash))
		fmt.Printf("Prev. Hash: %s\n", hex.EncodeToString(blk.PrevHash))
		fmt.Printf("Transaction Count: %d\n", len(blk.Transactions))
		for i, tx := range blk.Transactions {
			fmt.Printf("--------- TRANSACTION #%d ---------\n", i)
			fmt.Printf("TX ID: %s\n", hex.EncodeToString(tx.ID))
			fmt.Printf("Output count: %d\n", len(tx.Vout))
			for outIdx, out := range tx.Vout {
				fmt.Printf("Output #%d Value is: %d\n", outIdx, out.Value)
				fmt.Printf("Output #%d PubKey is: %s\n", outIdx, out.PubKeyHash)
			}
			fmt.Printf("Input count: %d\n", len(tx.Vin))
			for inIdx, in := range tx.Vin {
				fmt.Printf("Input #%d PubKey is %s\n", inIdx, in.PubKey)
			}
		}
		fmt.Println()

		if len(blk.PrevHash) == 0 {
			break
		}
	}
}
