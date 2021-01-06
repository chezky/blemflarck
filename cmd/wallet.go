package cmd

import (
	"fmt"
	"github.com/chezky/blemflarck/core"
	"github.com/spf13/cobra"
	"log"
)

var (
	createWalletCmd = &cobra.Command{
		Use: "create-wallet",
		Short: "create a new blemflarck wallet",
		Run: createWallet(),
	}

	printWalletCmd = &cobra.Command{
		Use: "print-wallets",
		Short: "print all the stored wallet addresses",
		Run: printWallets(),
	}
)

func createWallet() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		wallet, err := core.CreateWallet()
		if err != nil {
			log.Fatal("error creating wallet: ", err)
		}
		address, _ := wallet.GetAddress()
		fmt.Printf("Your wallet address is: %s\n", address)

		wallets, err := core.ReadWalletsFromFile()
		if err != nil {
			log.Fatal("error reading in wallets from file: ", err)
		}

		wallets.Wallets[string(address)] = wallet

		err = wallets.SaveToFile()
		if err != nil {
			fmt.Printf("error saving wallet to file during createWallet: %v\n", err)
		}
	}
}

func printWallets() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		wallets, err := core.ReadWalletsFromFile()
		if err != nil {
			log.Fatal("error reading in wallets from file: ", err)
		}

		idx := 1
		for add, _ := range wallets.Wallets {
			fmt.Printf("Wallet #%d address is: %s\n", idx, add)
			idx++
		}
	}
}