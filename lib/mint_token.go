package lib

import (
	"io/ioutil"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func MintToken(configuration Configuration, recipient flow.Address, URL string, creatorName string, creatorAddress flow.Address) {

	// get config
	node := configuration.Node
	serviceAddressHex := configuration.ServiceAddressHex
	servicePrivKeyHex := configuration.ServicePrivKeyHex
	serviceSigAlgoHex := configuration.ServiceSigAlgoHex

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
