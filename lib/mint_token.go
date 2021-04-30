package lib

import (
	"io/ioutil"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

//Execute transaction to mint new token into service account
func MintToken(configuration Configuration, recipient flow.Address, URL string, creatorName string, creatorAddress flow.Address, caption string, hash string, edition uint16) *flow.TransactionResult {

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

	// Get service account (as payer)
	serviceAddress := flow.HexToAddress(configuration.Account.Address)
	servicePrivKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, configuration.Account.Keys)
	if err != nil {
		panic("Could not decode priv key.")
	}

	// Send transaction
	result := SendTransaction(configuration, Account{Address: serviceAddress, Privkey: servicePrivKey}, transactioncode, arguments)
	return result
}
