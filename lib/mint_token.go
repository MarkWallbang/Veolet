package lib

import (
	"io/ioutil"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

//Execute transaction to mint new token into service account
func MintToken(configuration Configuration, runtimeenv string, recipient flow.Address, URL string, creatorName string, creatorAddress flow.Address, caption string, hash string, edition uint16) *flow.TransactionResult {

	// constants
	serviceSigAlgoHex := "ECDSA_P256"

	// creation timestamp
	var timestamp = uint64(time.Now().Unix())

	//Read script file
	transactioncode, err := ioutil.ReadFile("cadence/transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, configuration.Contractaddresses.NonFungibleToken, configuration.Contractaddresses.Veolet, "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(recipient))
	arguments = append(arguments, cadence.NewString(URL))
	arguments = append(arguments, cadence.NewString(creatorName))
	arguments = append(arguments, cadence.NewAddress(creatorAddress))
	arguments = append(arguments, cadence.NewUInt64(timestamp))
	arguments = append(arguments, cadence.NewString(caption))
	arguments = append(arguments, cadence.NewString(hash))
	arguments = append(arguments, cadence.NewUInt16(edition))

	// Send transaction
	result := SendTransaction(configuration.Network.Host, configuration.Account.Address, configuration.Account.Keys, serviceSigAlgoHex, transactioncode, arguments, false)
	return result
}
