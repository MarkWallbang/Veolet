package main

import (
	//"encoding/json"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"veolet/lib"

	//"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	//"github.com/onflow/flow-go-sdk/templates"
)

func main() {

	runtimeenv := flag.String("env", "emulator", "In which environment to run the program. emulator/testnet/mainnet")
	teststring := "testnet"
	runtimeenv = &teststring
	fmt.Println("Running program on " + *runtimeenv)

	// Read configurations
	file, _ := os.Open("flow.json")
	defer file.Close()
	byteFile, _ := ioutil.ReadAll(file)
	var configuration lib.FlowConfiguration
	json.Unmarshal(byteFile, &configuration)
	config := lib.GetConfig(configuration, *runtimeenv)

	var serviceaddress = flow.HexToAddress(config.Account.Address)
	contracts := lib.FetchContracts(*config, serviceaddress)
	for k, v := range contracts {
		fmt.Printf("Contract[%s] value[%s]\n", k, string(v))
	}

	/*
		// Deploy Veolet contract, and if in emulator mode, deploy NFT core contract
		if *runtimeenv == "emulator" {
			// Read NFT core contract code
			code, err := ioutil.ReadFile("cadence/contracts/NonFungibleToken.cdc")
			if err != nil {
				panic("Cannot read script file")
			}

			lib.DeployContract(*config, "emulator", templates.Contract{
				Name:   "NonFungibleToken",
				Source: string(code),
			})
		}
		code, err := ioutil.ReadFile("cadence/contracts/Veolet.cdc")
		if err != nil {
			panic("Cannot read script file")
		}
		// Replace address placeholder
		code = lib.ReplaceAddressPlaceholders(code, config.Contractaddresses.NonFungibleToken, "", "", "")

		lib.DeployContract(*config, *runtimeenv, templates.Contract{
			Name:   "Veolet",
			Source: string(code),
		})

		// Create new account using the service account as signer/payer
		newAddress, newPubKey, newPrivKey := lib.CreateNewAccount(*config, *runtimeenv)
		fmt.Println("Got new account address: " + newAddress)
		fmt.Println("With public key: " + newPubKey)
		fmt.Println("With private key: " + newPrivKey)

		// mint new token to service account wallet
		lib.MintToken(*config, *runtimeenv, serviceaddress, "test.com", "Creator Name", serviceaddress, "TestNFTT", "0784fb1h3", 1)

		// execute script to view tokens in given account
		script, err := ioutil.ReadFile("cadence/scripts/FetchCollection.cdc")
		if err != nil {
			panic("Cannot read script file")
		}

		// Replace placeholder address
		script = lib.ReplaceAddressPlaceholders(script, config.Contractaddresses.NonFungibleToken, "", "", "")

		//define arguments
		var scriptarguments []cadence.Value
		scriptarguments = append(scriptarguments, cadence.NewAddress(serviceaddress))

		result := lib.ExecuteScript(config.Network.Host, script, scriptarguments)
		fmt.Print("Token stash of serviceaddress: ")
		fmt.Println(result)

		//transfer tokens from account 1 to the newly created account
		lib.TransferToken(*config, *runtimeenv, config.Account.Address, config.Account.Keys, newAddress, 0)

		// Print the token stash of both accounts to verify that transfer worked
		result = lib.ExecuteScript(config.Network.Host, script, scriptarguments)
		fmt.Print("Token stash of serviceaddress: ")
		fmt.Println(result)

		scriptarguments[0] = cadence.NewAddress(flow.HexToAddress(newAddress))
		result = lib.ExecuteScript(config.Network.Host, script, scriptarguments)
		fmt.Print("Token stash of new address: ")
		fmt.Println(result)

		// finally use script to gather all fields of given NFT ID
		readscript, err := ioutil.ReadFile("cadence/scripts/ReadNFT.cdc")
		if err != nil {
			panic("Cannot read script file")
		}
		// Replace placeholder address
		readscript = lib.ReplaceAddressPlaceholders(readscript, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")
		var readarguments []cadence.Value
		readarguments = append(readarguments, cadence.NewAddress(flow.HexToAddress(newAddress)))
		readarguments = append(readarguments, cadence.NewUInt64(0))
		result = lib.ExecuteScript(config.Network.Host, readscript, readarguments)
		fmt.Print("Information of token 0: ")
		fmt.Println(result)*/
}
