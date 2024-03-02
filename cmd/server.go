package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"through/log"
	"through/server"
	"time"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start proxy server",
	Long:  `start proxy server.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
		defer stop()

		s, err := server.NewServer(ctx)
		if err != nil {
			log.Errorf("new server error: %v", err)
			return
		}

		// start server
		if err := s.Start(); err != nil {
			log.Errorf("server start error: %v", err)
			return
		}

		s.Stop()
		time.Sleep(1 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
