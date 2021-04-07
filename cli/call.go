package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildCallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "call <functionName> [arg1Type arg1Value] [arg2Type arg2Value]... [--view] [--out outType]",
		Short:                 "Call functions with args type and value",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Example: fmt.Sprintf(`%s call transfer address 0x4Ba80F138543E75AbF788eB3fE2726425586b0ff uint256 1
%s call totalSupply --view --out uint256123
%s call name --view --out string
%s call balanceOf address 0x4Ba80F138543E75AbF788eB3fE2726425586b0fD --view --out uint256`,
			cli.Name, cli.Name, cli.Name, cli.Name),
		Run: func(cmd *cobra.Command, args []string) {
			view, _ := cmd.Flags().GetBool("view")

			unit, err := cmd.Flags().GetString("unit")
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}
			if !stringInSlice(unit, UnitList) {
				fmt.Println("Error: ", errIllegalUnit)
				return
			}

			amountStr, err := cmd.Flags().GetString("value")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			amountWei, err := getAmountWei(amountStr, unit)
			if err != nil {
				fmt.Println("Error: ", errIllegalAmount)
				return
			}

			var valueArgs []string
			var inputTypeArgs abi.Arguments
			if len(args) > 1 {

				argsLen := len(args)
				if argsLen%2 == 0 {
					fmt.Println("Error: len error ", argsLen, args)
					return
				}

				for i := 1; i < argsLen-1; i += 2 {
					arg := args[i]
					if arg == "uint" {
						arg = "uint256"
					} else if arg == "int" {
						arg = "int256"
					}
					argType, err := abi.NewType(arg, "", nil)
					if err != nil {
						fmt.Println(err)
						return
					}
					argArg := abi.Argument{Type: argType}
					inputTypeArgs = append(inputTypeArgs, argArg)
					valueArgs = append(valueArgs, args[i+1])
				}
			}

			var outTypeArgs abi.Arguments
			if cmd.Flags().Changed("out") {
				if !view {
					fmt.Println("Error: --view not use")
					return
				}
				outTypes, err := cmd.Flags().GetString("out")
				if err != nil {
					fmt.Println("Error: ", err)
					return
				}
				outTypeList := strings.Split(outTypes, ",")

				for _, outTypeStr := range outTypeList {
					if outTypeStr[0] == '[' {
						fmt.Println("Error: unsupported arg type:", outTypeStr)
						return
					}
					if outTypeStr == "uint" {
						outTypeStr = "uint256"
					} else if outTypeStr == "int" {
						outTypeStr = "int256"
					}

					outType, err := abi.NewType(outTypeStr, "", nil)
					if err != nil {
						fmt.Println(err)
						return
					}
					outTypeArgs = append(outTypeArgs, abi.Argument{
						Type: outType,
					})
				}

			}

			name := args[0]
			method := abi.NewMethod(name, name, abi.Function, "", false, false, inputTypeArgs, outTypeArgs)

			if err := cli.BuildClient(); err != nil {
				fmt.Println(err)
				return
			}
			client := cli.client

			// view: abi will error when no out type
			// abi
			var bContract *bind.BoundContract
			if !view {
				a := abi.ABI{}
				a.Methods = make(map[string]abi.Method)
				a.Methods[method.Name] = method
				bContract = bind.NewBoundContract(cli.contractAddress, a, client, client, client)
			}

			// input args
			inputArgs, err := getConstructorArgs(method.Inputs, valueArgs)
			if err != nil {
				if len(method.Inputs) > 0 {
					var argName []string
					for _, input := range method.Inputs {
						argName = append(argName, input.Name+" "+input.Type.String())
					}
					fmt.Println(fmt.Errorf("%v(%v)", err.Error(), strings.Join(argName, ", ")))
					return // fmt.Errorf("%v(%v)", err.Error(), strings.Join(argName, ", "))
				}
				fmt.Println(err)
				return // err

			}

			if view {
				outByte, err := cli.view(method, inputArgs...)
				if err != nil {
					fmt.Printf("Error1: view function error(%v)\n", err)
					return
				}
				if len(outByte) == 0 {
					fmt.Println("Function always returns null")
					return
				}

				cli.showOut(method, outByte)
				return
			}

			var gasLimit uint64
			if cmd.Flags().Changed("gasLimit") {
				gasLimit, err = cmd.Flags().GetUint64("gasLimit")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			opts, err := cli.getTransactOpts("", gasLimit)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}
			ctx := context.Background()
			opts.Context = ctx
			opts.Value = amountWei
			// if amountWei.Cmp(big.NewInt(0)) > 0 {
			// 	method.Payable = true
			// }
			if cmd.Flags().Changed("gasPrice") {
				gasPriceStr, err := cmd.Flags().GetString("gasPrice")
				if err != nil {
					fmt.Println("Flags GetString error: ", err)
					return
				}
				gasPrice, err := getAmountWei(gasPriceStr, UnitETH)
				if err != nil {
					fmt.Println("getAmountWei error: ", err)
					return
				}
				opts.GasPrice = gasPrice
			}

			tx, err := bContract.Transact(opts, method.Name, inputArgs...)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(tx.Hash().String())

			fmt.Printf("Transaction waiting to be mined: 0x%x\n", tx.Hash())
			if _, err := bind.WaitMined(ctx, client, tx); err != nil {
				fmt.Printf("Error: wait tx mined error(%v)\n", err)
				return
			}
			showTransactionReceipt(cli.rpcURL, tx.Hash().String())

			fmt.Println("Call function success")

			return
		},
	}

	cmd.Flags().BoolP("view", "v", false, "only view function and get output")
	cmd.Flags().StringP("out", "o", "", "the out type list of the method, spilt by ',', only use with --view")
	// cmd.Flags().Bool("force", false, "force execute function")
	cmd.Flags().String("value", "", "the amount of unit send to the contract address")
	cmd.Flags().StringP("unit", "u", UnitETH, fmt.Sprintf("unit for send value. %s.", UnitString))

	cmd.Flags().StringP("gasPrice", "p", "", "the gas price in ETH")
	cmd.Flags().Uint64P("gasLimit", "g", 21000, "the gas limit")

	return cmd
}

func (cli *CLI) view(method abi.Method, params ...interface{}) ([]byte, error) {
	inputTypeArgsByte, err := method.Inputs.Pack(params...)
	if err != nil {
		return nil, err
	}

	input := append(method.ID, inputTypeArgsByte...)
	// fmt.Printf("input： 0x%x\n", input)

	msg := ethereum.CallMsg{From: cli.address, To: &cli.contractAddress, Data: input}
	ctx := context.TODO() // context.Background()
	if err := cli.BuildClient(); err != nil {
		return nil, err
	}

	if err := cli.BuildClient(); err != nil {
		return nil, err
	}
	return cli.client.CallContract(ctx, msg, nil)
}

func (cli *CLI) showOut(method abi.Method, outByte []byte) {
	if len(outByte) == 0 {
		fmt.Println("function return nil")
		return
	}

	if len(method.Outputs) == 0 {
		fmt.Printf("0x%x\n", outByte)
		return
	}

	out, err := method.Outputs.UnpackValues(outByte)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range out {
		showValue(v)
	}

	return
}

func showValue(v interface{}) {
	if address, ok := v.(common.Address); ok {
		fmt.Println(address.String())
	} else if addressSlice, ok := v.([]common.Address); ok {
		for _, address := range addressSlice {
			fmt.Println(address.String())
		}
	} else {
		fmt.Println(v)
	}
}
