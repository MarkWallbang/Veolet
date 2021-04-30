package lib

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func FetchContracts(config Configuration, address flow.Address) map[string][]byte {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}
	account, err := c.GetAccount(ctx, address)

	if err != nil {
		panic("failed to fetch account")
	}
	return account.Contracts
}

func FetchStorageCapacity(config Configuration, address flow.Address) (int, int) {
	// TODO implement method to check if user needs more capacity

	//Read script file
	code, err := ioutil.ReadFile("cadence/scripts/StorageUsed.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(address))

	result := ExecuteScript(config.Network.Host, code, arguments)
	resultarr := result.(cadence.Array).Values
	return int(resultarr[0].(cadence.UInt64)), int(resultarr[1].(cadence.UInt64))
}

func FetchCollection(config Configuration, target flow.Address) cadence.Value {
	// Function to fetch the token ID's of target account
	// Read script file
	fetchscript, err := ioutil.ReadFile("cadence/scripts/FetchCollection.cdc")
	if err != nil {
		panic("Could not read script file")
	}
	fetchscript = ReplaceAddressPlaceholders(fetchscript, config.Contractaddresses.NonFungibleToken, "", "", "")
	result := ExecuteScript(config.Network.Host, fetchscript, []cadence.Value{cadence.NewAddress(target)})
	if err != nil {
		fmt.Print(err)
		panic("Could not execute script")
	}
	return result
}
func FetchBalance(config Configuration, target flow.Address) uint64 {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("Failed to establish connection with Access API")
	}
	defer c.Close()

	flowaccount, err := c.GetAccount(ctx, target)
	if err != nil {
		panic("Could not get Account")
	}
	balance := flowaccount.Balance
	return balance
}
