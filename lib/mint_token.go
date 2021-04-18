package lib

import (
	"io/ioutil"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func MintToken(configuration Configuration, runtimeenv string, recipient flow.Address, URL string, creatorName string, creatorAddress flow.Address) {

	// get config
	var node string
	var serviceAddressHex string
	var servicePrivKeyHex string
	var serviceSigAlgoHex string
	if runtimeenv == "emulator" {
		node = configuration.Networks.Emulator.Host
		serviceAddressHex = configuration.Accounts.Emulator_account.Address
		servicePrivKeyHex = configuration.Accounts.Emulator_account.Keys
		serviceSigAlgoHex = "ECDSA_P256"
	} else if runtimeenv == "testnet" {
		node = configuration.Networks.Testnet.Host
		serviceAddressHex = configuration.Accounts.Testnet_account.Address
		servicePrivKeyHex = configuration.Accounts.Testnet_account.Keys
		serviceSigAlgoHex = "ECDSA_P256"
	}

	// creation timestamp
	var timestamp = uint64(time.Now().Unix())

	//Execute transaction to mint new token into service account
	transactioncode, err := ioutil.ReadFile("cadence/transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(recipient))
	arguments = append(arguments, cadence.NewString(URL))
	arguments = append(arguments, cadence.NewString(creatorName))
	arguments = append(arguments, cadence.NewAddress(creatorAddress))
	arguments = append(arguments, cadence.NewUInt64(timestamp))

	// Send transaction
	SendTransaction(node, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, transactioncode, arguments, false)
}
