package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = cobra.Command{
		Use: "blem",
		Short: "Blemflarck is a cryptocurrency based on the X web",
		Long: "Blemflarck is the cryptocurrency for the X web. Build with love and dedication",
	}
)

func init() {
	// flags and parameters of the create-chain cmd
	createChainCmd.Flags().StringVarP(&createChainAddress, "address", "a", "",  "Address to send genesis reward")
	createChainCmd.MarkFlagRequired("address")

	// flags and parameters of the send cmd
	sendCmd.Flags().StringVarP(&sendFrom, "from", "f", "", "Address of the sender")
	sendCmd.Flags().StringVarP(&sendTo, "to", "t", "", "Address of the receiver")
	sendCmd.Flags().IntVarP(&sendAmount, "amount", "a", 0,"Amount being transferred")
	sendCmd.MarkFlagRequired("from")
	sendCmd.MarkFlagRequired("to")
	sendCmd.MarkFlagRequired("amount")

	// Add the commands to the root command. This allows them to be executable.
	rootCmd.AddCommand(printWalletCmd)
	rootCmd.AddCommand(createWalletCmd)
	rootCmd.AddCommand(createChainCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(printChainCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}