package lib

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

// TODO: implement payer to be our service address all the time -> right now the proposer (user) also pays
func SendTransaction(config Configuration, proposer Account, transactioncode []byte, arguments []cadence.Value) *flow.TransactionResult {
	ctx := context.Background()
	flowClient, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("Could not connect to client")
	}

	acc, err := flowClient.GetAccount(ctx, proposer.Address)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(proposer.Privkey, accountKey.HashAlgo)

	// Get service account (as payer)
	serviceAddress, serviceKey, serviceSigner := GetServiceAccount(config)

	block, err := flowClient.GetLatestBlock(ctx, true)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(proposer.Address, accountKey.Index, accountKey.SequenceNumber).
		SetReferenceBlockID(block.ID).
		SetPayer(serviceAddress).
		AddAuthorizer(proposer.Address)

	if arguments != nil {
		for i := 0; i < len(arguments); i++ {
			err = tx.AddArgument(arguments[i])
			if err != nil {
				fmt.Println("err:", err.Error())
				panic(err)
			}
		}
	}

	err = tx.SignPayload(proposer.Address, accountKey.Index, signer)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	err = tx.SignEnvelope(serviceAddress, serviceKey.Index, serviceSigner)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	err = flowClient.SendTransaction(ctx, *tx)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	result, err := flowClient.GetTransactionResult(ctx, tx.ID())
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		fmt.Print(".")
		result, err = flowClient.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			fmt.Println("err:", err.Error())
			panic(err)
		}
	}

	return result
}
