package lib

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/examples"
	"google.golang.org/grpc"
)

// TODO: implement payer to be our service address all the time -> right now the proposer (user) also pays
func SendTransaction(node string, inputSignerAcctAddr string, inputSignerPrivateKey string, inputSignerSigner string, transactioncode []byte, arguments []cadence.Value, debug bool) *flow.TransactionResult {
	ctx := context.Background()
	flowClient, err := client.New(node, grpc.WithInsecure())
	examples.Handle(err)

	sigAlgo := crypto.StringToSignatureAlgorithm(inputSignerSigner)
	privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, inputSignerPrivateKey)
	examples.Handle(err)

	addr := flow.HexToAddress(inputSignerAcctAddr)
	acc, err := flowClient.GetAccount(context.Background(), addr)
	examples.Handle(err)

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)

	referenceBlockID := examples.GetReferenceBlockId(flowClient)
	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(addr, accountKey.Index, accountKey.SequenceNumber).
		SetReferenceBlockID(referenceBlockID).
		SetPayer(addr).
		AddAuthorizer(addr)

	if arguments != nil {
		for i := 0; i < len(arguments); i++ {
			err = tx.AddArgument(arguments[i])
			examples.Handle(err)
		}
	}

	err = tx.SignEnvelope(addr, accountKey.Index, signer)
	examples.Handle(err)

	err = flowClient.SendTransaction(ctx, *tx)
	examples.Handle(err)

	result, err := flowClient.GetTransactionResult(ctx, tx.ID())
	examples.Handle(err)

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		fmt.Print(".")
		result, err = flowClient.GetTransactionResult(ctx, tx.ID())
		examples.Handle(err)
	}

	return result
}
