package lib

import (
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/templates"
)

func DeployContract(configuration Configuration, runtimeenv string, contract templates.Contract) {

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

	code, err := ioutil.ReadFile("cadence/transactions/AddContract.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewString(contract.Name))
	arguments = append(arguments, cadence.NewString(contract.SourceHex()))

	SendTransaction(node, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, code, arguments, false)

}
