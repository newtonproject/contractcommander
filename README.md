
## contractcommander 

`contractcommander` project contains the following:
* Deploy contract with the source path, contract name and constructor arguments
* Call functions with function arguments, transfer add view support

## QuickStart

### Install solc

Get the binary from https://github.com/ethereum/solidity/releases or refer to https://solidity.readthedocs.io/en/latest/installing-solidity.html.

### Download from releases

Binary archives are published at https://release.cloud.diynova.com/newton/ContractCommander/.

### Building the source

The `solc` is required when deployed from source code, use command `npm install -g solc` to install solc.

### Windows

install command:

```bash
git clone https://github.com/newtonproject/contractcommander.git && cd contractcommander && make install
```

run contractcommander:

```bash
%GOPATH%/bin/contractcommander.exe
```

### Linux or Mac

install:

```bash
git clone https://github.com/newtonproject/contractcommander.git && cd contractcommander && make install
```

run contractcommander:

```bash
$GOPATH/bin/contractcommander
```

## Usage

### Help

Use command `contractcommander help` to display the usage.

```bash
Usage:
  contractcommander  [flags]
  contractcommander [command]

Available Commands:
  account     Manage NewChain accounts
  call        Call functions with args type and value
  deploy      Deploy NewChain contract
  help        Help about any command
  init        Initialize config file
  version     Get version of contractcommander CLI
  view        Get info from the contract by function name and args

Flags:
  -c, --config path               The path to config file (default "./config.toml")
  -a, --contractAddress address   Contract address
  -f, --from address              the from address who pay gas
  -h, --help                      help for contractcommander
  -i, --rpcURL url                Geth json rpc or ipc url (default "https://rpc1.newchain.newtonproject.org")
  -w, --walletPath directory      Wallet storage directory (default "./wallet/")

Use "contractcommander [command] --help" for more information about a command.

```

### Use config.toml

You can use a configuration file to simplify the command line parameters.

One available configuration file `config.toml` is as follows:


```conf
contractaddress = "0xC4c21B165D6C30366079F07fb5408178699aD6b7"
from = "0x4Ba80F138543E75AbF788eB3fE2726425586b0fD"
rpcurl = "https://rpc1.newchain.newtonproject.org"
walletpath = "./wallet/"
```

### Initialize config file

```bash
# Initialize config file
$ contractcommander init
```

### Create account

```bash
# Create an account with faucet
contractcommander account new --faucet

# Create 10 accounts
contractcommander account new -n 10 --faucet
```

### Deploy contract

```bash
# Deploy the `SimpleToken.sol` token
# contract SimpleToken {
#  constructor(string _name, string _symbol, uint8 _decimals, uint256 _initialsupply) public {}
#}
# Deploy SimpleToken with the name "HelloToken", the symbol "HT", the decimals "18" and the totalSupply "1024"
contractcommander deploy --sol SimpleToken.sol --name SimpleToken HelloToken HT 18 1024000000000000000000

# Deploy the `SimpleVote.sol` token
# contract SimpleVote {
#}
# Deploy SimpleVote
contractcommander deploy --sol SimpleVote.sol --name SimpleVote
```

In order to deploy contract from abi and bin, you should file compiler source contract with `solc`:

```bash
solc --bin --abi --optimize -o out SimpleToken.sol
```

the bin and abi file will be saved to folder `out`, then deploy with contractcommander:

```bash
contractcommander deploy --bin out/SimpleToken.bin --abi out/SimpleToken.abi HelloToken HT 18 1024000000000000000000
```

### Execute function on the NewChain

```bash
# Transfer token to address
contractcommander call transfer address 0x4Ba80F138543E75AbF788eB3fE2726425586b0fD uint256 1

# Execute function and send NEW to the contract
# SimpleVote: becomeCandidate
contractcommander call becomeCandidate --value 100

# Execute function and send NEW to the contract
# SimpleVote: vote
contractcommander call vote address 0x4Ba80F138543E75AbF788eB3fE2726425586b0fD --value 3
```


### View function

```bash
# Get contract name
contractcommander view name

# Get contract name and specify the type of the output variable
contractcommander view name --out string

# Get balanceOf address
contractcommander view balanceOf address 0x4Ba80F138543E75AbF788eB3fE2726425586b0fD --out uint256

# Get the totalSupply
contractcommander view totalSupply --out uint256
```

`contractcommander call --view` is the alias of `contractcommander view`

```bash
# Get contract name
contractcommander call name --view --out string

# Get contract info
# function info() public view returns (string, string, uint8, address ) {
#    return (name, symbol, decimals, owner);
#  }
contractcommander call info --view --out string,string,uint8,address


#submitTrade(bytes32 tradeID, bytes32[2] memory userAddress, uint256[2] memory coinID, uint256[2] memory coinAmount,
#       uint256[2] memory feeCoinID, bytes32[2] memory feeCoinAddress, uint256[2] memory feeCoinAmount, 
#       bytes32[2] memory receivingAddress) {}
contractcommander call submitTrade bytes32 0000000000000000000000000000000000000000000000000000000000000001 bytes32[2] 0000000000000000000000000000000000000000000000000000000000000001,0000000000000000000000000000000000000000000000000000000000000002 uint256[2] 1,2 uint256[2] 1,1  uint256[2] 1,2 bytes32[2] 0000000000000000000000000000000000000000000000000000000000000001,0000000000000000000000000000000000000000000000000000000000000002 uint256[2] 1,1 bytes32[2] 1000000000000000000000000000000000000000000000000000000000000001,1000000000000000000000000000000000000000000000000000000000000002
```


## Types 

argType Example |
---|
bool|
bool[]|
bool[2]|
bool[2][]|
bool[][]|
bool[][2]|
bool[2][2]|
bool[2][][2]|
bool[2][2][2]|
bool[][][]|
bool[][2][]|
int8|
int16|
int32|
int64|
int256|
int8[]|
int8[2]|
int16[]|
int16[2]|
int32[]|
int32[2]|
int64[]|
int64[2]|
int256[]|
int256[2]|
uint8|
uint16|
uint32|
uint64|
uint256|
uint8[]|
uint8[2]|
uint16[]|
uint16[2]|
uint32[]|
uint32[2]|
uint64[]|
uint64[2]|
uint256[]|
uint256[2]|
bytes32|
bytes[]|
bytes[2]|
bytes32[]|
bytes32[2]|
string|
string[]|
string[2]|
address|
address[]|
address[2]|