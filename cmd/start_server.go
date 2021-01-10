package cmd

import (
	"github.com/chezky/blemflarck/p2p"
	"github.com/spf13/cobra"
	"log"
)

var (
	startServerCmd = &cobra.Command{
		Use: "start",
		Short: "Start a node",
		Run: startServer(),
	}
)

func startServer() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		err := p2p.StartServer()
		if err != nil {
			log.Fatal(err)
		}
	}
}
