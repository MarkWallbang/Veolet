package main

// [1]
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
)

// [2]
func GenerateKeys(sigAlgoName string) (string, string) {
	seed := make([]byte, crypto.MinSeedLength)
	_, err := rand.Read(seed)
	if err != nil {
		panic(err)
	}

	// [3]
	sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoName)
	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		panic(err)
	}

	// [4]
	publicKey := privateKey.PublicKey()

	pubKeyHex := hex.EncodeToString(publicKey.Encode())
	privKeyHex := hex.EncodeToString(privateKey.Encode())

	return pubKeyHex, privKeyHex
}

func CreateNewAccount(node string, serviceAddressHex string, servicePrivKeyHex string, serviceSigAlgoHex string) string {
	sigAlgoName := "ECDSA_P256"
	hashAlgoName := "SHA3_256"
	pubKey, privKey := GenerateKeys("ECDSA_P256")
	fmt.Println("Generated new KeyPair:")
	fmt.Println("Public Key: " + pubKey)
	fmt.Println("Private Key: " + privKey)

	// [16]
	gasLimit := uint64(100)

	// [17]
	txID := CreateAccount(node, pubKey, sigAlgoName, hashAlgoName, nil, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, gasLimit) // statt nil -> string(code)

	fmt.Println("Transaction ID: " + txID)

	// [18]
	blockTime := 10 * time.Second
	time.Sleep(blockTime)

	// [19]
	address := GetAddress(node, txID)
	fmt.Println("New Account Address: " + address)

	return address
}
func CreateAccount(node string,
	publicKeyHex string,
	sigAlgoName string,
	hashAlgoName string,
	code *string,
	serviceAddressHex string,
	servicePrivKeyHex string,
	serviceSigAlgoName string,
	gasLimit uint64) string {

	ctx := context.Background()

	sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoName)
	publicKey, err := crypto.DecodePublicKeyHex(sigAlgo, publicKeyHex)
	if err != nil {
		panic(err)
	}

	hashAlgo := crypto.StringToHashAlgorithm(hashAlgoName)

	// [4]
	accountKey := flow.NewAccountKey().
		SetPublicKey(publicKey).
		SetSigAlgo(sigAlgo).
		SetHashAlgo(hashAlgo).
		SetWeight(flow.AccountKeyWeightThreshold)

	// [6]
	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	serviceSigAlgo := crypto.StringToSignatureAlgorithm(serviceSigAlgoName)
	servicePrivKey, err := crypto.DecodePrivateKeyHex(serviceSigAlgo, servicePrivKeyHex)
	if err != nil {
		panic(err)
	}

	serviceAddress := flow.HexToAddress(serviceAddressHex)
	serviceAccount, err := c.GetAccountAtLatestBlock(ctx, serviceAddress)
	if err != nil {
		panic(err)
	}

	// [7]
	serviceAccountKey := serviceAccount.Keys[0]
	serviceSigner := crypto.NewInMemorySigner(servicePrivKey, serviceAccountKey.HashAlgo)

	// [8] für contract statt nil -> []templates.Contract{{
	//Name:   "HelloWorld",
	//Source: code,
	//}}
	tx := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)
	tx.SetProposalKey(serviceAddress, serviceAccountKey.Index, serviceAccountKey.SequenceNumber)
	tx.SetPayer(serviceAddress)
	tx.SetGasLimit(uint64(gasLimit))

	// Get the latest sealed block to use as a reference block
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		panic("failed to fetch latest block")
	}

	tx.SetReferenceBlockID(latestBlock.ID)

	err = tx.SignEnvelope(serviceAddress, serviceAccountKey.Index, serviceSigner)
	if err != nil {
		panic(err)
	}

	// [9]
	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		panic(err)
	}

	// [10]
	return tx.ID().String()
}

// [11]
func GetAddress(node string, txIDHex string) string {
	ctx := context.Background()
	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	// [12]
	txID := flow.HexToID(txIDHex)
	result, err := c.GetTransactionResult(ctx, txID)
	if err != nil {
		panic("failed to get transaction result")
	}

	// [13]
	var address flow.Address

	if result.Status == flow.TransactionStatusSealed {
		for _, event := range result.Events {
			if event.Type == flow.EventAccountCreated {
				accountCreatedEvent := flow.AccountCreatedEvent(event)
				address = accountCreatedEvent.Address()
			}
		}
	}

	// [14]
	return address.Hex()
}
