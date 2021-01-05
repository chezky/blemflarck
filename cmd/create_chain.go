package cmd

import (
	"github.com/chezky/blemflarck/core"
	"github.com/spf13/cobra"
	"log"
)

var (
	createChainAddress string

	createChainCmd = &cobra.Command{
		Use: "create-chain",
		Short: "Create a new blockchain",
		Long: "Create a new blockchain, if no current one exists, then the genesis reward will go to the address",
		Run: createChain(),
	}
)

func createChain() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !core.CheckValidAddress([]byte(createChainAddress)) {
			log.Fatal("Please enter a valid address!")
		}
		if _, err := core.CreateBlockchain(createChainAddress); err != nil {
			log.Fatal("error creating blockchain", err)
		}
	}
}