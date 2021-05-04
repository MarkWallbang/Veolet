package veolet

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
)

func amain() {
	// Read Flow configurations
	config := GetConfig("testnet")
	targetaddress := flow.HexToAddress("022a8b9defc588b3")

	//var serviceaddress = flow.HexToAddress(config.Account.Address)

	res, _ := FetchCollectionNFTs(*config, targetaddress)
	//res := FetchNFT(*config, targetaddress, 36)
	//t := res.ToGoValue().([]interface{})

	/*for id, value := range t {
		fmt.Print(id, value)
	}
	fmt.Print(t)*/
	res_json := ConvertNFTsToMap(res)
	_ = res_json
	//fmt.Println(string(res_json))
	res_go := res.ToGoValue().([]interface{})
	//fmt.Print(res_go)

	for _, value := range res_go {
		for j, field := range value.([]interface{}) {
			fmt.Println(field)
			if j == 3 {
				bytes := field.([8]byte)
				fmt.Println(flow.BytesToAddress(bytes[:]))
			}
		}
	}

	//fmt.Print(err)

	//cap, used := FetchStorageCapacity(*config, targetaddress)
	//fmt.Println(cap, used)
}
