package lib

import (
	"io/ioutil"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

//Execute transaction to mint new token into service account
func MintToken(configuration Configuration, runtimeenv string, recipient flow.Address, URL string, creatorName string, creatorAddress flow.Address, caption string, hash string, edition uint16) {

	// get config
	var node string
	var serviceAddressHex string
	var servicePrivKeyHex string
	var serviceSigAlgoHex string
	var NFTContractAddress string
	var VeoletContractAddress string
	if runtimeenv == "emulator" {
		node = configuration.Networks.Emulator.Host
		serviceAddressHex = configuration.Accounts.Emulator_account.Address
		servicePrivKeyHex = configuration.Accounts.Emulator_account.Keys
		serviceSigAlgoHex = "ECDSA_P256"
		NFTContractAddress = configuration.Contractaddresses.Emulator.NonFungibleToken
		VeoletContractAddress = configuration.Contractaddresses.Emulator.Veolet
	} else if runtimeenv == "testnet" {
		node = configuration.Networks.Testnet.Host
		serviceAddressHex = configuration.Accounts.Testnet_account.Address
		servicePrivKeyHex = configuration.Accounts.Testnet_account.Keys
		serviceSigAlgoHex = "ECDSA_P256"
		NFTContractAddress = configuration.Contractaddresses.Testnet.NonFungibleToken
		VeoletContractAddress = configuration.Contractaddresses.Testnet.Veolet
	}

	// creation timestamp
	var timestamp = uint64(time.Now().Unix())

	//Read script file
	transactioncode, err := ioutil.ReadFile("cadence/transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, NFTContractAddress, VeoletContractAddress, "", "")

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
	SendTransaction(node, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, transactioncode, arguments, false)
}
