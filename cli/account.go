package cli

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [new|list]",
		Short: "Manage NewChain accounts",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			return
		},
	}

	cmd.AddCommand(cli.buildAccountNewCmd())
	cmd.AddCommand(cli.buildAccountListCmd())

	return cmd
}

func (cli *CLI) buildAccountNewCmd() *cobra.Command {
	accountNewCmd := &cobra.Command{
		Use:   "new [--faucet] [--numOfNew amount]",
		Short: "create a new account",
		Args:  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			walletPath := cli.walletPath
			wallet := keystore.NewKeyStore(walletPath,
				keystore.LightScryptN, keystore.LightScryptP)

			if cli.walletPassword == "" {
				cli.walletPassword, err = getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
				if err != nil {
					fmt.Println("Error: ", err)
					return
				}
			}

			numOfNew, err := cmd.Flags().GetInt("numOfNew")
			if err != nil {
				numOfNew = viper.GetInt("account.numOfNew")
			}
			if numOfNew <= 0 {
				fmt.Printf("number[%d] of new account less then 1\n", numOfNew)
				numOfNew = 1
			}

			faucet, _ := cmd.Flags().GetBool("faucet")

			for i := 0; i < numOfNew; i++ {
				account, err := wallet.NewAccount(cli.walletPassword)
				if err != nil {
					fmt.Println("Account error:", err)
					return
				}
				if faucet {
					getFaucet(cli.faucet, account.Address.String())
				}
				fmt.Println(account.Address.Hex())
				if cli.address == (common.Address{}) {
					cli.address = account.Address
				}
			}
		},
	}

	accountNewCmd.Flags().IntP("numOfNew", "n", 1, "number of the new account")
	accountNewCmd.Flags().Bool("faucet", false, "get faucet for new account")
	return accountNewCmd
}

func (cli *CLI) buildAccountListCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:   "list",
		Short: "list all accounts in the wallet path",
		Args:  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			walletPath := cli.walletPath
			wallet := keystore.NewKeyStore(walletPath,
				keystore.LightScryptN, keystore.LightScryptP)
			if len(wallet.Accounts()) == 0 {
				fmt.Println("Empty wallet, create account first.")
				return
			}

			for _, account := range wallet.Accounts() {
				fmt.Println(account.Address.Hex())
			}
		},
	}

	return accountListCmd
}
