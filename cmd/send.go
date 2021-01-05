package cmd

import (
	"fmt"
	"github.com/chezky/blemflarck/core"
	"github.com/spf13/cobra"
	"log"
)

var (
	sendTo string
	sendFrom string
	sendAmount int

	sendCmd = &cobra.Command{
		Use: "send",
		Short: "Send blemflarcks from one address to another",
		Long: "Create a transfer between address A to address B.",
		Run: send(),
	}
)

func send() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		bc, err := core.CreateBlockchain(sendFrom)
		if err != nil {
			log.Fatal(err)
		}
		tx, err := bc.NewTransaction(sendFrom, sendTo, sendAmount)
		if err != nil {
			log.Fatal(err)
		}
		cbTX, err := core.NewCoinbaseTransaction([]byte(sendFrom))
		if err != nil {
			log.Fatal(err)
		}
		if err := bc.AddBlock([]core.Transaction{tx, cbTX}); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Successfully sent %d coins from address: \n%s\n to address: \n%s\n", sendAmount, sendFrom, sendTo)
	}
}
