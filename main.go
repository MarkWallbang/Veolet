package main

import (
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func main() {
	//pubKey, privKey := GenerateKeys("ECDSA_P256")
	//fmt.Println(pubKey)
	//fmt.Println(privKey)

	// [15]
	node := "127.0.0.1:3569"

	serviceAddressHex := "f8d6e0586b0a20c7"
	servicePrivKeyHex := "e010314865d947dac99d5e7c3d81b442ba1a4550ebc20ced1e2adb7dd6a053b1"
	serviceSigAlgoHex := "ECDSA_P256"
	/*
	   sigAlgoName := "ECDSA_P256"
	   hashAlgoName := "SHA3_256"

	   code,err := ioutil.ReadFile("HelloWorldContract.cdc")
	   if err != nil{
	   panic("failed to load Cadence script")
	   }



	   // [16]
	   gasLimit := uint64(100)

	   // [17]
	   txID := CreateAccount(node, pubKey, sigAlgoName, hashAlgoName, string(code), serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, gasLimit)

	   fmt.Println(txID)

	   // [18]
	   blockTime := 10 * time.Second
	   time.Sleep(blockTime)

	   // [19]
	   address := GetAddress(node, txID)
	   fmt.Println(address)
	*/
	script, err := ioutil.ReadFile("cadence/scripts/Tester.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	//var arguments = nil
	ExecuteScript(node, script, nil)

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
	/*var arguments []cadence.Value
	var serviceaddress = flow.HexToAddress("f8d6e0586b0a20c7")
	arguments = append(arguments, cadence.NewString("asdasdas"))
	arguments = append(arguments, cadence.NewString("www.testurl.com"))
	arguments = append(arguments, cadence.NewAddress(serviceaddress))*/

	// call in the script: recipient: Address, initMediaURL: String, initCreatorName: String, initCreatorAddress: Address, initCreatedDate: UInt64

	SendTransaction(serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, transactioncode, arguments)
}
