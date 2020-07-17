package cli

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

const defaultConfigFile = "./config.toml"
const defaultWalletPath = "./wallet/"
const defaultRPCURL = "https://rpc1.newchain.newtonproject.org"
const defaultContractAddress = ""

func defaultConfig(cli *CLI) {
	viper.BindPFlag("walletPath", cli.rootCmd.PersistentFlags().Lookup("walletPath"))
	viper.BindPFlag("rpcURL", cli.rootCmd.PersistentFlags().Lookup("rpcURL"))
	viper.BindPFlag("contractAddress", cli.rootCmd.PersistentFlags().Lookup("contractAddress"))
	viper.BindPFlag("from", cli.rootCmd.PersistentFlags().Lookup("from"))

	viper.SetDefault("walletPath", defaultWalletPath)
	viper.SetDefault("rpcURL", defaultRPCURL)
	viper.SetDefault("contractAddress", defaultContractAddress)
}

func setupConfig(cli *CLI) error {

	//var ret bool
	var err error

	defaultConfig(cli)

	viper.SetConfigName(defaultConfigFile)
	viper.AddConfigPath(".")
	cfgFile := cli.config
	if cfgFile != "" {
		if _, err = os.Stat(cfgFile); err == nil {
			viper.SetConfigFile(cfgFile)
			err = viper.ReadInConfig()
		} else {
			// The default configuration is enabled.
			//fmt.Println(err)
			err = nil
		}
	} else {
		// The default configuration is enabled.
		err = nil
	}

	if rpcURL := viper.GetString("rpcURL"); rpcURL != "" {
		cli.rpcURL = rpcURL
	}
	if faucet := viper.GetString("faucet"); faucet != "" {
		cli.faucet = faucet
	}
	if cli.faucet == "" {
		cli.faucet = cli.rpcURL
	}
	if walletPath := viper.GetString("walletPath"); walletPath != "" {
		cli.walletPath = walletPath
	}
	if contractAddress := viper.GetString("contractAddress"); common.IsHexAddress(contractAddress) {
		cli.contractAddress = common.HexToAddress(contractAddress)
	}
	if fromAddress := viper.GetString("from"); common.IsHexAddress(fromAddress) {
		cli.address = common.HexToAddress(fromAddress)
	}
	if fromAddress := viper.GetString("address"); common.IsHexAddress(fromAddress) {
		cli.address = common.HexToAddress(fromAddress)
	}

	return err
}
