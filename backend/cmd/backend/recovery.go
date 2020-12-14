package main

import (
	"errors"
	"fmt"

	"github.com/anthonyronning/ujebr/backend"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	recoverCmd := &cobra.Command{
		Use:   "recover",
		Short: "recover a wallet",
		Run: func(cmd *cobra.Command, args []string) {
			if err := recoverAction(); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(recoverCmd)
}

func recoverAction() error {
	// config parameters
	bitcoinUrl := viper.GetString("bitcoind.url")
	bitcoinPort := viper.GetInt("bitcoind.port")
	bitcoinUser := viper.GetString("bitcoind.user")
	bitcoinPass := viper.GetString("bitcoind.pass")

	bwtUrl := viper.GetString("bwt.url")
	bwtPort := viper.GetInt("bwt.port")

	recoverAddress := viper.GetString("recover_address")
	if recoverAddress == "" {
		return errors.New("recover_address must be specified")
	}

	// init
	r, err := backend.NewRecovery(&backend.RecoveryCfg{
		BitcoindHost: bitcoinUrl,
		BitcoindPort: bitcoinPort,
		BitcoindUser: bitcoinUser,
		BitcoindPass: bitcoinPass,
		BwtUrl:       bwtUrl,
		BwtPort:      bwtPort,
	})
	if err != nil {
		return err
	}

	// recover
	unsignedTransactions, err := r.Recover(recoverAddress, "")
	if err != nil {
		return err
	}

	fmt.Sprintln("Here are the unsigned transactions that would be created:")
	for _, unsignedTransaction := range unsignedTransactions {
		fmt.Println(unsignedTransaction)
	}

	/*
		fmt.Print("If acceptable, enter seed phrase to recover: ")
		inputReader := bufio.NewReader(os.Stdin)
		input, _ := inputReader.ReadString('\n')
		fmt.Println(input)
	*/

	return nil
}
