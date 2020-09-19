package cli

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "deploy <--sol source1.sol,source2.sol> <--name contractName> [arg1] [arg2]...",
		Short:                 "Deploy NewChain contract",
		DisableFlagsInUseLine: true,
		Example: fmt.Sprintf(`%s deploy --sol SimpleToken.sol --name SimpleToken HelloToken HT 18 1000000000000000000"
%s deploy --abi SimpleToken.abi --bin SimpleToken.bin --name SimpleToken HelloToken HT 18 1000000000000000000`, cli.Name, cli.Name),
		Run: func(cmd *cobra.Command, args []string) {
			save, _ := cmd.Flags().GetBool("save")
			fromAddress := viper.GetString("from")
			cli.address = common.HexToAddress(fromAddress)
			if cli.address == (common.Address{}) {
				fmt.Println("Error: not set from address of owner")
				fmt.Println(cmd.UsageString())
				return
			}

			if cli.contractAddress == (common.Address{}) {
				save = true
			}

			if cmd.Flags().Changed("sol") {
				if cmd.Flags().Changed("bin") || cmd.Flags().Changed("abi") {
					fmt.Println("`sol` cannot be used at the same time with `bin` or `abi")
					return
				}

				solFile, err := cmd.Flags().GetString("sol")
				if err != nil || solFile == "" {
					fmt.Println("Error: not set file of contract source")
					fmt.Println(cmd.UsageString())
					return
				}
				contractName, err := cmd.Flags().GetString("name")
				if err != nil || contractName == "" {
					fmt.Println("Error: not set file of contract source")
					fmt.Println(cmd.UsageString())
					return
				}

				solc, _ := cmd.Flags().GetString("solc")
				if err := cli.deploySol(solFile, contractName, args, solc); err != nil {
					fmt.Println("Error: ", err)
					return
				}
			} else {
				if !cmd.Flags().Changed("bin") || !cmd.Flags().Changed("abi") {
					fmt.Println("`bin` and `abi` must be used at the same time")
					return
				}

				binFile, err := cmd.Flags().GetString("bin")
				if err != nil || binFile == "" {
					fmt.Println("Error: not set file of bin or set to empty")
					fmt.Println(cmd.UsageString())
					return
				}
				abiFile, err := cmd.Flags().GetString("abi")
				if err != nil || abiFile == "" {
					fmt.Println("Error: not set file of abi or set to empty")
					fmt.Println(cmd.UsageString())
					return
				}

				if err := cli.deploySolFromBinAndABI(binFile, abiFile, args); err != nil {
					fmt.Println("Error: ", err)
					return
				}
			}

			if save {
				viper.Set("contractaddress", cli.contractAddress.String())
				viper.WriteConfigAs(cli.config)
			}
		},
	}

	cmd.Flags().StringP("sol", "s", "", "the path of the contract source")
	cmd.Flags().StringP("name", "n", "", "the name of the contract to deploy")
	cmd.Flags().Bool("save", false, "save contract address to config file")
	cmd.Flags().String("solc", "solc", "solidity compiler to use if source builds are requested")

	cmd.Flags().String("bin", "", "the path of the binary of the contracts in hex")
	cmd.Flags().String("abi", "", "the path of the ABI specification of the contracts")

	return cmd
}
