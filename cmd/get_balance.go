package cmd

import (
	"fmt"
	"github.com/chezky/blemflarck/core"
	"github.com/spf13/cobra"
	"log"
)

var (
	getBalanceAddress string

	getBalanceCmd = &cobra.Command{
		Use: "get-balance",
		Short: "Get the balance of an address",
		Run: getBalance(),
	}
)

func getBalance() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !core.ChainExists() {
			log.Fatal("Chain does not exist! Please create one first.")
		}
		bc, err := core.CreateBlockchain("")
		if err != nil {
			log.Fatal(err)
		}

		utxo := core.UTXO{Blockchain: bc}

		UTXOs, err := utxo.FindUTXOs()
		if err != nil {
			log.Fatal(err)
		}

		acc := 0

		for _, outs := range UTXOs {
			for _, out := range outs.Outputs {
				if out.CanBeUnlocked([]byte(getBalanceAddress)) {
					acc += out.Value
				}
			}
		}

		fmt.Printf("Total balance for address %s is %d blemflarck(s)\n", getBalanceAddress, acc)
	}
}