package lib

import (
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func TransferToken(node string, senderHex string, senderPrivKeyHex string, recipientHex string, tokenID uint64) {
	// get config
	sigAlgoName := "ECDSA_P256"

	//Read transaction code
	transactioncode, err := ioutil.ReadFile("cadence/transactions/Transfer.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(flow.HexToAddress(recipientHex)))
	arguments = append(arguments, cadence.NewUInt64(tokenID))

	// Send transaction
	SendTransaction(node, senderHex, senderPrivKeyHex, sigAlgoName, transactioncode, arguments, false)
}
