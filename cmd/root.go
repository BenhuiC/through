package cmd

import (
	"fmt"
	"os"
	"through/config"
	"through/pkg"
	"through/pkg/log"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "through",
	Short: "A tool for bypass network restrictions",
	Long:  `A tool for bypass network restrictions.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cfgFile string

func init() {
	cobra.OnInitialize(Init)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./through.yaml", "config file (default is $HOME/.through.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Init() {
	var err error
	if err = config.Init(cfgFile); err != nil {
		fmt.Println("init config error:", err)
		return
	}

	if err = log.Init(); err != nil {
		fmt.Println("init log error:", err)
		return
	}
	pkg.InitMetrics()
}
