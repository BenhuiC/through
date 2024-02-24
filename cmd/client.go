package cmd

import (
	"context"
	"os"
	"os/signal"
	"through/client"
	"through/log"
	"time"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "start proxy client",
	Long:  `start proxy client.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
		defer stop()

		// start client
		c, err := client.NewClient(ctx)
		if err != nil {
			log.Errorf("new client error: %v", err)
			return
		}

		if err = c.Start(); err != nil {
			log.Errorf("client start error: %v", err)
			return
		}

		c.Stop()
		time.Sleep(3 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
