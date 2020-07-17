package cli

import (
	"github.com/spf13/cobra"
)

func (cli *CLI) buildRootCmd() {

	if cli.rootCmd != nil {
		cli.rootCmd.ResetFlags()
		cli.rootCmd.ResetCommands()
	}

	rootCmd := &cobra.Command{
		Use:              cli.Name,
		Short:            cli.Name + " is commandline client for users to interact with the SimpleToken contract.",
		Run:              cli.help,
		PersistentPreRun: cli.setup,
	}
	cli.rootCmd = rootCmd

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cli.config, "config", "c", defaultConfigFile, "The `path` to config file")
	rootCmd.PersistentFlags().StringP("walletPath", "w", defaultWalletPath, "Wallet storage `directory`")
	rootCmd.PersistentFlags().StringP("rpcURL", "i", defaultRPCURL, "Geth json rpc or ipc `url`")
	rootCmd.PersistentFlags().StringP("contractAddress", "a", defaultContractAddress, "Contract `address`")
	rootCmd.PersistentFlags().StringP("from", "f", "", "the from `address` who pay gas")

	// Basic commands
	rootCmd.AddCommand(cli.buildInitCmd())    // init
	rootCmd.AddCommand(cli.buildVersionCmd()) // version

	// account
	rootCmd.AddCommand(cli.buildAccountCmd())

	// Aux commands
	rootCmd.AddCommand(cli.buildBalanceCmd()) // balance
	rootCmd.AddCommand(cli.buildFaucetCmd())  // faucet

	// deploy
	rootCmd.AddCommand(cli.buildDeployCmd())

	// call functions
	rootCmd.AddCommand(cli.buildCallCmd())

	// view functions
	rootCmd.AddCommand(cli.buildViewCmd())
}
