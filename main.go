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

	bc, err := core.CreateBlockchain()
	if err != nil {
		log.Fatal(err)
	}

	if err := bc.AddBlock([]byte("dope as heck")); err != nil {
		fmt.Println(err)
	}
	if err := bc.AddBlock([]byte("dope as heck")); err != nil {
		fmt.Println(err)
	}
	if err := bc.AddBlock([]byte("dope as heck")); err != nil {
		fmt.Println(err)
	}

	iter, err := bc.NewIterator()
	if err != nil {
		log.Fatal(err)
	}

	for {
		blk := iter.Next()

		fmt.Printf("Block height: %d\n", blk.Height)
		fmt.Printf("Hash: %s\n", hex.EncodeToString(blk.Hash))
		fmt.Printf("Prev. Hash: %s\n", hex.EncodeToString(blk.PrevHash))
		fmt.Printf("Data: %s\n", blk.Data)
		fmt.Println()

		if len(blk.PrevHash) == 0 {
			break
		}
	}
}