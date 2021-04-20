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
	"github.com/onflow/flow-go-sdk/templates"
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
	var NFTContractAddress string
	var VeoletContractAddress string
	//	var serviceSigAlgoHex string
	if *runtimeenv == "emulator" {
		node = configuration.Networks.Emulator.Host
		serviceAddressHex = configuration.Accounts.Emulator_account.Address
		servicePrivKeyHex = configuration.Accounts.Emulator_account.Keys
		NFTContractAddress = configuration.Contractaddresses.Emulator.NonFungibleToken
		VeoletContractAddress = configuration.Contractaddresses.Emulator.Veolet
		//serviceSigAlgoHex = "ECDSA_P256"
	} else if *runtimeenv == "testnet" {
		node = configuration.Networks.Testnet.Host
		serviceAddressHex = configuration.Accounts.Testnet_account.Address
		servicePrivKeyHex = configuration.Accounts.Testnet_account.Keys
		NFTContractAddress = configuration.Contractaddresses.Testnet.NonFungibleToken
		VeoletContractAddress = configuration.Contractaddresses.Testnet.Veolet
		//serviceSigAlgoHex = "ECDSA_P256"
	}

	var serviceaddress = flow.HexToAddress(serviceAddressHex)

	// Deploy Veolet contract, and if in emulator mode, deploy NFT core contract
	if *runtimeenv == "emulator" {
		// Read NFT core contract code
		code, err := ioutil.ReadFile("cadence/contracts/NonFungibleToken.cdc")
		if err != nil {
			panic("Cannot read script file")
		}

		lib.DeployContract(configuration, "emulator", templates.Contract{
			Name:   "NonFungibleToken",
			Source: string(code),
		})
	}
	code, err := ioutil.ReadFile("cadence/contracts/Veolet.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	// Replace address placeholder
	code = lib.ReplaceAddressPlaceholders(code, NFTContractAddress, "", "", "")

	lib.DeployContract(configuration, *runtimeenv, templates.Contract{
		Name:   "Veolet",
		Source: string(code),
	})

	// Create new account using the service account as signer/payer
	newAddress, newPubKey, newPrivKey := lib.CreateNewAccount(configuration, *runtimeenv)
	fmt.Println("Got new account address: " + newAddress)
	fmt.Println("With public key: " + newPubKey)
	fmt.Println("With private key: " + newPrivKey)

	// mint new token to service account wallet
	lib.MintToken(configuration, *runtimeenv, serviceaddress, "test.com", "Creator Name", serviceaddress, "TestNFTT", "0784fb1h3", 1)

	// execute script to view tokens in given account
	script, err := ioutil.ReadFile("cadence/scripts/FetchCollection.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Replace placeholder address
	script = lib.ReplaceAddressPlaceholders(script, NFTContractAddress, "", "", "")

	//define arguments
	var scriptarguments []cadence.Value
	scriptarguments = append(scriptarguments, cadence.NewAddress(serviceaddress))

	result := lib.ExecuteScript(node, script, scriptarguments)
	fmt.Print("Token stash of serviceaddress: ")
	fmt.Println(result)

	//transfer tokens from account 1 to the newly created account
	lib.TransferToken(configuration, *runtimeenv, serviceAddressHex, servicePrivKeyHex, newAddress, 0)

	// Print the token stash of both accounts to verify that transfer worked
	result = lib.ExecuteScript(node, script, scriptarguments)
	fmt.Print("Token stash of serviceaddress: ")
	fmt.Println(result)

	scriptarguments[0] = cadence.NewAddress(flow.HexToAddress(newAddress))
	result = lib.ExecuteScript(node, script, scriptarguments)
	fmt.Print("Token stash of new address: ")
	fmt.Println(result)

	// finally use script to gather all fields of given NFT ID
	readscript, err := ioutil.ReadFile("cadence/scripts/ReadNFT.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	// Replace placeholder address
	readscript = lib.ReplaceAddressPlaceholders(readscript, NFTContractAddress, VeoletContractAddress, "", "")
	var readarguments []cadence.Value
	readarguments = append(readarguments, cadence.NewAddress(flow.HexToAddress(newAddress)))
	readarguments = append(readarguments, cadence.NewUInt64(0))
	result = lib.ExecuteScript(node, readscript, readarguments)
	fmt.Print("Information of token 0: ")
	fmt.Println(result)
}
