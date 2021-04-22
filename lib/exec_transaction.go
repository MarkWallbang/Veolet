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
func SendTransaction(node string, inputSignerAcctAddr string, inputSignerPrivateKey string, inputSignerSigner string, transactioncode []byte, arguments []cadence.Value, payerAddressHex string, payerPrivKeyHex string) *flow.TransactionResult {
	ctx := context.Background()
	flowClient, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("Could not connect to client")
	}

	sigAlgo := crypto.StringToSignatureAlgorithm(inputSignerSigner)
	privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, inputSignerPrivateKey)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	addr := flow.HexToAddress(inputSignerAcctAddr)
	acc, err := flowClient.GetAccount(context.Background(), addr)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)

	// Get creds of payer account
	payerPrivKey, err := crypto.DecodePrivateKeyHex(sigAlgo, payerPrivKeyHex)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}
	payerAddress := flow.HexToAddress(payerAddressHex)
	payerAcc, err := flowClient.GetAccount(context.Background(), payerAddress)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}
	payerAccountKey := payerAcc.Keys[0]
	payerSigner := crypto.NewInMemorySigner(payerPrivKey, payerAccountKey.HashAlgo)

	block, err := flowClient.GetLatestBlock(ctx, true)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}
	referenceBlockID := block.ID
	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(addr, accountKey.Index, accountKey.SequenceNumber).
		SetReferenceBlockID(referenceBlockID).
		SetPayer(payerAddress).
		AddAuthorizer(addr)

	if arguments != nil {
		for i := 0; i < len(arguments); i++ {
			err = tx.AddArgument(arguments[i])
			if err != nil {
				fmt.Println("err:", err.Error())
				panic(err)
			}
		}
	}

	err = tx.SignPayload(addr, accountKey.Index, signer)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	err = tx.SignEnvelope(payerAddress, payerAccountKey.Index, payerSigner)
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
