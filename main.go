package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"veolet/lib"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func main() {

	runtimeenv := flag.String("env", "emulator", "In which environment to run the program. emulator/testnet/mainnet")
	fmt.Println("Running program on " + *runtimeenv)

	// Read configurations
	// TODO need to adapt based on enviroment
	// -> then also the import adresses of the scripts and transactions need to be adapted
	file, _ := os.Open("flow.json")
	defer file.Close()
	byteFile, _ := ioutil.ReadAll(file)
	var configuration lib.Configuration
	json.Unmarshal(byteFile, &configuration)

	var node string
	var serviceAddressHex string
	var servicePrivKeyHex string
	var serviceSigAlgoHex string
	if *runtimeenv == "emulator" {
		node = configuration.Networks.Emulator.Host
		serviceAddressHex = configuration.Accounts.Emulator_account.Address
		servicePrivKeyHex = configuration.Accounts.Emulator_account.Keys
		serviceSigAlgoHex = "ECDSA_P256"
	} else if *runtimeenv == "testnet" {
		node = configuration.Networks.Testnet.Host
		serviceAddressHex = configuration.Accounts.Testnet_account.Address
		servicePrivKeyHex = configuration.Accounts.Testnet_account.Keys
		serviceSigAlgoHex = "ECDSA_P256"
	}

	var serviceaddress = flow.HexToAddress(serviceAddressHex)

	// Create new account using the service account as signer/payer
	newAddress, newPubKey, newPrivKey := lib.CreateNewAccount(node, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex)
	fmt.Println("Got new account address: " + newAddress)
	fmt.Println("With public key: " + newPubKey)
	fmt.Println("With private key: " + newPrivKey)

	// mint new token to service account wallet
	lib.MintToken(configuration, *runtimeenv, serviceaddress, "test.com", "Creator Name", serviceaddress)

	// execute script to view tokens in given account
	script, err := ioutil.ReadFile("cadence/scripts/FetchCollection.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	//define arguments
	var scriptarguments []cadence.Value
	scriptarguments = append(scriptarguments, cadence.NewAddress(serviceaddress))

	result := lib.ExecuteScript(node, script, scriptarguments)
	fmt.Print("Token stash of serviceaddress: ")
	fmt.Println(result)

	//transfer tokens from account 1 to the newly created account
	lib.TransferToken(node, serviceAddressHex, servicePrivKeyHex, newAddress, 0)

	// Print the token stash of both accounts to verify that transfer worked
	result = lib.ExecuteScript(node, script, scriptarguments)
	fmt.Print("Token stash of serviceaddress: ")
	fmt.Println(result)

	scriptarguments[0] = cadence.NewAddress(flow.HexToAddress(newAddress))
	result = lib.ExecuteScript(node, script, scriptarguments)
	fmt.Print("Token stash of new address: ")
	fmt.Println(result)

}
