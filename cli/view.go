package cli

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "view <functionName> [arg1Type arg1Value] [arg2Type arg2Value]...",
		Short:                 "Get info from the contract by function name and args",
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]


			var valueArgs []string
			var inputTypeArgs abi.Arguments
			if len(args) > 1 {

				argsLen := len(args)
				if argsLen%2 == 0 {
					fmt.Println("len error ", argsLen)
					return
				}

				for i := 1; i < argsLen-1; i += 2 {
					argType, err := abi.NewType(args[i], "", nil)
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

			method := abi.NewMethod(name, name, abi.Function, "view", false, false, inputTypeArgs, outTypeArgs)

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

			outByte, err := cli.view(method, inputArgs...)
			if err != nil {
				fmt.Printf("Error: view function error(%v)\n", err)
				return
			}
			if len(outByte) == 0 {
				fmt.Println("Function always returns null")
				return
			}

			cli.showOut(method, outByte)

			return

		},
	}

	cmd.Flags().StringP("out", "o", "", "the out type list of the method, spilt by ','")

	return cmd
}
