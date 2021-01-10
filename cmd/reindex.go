package cmd

import (
	"github.com/chezky/blemflarck/core"
	"github.com/spf13/cobra"
	"log"
)

var (
	reindexCmd = &cobra.Command{
		Use: "reindex",
		Short: "Reindex the UTXO's",
		Long: "Reindex the unspent transaction outputs of the chain. This could take a bit of time, as it runs through the entire blockchain",
		Run: reindex(),
	}
)

func reindex() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !core.ChainExists() {
			log.Fatal("No blockchain exists yet! Create with the 'create-chain' cmd.")
		}

		bc, err := core.CreateBlockchain("")
		if err != nil {
			log.Fatal(err)
		}

		utxo := core.UTXO{Blockchain: bc}

		if err := utxo.Reindex(); err != nil {
			log.Fatal(err)
		}
	}
}