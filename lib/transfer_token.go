package lib

import (
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func TransferToken(configuration Configuration, runtimeenv string, senderHex string, senderPrivKeyHex string, recipientHex string, tokenID uint64) {
	// get config
	var node string
	var NFTContractAddress string
	var VeoletContractAddress string
	if runtimeenv == "emulator" {
		node = configuration.Networks.Emulator.Host
		NFTContractAddress = configuration.Contractaddresses.Emulator.NonFungibleToken
		VeoletContractAddress = configuration.Contractaddresses.Emulator.Veolet
	} else if runtimeenv == "testnet" {
		node = configuration.Networks.Testnet.Host
		NFTContractAddress = configuration.Contractaddresses.Testnet.NonFungibleToken
		VeoletContractAddress = configuration.Contractaddresses.Testnet.Veolet
	}
	sigAlgoName := "ECDSA_P256"

	//Read transaction code
	transactioncode, err := ioutil.ReadFile("cadence/transactions/Transfer.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, NFTContractAddress, VeoletContractAddress, "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(flow.HexToAddress(recipientHex)))
	arguments = append(arguments, cadence.NewUInt64(tokenID))

	// Send transaction
	SendTransaction(node, senderHex, senderPrivKeyHex, sigAlgoName, transactioncode, arguments, false)
}
