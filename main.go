package veolet

import "fmt"

func main() {
	config := GetConfig("testnet")
	servaddress, _, _ := GetServiceAccount(*config)

	contr := FetchContracts(*config, servaddress)

	fmt.Println(contr)
}
