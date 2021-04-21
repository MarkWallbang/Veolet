package lib

import (
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func TransferToken(configuration Configuration, runtimeenv string, senderHex string, senderPrivKeyHex string, recipientHex string, tokenID uint64) *flow.TransactionResult {

	// constants
	sigAlgoName := "ECDSA_P256"

	//Read transaction code
	transactioncode, err := ioutil.ReadFile("cadence/transactions/Transfer.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, configuration.Contractaddresses.NonFungibleToken, configuration.Contractaddresses.Veolet, "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(flow.HexToAddress(recipientHex)))
	arguments = append(arguments, cadence.NewUInt64(tokenID))

	// Send transaction
	result := SendTransaction(configuration.Network.Host, senderHex, senderPrivKeyHex, sigAlgoName, transactioncode, arguments, false)
	return result
}
