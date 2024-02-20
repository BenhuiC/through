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

		// start server
		client.Start(ctx)

		<-ctx.Done()
		log.Info("client stopping ...")
		time.Sleep(3 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
