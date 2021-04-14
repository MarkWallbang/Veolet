package main

import (
	"fmt"
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func main() {

	// [15]
	node := "127.0.0.1:3569"

	serviceAddressHex := "f8d6e0586b0a20c7"
	servicePrivKeyHex := "e010314865d947dac99d5e7c3d81b442ba1a4550ebc20ced1e2adb7dd6a053b1"
	serviceSigAlgoHex := "ECDSA_P256"

	// Create new account using the service account as signer/payer
	newAddress, newPubKey, newPrivKey := CreateNewAccount(node, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex)
	fmt.Println("Got new account address: " + newAddress)
	fmt.Println("With public key: " + newPubKey)
	fmt.Println("With private key: " + newPrivKey)

	//Setup Veolet wallet for the new created account
	setupcode, err := ioutil.ReadFile("cadence/transactions/SetupAccount.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	SendTransaction(node, newAddress, newPrivKey, serviceSigAlgoHex, setupcode, nil, false)

	//Execute transaction to mint new token into service account
	transactioncode, err := ioutil.ReadFile("cadence/transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	var serviceaddress = flow.HexToAddress("f8d6e0586b0a20c7")
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(serviceaddress))
	arguments = append(arguments, cadence.NewString("www.testurl.com"))
	arguments = append(arguments, cadence.NewString("Data Simon"))
	arguments = append(arguments, cadence.NewAddress(serviceaddress))
	arguments = append(arguments, cadence.NewUInt64(123))

	SendTransaction(node, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, transactioncode, arguments, false)

	// execute script to view tokens in given account
	script, err := ioutil.ReadFile("cadence/scripts/FetchCollection.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	//define arguments
	var scriptarguments []cadence.Value
	scriptarguments = append(scriptarguments, cadence.NewAddress(serviceaddress))

	result := ExecuteScript(node, script, scriptarguments)
	fmt.Print(result)
}
