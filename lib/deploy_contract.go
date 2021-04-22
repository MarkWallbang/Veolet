package lib

import (
	"io/ioutil"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
)

func DeployContract(configuration Configuration, runtimeenv string, contract templates.Contract) *flow.TransactionResult {

	code, err := ioutil.ReadFile("cadence/transactions/AddContract.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewString(contract.Name))
	arguments = append(arguments, cadence.NewString(contract.SourceHex()))

	result := SendTransaction(configuration.Network.Host, configuration.Account.Address, configuration.Account.Keys, "ECDSA_P256", code, arguments, configuration.Account.Address, configuration.Account.Keys)
	return result
}
