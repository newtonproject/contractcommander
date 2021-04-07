package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	prompt0 "github.com/ethereum/go-ethereum/console/prompt"
)

var (
	big10        = big.NewInt(10)
	big1NEWInWEI = new(big.Int).Exp(big10, big.NewInt(18), nil)

	errClientNil           = errors.New("Failed to connect to the NewChain client")
	errCliNil              = errors.New("Cli error")
	errCliTranNil          = errors.New("Cli tran error")
	errBigSetString        = errors.New("conver string to big error")
	errLessThan0Wei        = errors.New("The transaction amount is less than 0 WEI")
	errIllegalAmount       = errors.New("Illegal Amount")
	errIllegalUnit         = errors.New("Illegal Unit")
	errRequiredFromAddress = errors.New(`required flag(s) "from" not set`)
)

var IsDecimalString = regexp.MustCompile(`^[1-9]\d*$|^0$|^0\.\d*$|^[1-9](\d)*\.(\d)*$`).MatchString

func showSuccess(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

// getPassPhrase retrieves the password associated with an account,
// requested interactively from the user.
func getPassPhrase(prompt string, confirmation bool) (string, error) {
	// prompt the user for the password
	if prompt != "" {
		fmt.Println(prompt)
	}
	password, err := prompt0.Stdin.PromptPassword("Enter passphrase (empty for no passphrase): ")
	if err != nil {
		return "", err
	}
	if confirmation {
		confirm, err := prompt0.Stdin.PromptPassword("Enter same passphrase again: ")
		if err != nil {
			return "", err
		}
		if password != confirm {
			return "", fmt.Errorf("Passphrases do not match")
		}
	}
	return password, nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func getAmountWei(amountStr, unit string) (*big.Int, error) {
	if amountStr == "" {
		return big.NewInt(0), nil
	}
	switch unit {
	case UnitETH:
		index := strings.IndexByte(amountStr, '.')
		if index <= 0 {
			amountWei, ok := new(big.Int).SetString(amountStr, 10)
			if !ok {
				return nil, errBigSetString
			}
			return new(big.Int).Mul(amountWei, big1NEWInWEI), nil
		}
		amountStrInt := amountStr[:index]
		amountStrDec := amountStr[index+1:]
		amountStrDecLen := len(amountStrDec)
		if amountStrDecLen > 18 {
			return nil, errIllegalAmount
		}
		amountStrInt = amountStrInt + strings.Repeat("0", 18)
		amountStrDec = amountStrDec + strings.Repeat("0", 18-amountStrDecLen)

		amountStrIntBig, ok := new(big.Int).SetString(amountStrInt, 10)
		if !ok {
			return nil, errBigSetString
		}
		amountStrDecBig, ok := new(big.Int).SetString(amountStrDec, 10)
		if !ok {
			return nil, errBigSetString
		}

		return new(big.Int).Add(amountStrIntBig, amountStrDecBig), nil
	case UnitWEI:
		amountWei, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			return nil, errBigSetString
		}
		return amountWei, nil
	}

	return nil, errIllegalUnit
}

func getWeiAmountTextUnitByUnit(amount *big.Int, unit string) string {
	if amount == nil {
		return fmt.Sprintf("0 %v", UnitWEI)
	}
	amountStr := amount.String()
	amountStrLen := len(amountStr)
	if unit == "" {
		if amountStrLen <= 18 {
			// show in WEI
			unit = UnitWEI
		} else {
			unit = UnitETH
		}
	}

	return fmt.Sprintf("%s %s", getWeiAmountTextByUnit(amount, unit), unit)
}

func getWeiAmountTextByUnit(amount *big.Int, unit string) string {
	if amount == nil {
		return "0"
	}
	amountStr := amount.String()
	amountStrLen := len(amountStr)

	switch unit {
	case UnitETH:
		var amountStrDec, amountStrInt string
		if amountStrLen <= 18 {
			amountStrDec = strings.Repeat("0", 18-amountStrLen) + amountStr
			amountStrInt = "0"
		} else {
			amountStrDec = amountStr[amountStrLen-18:]
			amountStrInt = amountStr[:amountStrLen-18]
		}
		amountStrDec = strings.TrimRight(amountStrDec, "0")
		if len(amountStrDec) <= 0 {
			return amountStrInt
		}
		return amountStrInt + "." + amountStrDec

	case UnitWEI:
		return amountStr
	}

	return errIllegalUnit.Error()
}

func createNewAccount(walletPath string, numOfNew int) error {

	wallet := keystore.NewKeyStore(walletPath,
		keystore.LightScryptN, keystore.LightScryptP)

	walletPassword, err := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	for i := 0; i < numOfNew; i++ {
		account, err := wallet.NewAccount(walletPassword)
		if err != nil {
			fmt.Println("Account error:", err)
			return err
		}
		fmt.Println(account.Address.Hex())
	}

	return nil
}

func showTransactionReceipt(url, txStr string) {
	var jsonStr = []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"],"id":1}`, txStr))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	clientHttp := &http.Client{}

	resp, err := clientHttp.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var body json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			fmt.Println(err)
			return
		}

		bodyStr, err := json.MarshalIndent(body, "", "    ")
		if err != nil {
			fmt.Println("JSON marshaling failed: ", err)
			return
		}
		fmt.Printf("%s\n", bodyStr)

		return
	}
}

func getFaucet(faucet, address string) {
	url := fmt.Sprintf("%s/faucet?address=%s", faucet, address)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Get error: %v\n", err)
		return
	}
	if resp.StatusCode == 200 {
		fmt.Printf("Get faucet for %s\n", address)
	}
}
