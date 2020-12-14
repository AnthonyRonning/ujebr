package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "UJEBR",
	Long:  `UJEBR program`,
}

func main() {
	viper.SetEnvPrefix("ujebr") // Set the environment prefix to UJEBR_*
	viper.AutomaticEnv()        // Automatically search for environment variables

	rootCmd.PersistentFlags().String("bitcoind.url", "", "url for bitcoind")
	viper.BindPFlag("bitcoind.url", rootCmd.PersistentFlags().Lookup("bitcoind.url"))
	viper.SetDefault("bitcoind.url", "127.0.0.1")

	rootCmd.PersistentFlags().String("bitcoind.port", "", "port for bitcoind")
	viper.BindPFlag("bitcoind.port", rootCmd.PersistentFlags().Lookup("bitcoind.port"))
	viper.SetDefault("bitcoind.port", 18332)

	rootCmd.PersistentFlags().String("bitcoind.user", "", "user for bitcoind")
	viper.BindPFlag("bitcoind.user", rootCmd.PersistentFlags().Lookup("bitcoind.user"))
	viper.SetDefault("bitcoind.user", "")

	rootCmd.PersistentFlags().String("bitcoind.pass", "", "pass for bitcoind")
	viper.BindPFlag("bitcoind.pass", rootCmd.PersistentFlags().Lookup("bitcoind.pass"))
	viper.SetDefault("bitcoind.pass", "")

	rootCmd.PersistentFlags().String("bwt.url", "", "url for bwt")
	viper.BindPFlag("bwt.url", rootCmd.PersistentFlags().Lookup("bwt.url"))
	viper.SetDefault("bwt.url", "http://127.0.0.1")

	rootCmd.PersistentFlags().String("bwt.port", "", "port for bwt")
	viper.BindPFlag("bwt.port", rootCmd.PersistentFlags().Lookup("bwt.port"))
	viper.SetDefault("bwt.port", "3060")

	rootCmd.PersistentFlags().String("recover_address", "", "recover address to send funds to")
	viper.BindPFlag("recover_address", rootCmd.PersistentFlags().Lookup("recover_address"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
