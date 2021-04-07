package cli

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

type BlockChain int

const (
	UnknownChain BlockChain = iota
	NewChain
	Ethereum
)

const (
	newchainPublicKey = "c829d38b9fc274c8cb13b239a2b473ec04363167a84f2b4d6666b286f78c92515228bb895ac3802285cde0bac18592efbaffeb1bc14e1da00139b7dbf5248375"
	ethereumPublicKey = "979b7fa28feeb35a4741660a16076f1943202cb72b6af70d327f053e248bab9ba81760f39d0701ef1d8f89cc1fbd2cacba0710a12cd5314d5e0c9021aa3637f9"
)

func (bc BlockChain) String() string {
	switch bc {
	case NewChain:
		return "NewChain"
	case Ethereum:
		return "Ethereum"
	}

	return "UnknownChain"
}

func getBlockChain() (BlockChain, error) {
	// Check NewChain
	b, err := hex.DecodeString(newchainPublicKey)
	if err != nil {
		return UnknownChain, err
	} else if len(b) != 64 {
		return UnknownChain, fmt.Errorf("wrong length, want %d hex chars\n", 128)
	}
	b = append([]byte{0x4}, b...)

	x, _ := elliptic.Unmarshal(crypto.S256(), b)
	if x != nil {
		// OK
		return NewChain, nil
	}

	// Check Ethereum
	be, err := hex.DecodeString(ethereumPublicKey)
	if err != nil {
		return UnknownChain, err
	} else if len(be) != 64 {
		return UnknownChain, fmt.Errorf("wrong length, want %d hex chars\n", 128)
	}
	be = append([]byte{0x4}, be...)

	xb, _ := elliptic.Unmarshal(crypto.S256(), be)
	if xb != nil {
		// OK
		return Ethereum, nil
	}

	return UnknownChain, nil
}

type Unit string

var (
	UnitETH = "ETH"
	UnitWEI = "WEI"

	// UnitList is array for Unit string
	// UnitList = []string{"Wei", "Ada", "Babbage", "Shannon", "Szabo", "Finney", "Ether", "Einstein", "Douglas", "Gwei"}
	UnitList []string

	// UnitString is for Unit string
	// UnitString = "Available unit: Wei, Ada, Babbage, Shannon, Szabo, Finney, Ether, Einstein, Douglas, Gwei"
	UnitString string
)

func InitUnit(bc BlockChain) {
	if bc == NewChain {
		UnitETH = "NEW"
		UnitWEI = "ISAAC"
	}

	UnitList = []string{UnitETH, UnitWEI}
	UnitString = fmt.Sprintf("Available unit: %s", strings.Join(UnitList, ","))
}

func (bc BlockChain) Init() {
	InitUnit(bc)
}
