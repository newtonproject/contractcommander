package cli

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "balance [--unit NEW|WEI] [address1] [address2]...",
		Short:                 "Get balance of address",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

			unit, _ := cmd.Flags().GetString("unit")
			if unit != "" && !stringInSlice(unit, UnitList) {
				fmt.Printf("Unit(%s) for invalid. %s.\n", unit, UnitString)
				fmt.Fprint(os.Stderr, cmd.UsageString())
				return
			}

			var addressList []common.Address

			if len(args) <= 0 {
				if err := cli.openWallet(true); err != nil {
					fmt.Println(err)
					return
				}

				for _, account := range cli.wallet.Accounts() {
					addressList = append(addressList, account.Address)
				}

			} else {
				for _, addressStr := range args {
					addressList = append(addressList, common.HexToAddress(addressStr))
				}
			}

			for _, address := range addressList {
				balance, err := cli.getBalance(address)
				if err != nil {
					fmt.Println("Balance error:", err)
					return
				}
				fmt.Printf("Address[%s] Balance[%s]\n", address.Hex(), getWeiAmountTextUnitByUnit(balance, unit))
			}

			return
		},
	}

	cmd.Flags().StringP("unit", "u", "", fmt.Sprintf("unit for balance. %s.", UnitString))

	return cmd
}

func (cli *CLI) getBalance(address common.Address) (*big.Int, error) {
	if err := cli.BuildClient(); err != nil {
		return nil, err
	}
	return cli.client.BalanceAt(context.Background(), address, nil)
}
