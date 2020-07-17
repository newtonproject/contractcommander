package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
)

func (cli *CLI) deploySol(solFlag, contractName string, args []string, solc string) error {
	var contracts map[string]*compiler.Contract
	var err error
	var names []string

	solFlagSlice := strings.Split(solFlag, ",")
	contracts, err = compiler.CompileSolidity(solc, solFlagSlice...)
	if err != nil {
		return err
	}

	for name, contract := range contracts {
		nameParts := strings.Split(name, ":")
		namePart := nameParts[len(nameParts)-1]
		names = append(names, namePart)
		if namePart == contractName { // contractName
			abiByte, err := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
			if err != nil {
				return err
			}
			parsed, err := abi.JSON(strings.NewReader(string(abiByte)))
			if err != nil {
				return err
			}

			constructorArgs, err := getConstructorArgs(parsed.Constructor.Inputs, args)
			if err != nil {
				if len(parsed.Constructor.Inputs) > 0 {
					var argName []string
					for _, input := range parsed.Constructor.Inputs {
						argName = append(argName, input.Name+" "+input.Type.String())
					}
					return fmt.Errorf("%v(%v)", err.Error(), strings.Join(argName, ", "))
				}
				return err

			}
			if len(constructorArgs) > 0 {
				fmt.Printf("The contract %s will be deployed with args as follow:\n", contractName)
				func(inputs abi.Arguments, constructorArgs []interface{}) {
					if len(inputs) != len(constructorArgs) {
						fmt.Println("get args error")
						return
					}
					for i, input := range inputs {
						if input.Type.T == abi.AddressTy {
							fmt.Printf("\t%v(%v): %v\n", input.Name, input.Type, constructorArgs[i].(common.Address).String())
						} else if addressSlice, ok := constructorArgs[i].([]common.Address); ok {
							var addressArray []string
							for _, address := range addressSlice {
								addressArray = append(addressArray, address.String())
							}
							fmt.Printf("\t%v(%v): %v\n", input.Name, input.Type, strings.Join(addressArray, ","))
						} else {
							fmt.Printf("\t%v(%v): %v\n", input.Name, input.Type, constructorArgs[i])
						}
					}
				}(parsed.Constructor.Inputs, constructorArgs)
			} else {
				fmt.Printf("The contract %s will be deployed with no args\n", contractName)
			}

			if err := cli.deployContract(parsed, common.FromHex(contract.Code), constructorArgs); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("no the given contract name, name list: %v", names[:])
}

func getConstructorArgs(inputs abi.Arguments, args []string) ([]interface{}, error) {
	if len(inputs) != len(args) {
		return nil, errors.New("args length error")
	}

	if len(inputs) == 0 {
		return nil, nil
	}

	var cArgs []interface{}

	for i, arg := range args {
		value, err := getValueByAbiType(inputs[i].Type, arg)
		if err != nil {
			return nil, err
		}
		cArgs = append(cArgs, value)
	}

	return cArgs, nil
}

func getValueByAbiType(t abi.Type, value string) (interface{}, error) {
	switch t.T {
	case abi.SliceTy:
		valueSlice := strings.Split(value, ",")
		refSlice := reflect.MakeSlice(t.Type, len(valueSlice), len(valueSlice))
		for i, v := range valueSlice {
			ret, err := getValueByAbiType(*t.Elem, v)
			if err != nil {
				return nil, err
			}
			refSlice.Index(i).Set(reflect.ValueOf(ret))
		}
		return refSlice.Interface(), nil
	case abi.ArrayTy:
		valueSlice := strings.Split(value, ",")
		refSlice := reflect.New(t.Type).Elem()
		for i, v := range valueSlice {
			ret, err := getValueByAbiType(*t.Elem, v)
			if err != nil {
				return nil, err
			}
			refSlice.Index(i).Set(reflect.ValueOf(ret))
		}
		return refSlice.Interface(), nil
	case abi.StringTy: // variable arrays are written at the end of the return bytes
		return value, nil
	case abi.IntTy, abi.UintTy:
		if ret, ok := big.NewInt(0).SetString(value, 10); ok {
			switch t.Type.Kind() {
			case reflect.Ptr: // *big.Int
				return ret, nil
			case reflect.Int:
				return int(ret.Int64()), nil
			case reflect.Int8:
				return int8(ret.Int64()), nil
			case reflect.Int16:
				return int16(ret.Int64()), nil
			case reflect.Int32:
				return int32(ret.Int64()), nil
			case reflect.Int64:
				return int64(ret.Int64()), nil
			case reflect.Uint:
				return uint(ret.Int64()), nil
			case reflect.Uint8:
				return uint8(ret.Int64()), nil
			case reflect.Uint16:
				return uint16(ret.Int64()), nil
			case reflect.Uint32:
				return uint32(ret.Int64()), nil
			case reflect.Uint64:
				return uint64(ret.Int64()), nil
			}
		}
	case abi.BoolTy:
		if value == "true" {
			return true, nil
		} else if value == "false" {
			return false, nil
		}
	case abi.AddressTy:
		if common.IsHexAddress(value) {
			return common.HexToAddress(value), nil
		}
	case abi.HashTy:
		ret := common.HexToHash(value)
		if ret != (common.Hash{}) {
			return ret, nil
		}
	case abi.BytesTy:
		ret := common.FromHex(value)
		if len(ret) > 0 {
			return ret, nil
		}
	case abi.FixedBytesTy:
		ret := common.FromHex(value)
		size := len(ret)
		if size > t.Size {
			size = t.Size
		}
		array := reflect.New(t.Type).Elem()
		reflect.Copy(array, reflect.ValueOf(ret[0:size]))
		return array.Interface(), nil
	case abi.FunctionTy:
		return nil, errors.New("not support FunctionTy")
	default:
		return nil, fmt.Errorf("unknown type %v", t.T)
	}

	return nil, fmt.Errorf("get value %s as type %v error", value, t.Type.String())
}

func (cli *CLI) deployContract(parsed abi.ABI, bytecode []byte, params []interface{}) error {
	opts, err := cli.getTransactOpts("")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	opts.Context = ctx

	cli.BuildClient()
	client := cli.client

	contractAddress, tx, _, err := bind.DeployContract(opts, parsed, bytecode, client, params...)
	if err != nil {
		return err
	}

	fmt.Printf("Contract deploy at address %s\n", contractAddress.String())
	fmt.Printf("Transaction waiting to be mined: 0x%x\n", tx.Hash())
	cli.contractAddress = contractAddress
	bind.WaitDeployed(opts.Context, client, tx)

	fmt.Println("Contract deploy success")

	return nil
}
