package main

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/examples"
	"github.com/onflow/flow-go-sdk/test"
	"google.golang.org/grpc"
)

func SendTransaction(inputSignerAcctAddr string, inputSignerPrivateKey string, inputSignerSigner string, transactioncode []byte, arguments []cadence.Value) {
	ctx := context.Background()
	flowClient, err := client.New("127.0.0.1:3569", grpc.WithInsecure())
	examples.Handle(err)

	sigAlgo := crypto.StringToSignatureAlgorithm(inputSignerSigner)
	privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, inputSignerPrivateKey)
	examples.Handle(err)

	addr := flow.HexToAddress(inputSignerAcctAddr)
	acc, err := flowClient.GetAccount(context.Background(), addr)
	examples.Handle(err)

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)
	//addr, accountKey, signer
	//serviceAcctAddr, serviceAcctKey, serviceSigner

	message := test.GreetingGenerator().Random()
	greeting := cadence.NewString(message)

	referenceBlockID := examples.GetReferenceBlockId(flowClient)
	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(addr, accountKey.Index, accountKey.SequenceNumber).
		SetReferenceBlockID(referenceBlockID).
		SetPayer(addr).
		AddAuthorizer(addr)

	for i := 0; i < len(arguments); i++ {
		err = tx.AddArgument(arguments[i])
		examples.Handle(err)
	}

	fmt.Println("Sending transaction:")
	fmt.Println()
	fmt.Println("----------------")
	fmt.Println("Script:")
	fmt.Println(string(tx.Script))
	fmt.Println("Arguments:")
	fmt.Printf("greeting: %s\n", greeting)
	fmt.Println("----------------")
	fmt.Println()

	err = tx.SignEnvelope(addr, accountKey.Index, signer)
	examples.Handle(err)

	err = flowClient.SendTransaction(ctx, *tx)
	examples.Handle(err)

	_ = examples.WaitForSeal(ctx, flowClient, tx.ID())
}
