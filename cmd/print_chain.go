package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/chezky/blemflarck/core"
	"github.com/spf13/cobra"
	"log"
)

var (
	printChainCmd = &cobra.Command{
		Use: "print-chain",
		Short: "Print out the current blockchain",
		Run: printChain(),
	}
)

func printChain() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !core.ChainExists() {
			log.Fatal("No blockchain exists yet! Create with the 'create-chain' cmd.")
		}
		bc, err := core.CreateBlockchain("")
		if err != nil {
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
					fmt.Printf("Output #%d PubKeyHash is: %s\n", outIdx, hex.EncodeToString(out.PubKeyHash))
				}
				fmt.Printf("Input count: %d\n", len(tx.Vin))
				for inIdx, in := range tx.Vin {
					fmt.Printf("Input #%d PubKey: %s\n", inIdx, hex.EncodeToString(in.PubKey))
					fmt.Printf("Input #%d Signature: %s\n", inIdx, hex.EncodeToString(in.Signature))
				}
			}
			fmt.Println()

			if len(blk.PrevHash) == 0 {
				break
			}
		}
	}
}
