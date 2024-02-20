package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"through/log"
	"through/server"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start proxy server",
	Long:  `start proxy server.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
		defer stop()

		// start server
		if err := server.Start(ctx); err != nil {
			log.Error("server start error: %v", err)
			return
		}

		<-ctx.Done()
		server.Stop()
		log.Info("server stopping ...")
		//time.Sleep(3 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
