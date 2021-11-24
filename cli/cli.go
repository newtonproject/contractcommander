package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

var (
	buildCommit string
	buildDate   string
)

// CLI represents a command-line interface. This class is
// not threadsafe.
type CLI struct {
	Name       string
	rootCmd    *cobra.Command
	version    string
	walletPath string
	rpcURL     string
	faucet     string
	config     string
	//testing    bool

	contractAddress common.Address
	client          *ethclient.Client
	wallet          *keystore.KeyStore
	account         accounts.Account
	walletPassword  string
	address         common.Address
}

// NewCLI returns an initialized CLI
func NewCLI() *CLI {
	version := "v0.6.4"
	if buildCommit != "" {
		version = fmt.Sprintf("%s-%s", version, buildCommit)
	}
	if buildDate != "" {
		version = fmt.Sprintf("%s-%s", version, buildDate)
	}

	bc, _ := getBlockChain()
	version = fmt.Sprintf("%s-%s", version, bc.String())

	// init BlockChain
	bc.Init()

	cli := &CLI{
		Name:       "contractcommander",
		rootCmd:    nil,
		version:    version,
		walletPath: "",
		rpcURL:     "",
		//	testing:         false,
		config:         "",
		client:         nil,
		walletPassword: "",
	}

	cli.buildRootCmd()
	return cli
}

// BuildClient BuildClient
func (cli *CLI) BuildClient() error {
	var err error
	if cli.client == nil {
		cli.client, err = ethclient.Dial(cli.rpcURL)
		if err != nil {
			return fmt.Errorf("Failed to connect to the NewChain client: %v", err)
		}
	}
	return nil
}

func (cli *CLI) openWallet(check bool) error {
	if cli.wallet == nil {
		cli.wallet = keystore.NewKeyStore(cli.walletPath,
			keystore.LightScryptN, keystore.LightScryptP)
	}

	if check && len(cli.wallet.Accounts()) == 0 {
		return errors.New("empty wallet, create account first")
	}
	return nil
}

func (cli *CLI) buildWallet() error {
	if cli.wallet == nil {
		cli.wallet = keystore.NewKeyStore(cli.walletPath,
			keystore.LightScryptN, keystore.LightScryptP)
		if len(cli.wallet.Accounts()) == 0 {
			return fmt.Errorf("Empty wallet, create account first")
		}
	}

	return nil
}

func (cli *CLI) buildAccount(addressStr string) error {

	err := cli.buildWallet()
	if err != nil {
		return err
	}

	var address common.Address
	if !common.IsHexAddress(addressStr) {
		if cli.address == (common.Address{}) {
			return fmt.Errorf("Error: address(%s) invalid", addressStr)
		}
		address = cli.address
	} else {
		address = common.HexToAddress(addressStr)
		cli.address = address
	}
	cli.account, err = cli.wallet.Find(accounts.Account{Address: address})
	if err != nil {
		return fmt.Errorf("Error: Can not get the keystore file of address %s", address.String())
	}

	return nil
}

func (cli *CLI) getTransactOpts(address string, gasLimit uint64) (*bind.TransactOpts, error) {
	err := cli.buildAccount(address)
	if err != nil {
		return nil, err
	}

	var trials int
	//var walletPassword string
	var keyJSON []byte
	for trials = 0; trials <= 3; trials++ {
		keyJSON, err = cli.wallet.Export(cli.account, cli.walletPassword, cli.walletPassword)
		if err == nil {
			break
		}
		if trials >= 3 {
			return nil, fmt.Errorf("Error: Failed to unlock account %s (%v)", cli.account.Address.String(), err)

		}
		prompt := fmt.Sprintf("Unlocking account %s | Attempt %d/%d", cli.account.Address.String(), trials+1, 3)
		cli.walletPassword, _ = getPassPhrase(prompt, false)
	}

	cli.BuildClient()
	chainId, err := cli.client.ChainID(context.Background())
	if err != nil {
		fmt.Println("ChainID Error: ", err)
		return nil, err
	}

	json, err := ioutil.ReadAll(bytes.NewReader(keyJSON))
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(json, cli.walletPassword)
	if err != nil {
		return nil, err
	}
	keyAddr := crypto.PubkeyToAddress(key.PrivateKey.PublicKey)
	opts := &bind.TransactOpts{
		From: keyAddr,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, errors.New("not authorized to sign this account")
			}
			if tx.Gas() < gasLimit && tx.To() != nil {
				tx = types.NewTransaction(tx.Nonce(), *tx.To(), tx.Value(), gasLimit, tx.GasPrice(), tx.Data())
			}
			signer := types.NewEIP155Signer(chainId)
			signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key.PrivateKey)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}

	return opts, err
}

// Execute parses the command line and processes it.
func (cli *CLI) Execute() {
	cli.rootCmd.Execute()
}

// setup turns up the CLI environment, and gets called by Cobra before
// a command is executed.
func (cli *CLI) setup(cmd *cobra.Command, args []string) {
	err := setupConfig(cli)
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(os.Stderr, cmd.UsageString())
		os.Exit(1)
	}
}

func (cli *CLI) help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())

	os.Exit(-1)

}

// TestCommand test command
func (cli *CLI) TestCommand(command string) string {
	//cli.testing = true
	result := cli.Run(strings.Fields(command)...)
	//	cli.testing = false
	return result
}

// Run executes CLI with the given arguments. Used for testing. Not thread safe.
func (cli *CLI) Run(args ...string) string {
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()

	os.Stdout = w

	cli.rootCmd.SetArgs(args)
	cli.rootCmd.Execute()
	cli.buildRootCmd()

	w.Close()

	os.Stdout = oldStdout

	var stdOut bytes.Buffer
	io.Copy(&stdOut, r)
	return stdOut.String()
}

// Embeddable returns a CLI that you can embed into your own Go programs. This
// is not thread-safe.
func (cli *CLI) Embeddable() *CLI {

	return cli
}

// SetPassword SetPassword
func (cli *CLI) SetPassword(_passPhrase string) *CLI {
	cli.walletPassword = _passPhrase
	return cli
}
