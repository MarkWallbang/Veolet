package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"veolet/lib"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func main() {

	// Read Flow configurations
	file, _ := os.Open("flow.json")
	defer file.Close()
	byteFile, _ := ioutil.ReadAll(file)
	var configuration lib.FlowConfiguration
	json.Unmarshal(byteFile, &configuration)
	config := lib.GetConfig(configuration, "testnet")
	targetaddress := flow.HexToAddress("022a8b9defc588b3")
	//var serviceaddress = flow.HexToAddress(config.Account.Address)

	res := lib.FetchCollection(*config, targetaddress)
	fmt.Print(res)
	//cap, used := lib.FetchStorageCapacity(*config, targetaddress)
	//fmt.Println(cap, used)
	result := lib.FetchNFT(*config, targetaddress, 36)
	realresult := result.(cadence.Optional).Value.(cadence.Optional).Value.(cadence.Resource).Fields
	real2 := result.ToGoValue()
	fmt.Print(realresult)
	fmt.Print(real2)
	//fmt.Print(lib.FetchBalance(*config, targetaddress))

	/*
		address, _ := lib.CreateNewAccount(*config)
		fmt.Print(address)

		res := lib.MintToken(*config, address, "https://veoletimages.s3.eu-central-1.amazonaws.com/susi.jpeg", "Susi", address, "Susis NFT", "0784fb1h3", 1)
		fmt.Print(res)*/
}
