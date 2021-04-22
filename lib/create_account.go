package lib

// [1]
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

func CreateNewAccount(configuration Configuration, runtimeenv string) (string, string, string) {
	sigAlgoName := "ECDSA_P256"
	hashAlgoName := "SHA3_256"
	pubKey, privKey := GenerateKeys("ECDSA_P256")

	node := configuration.Network.Host
	serviceAddressHex := configuration.Account.Address
	servicePrivKeyHex := configuration.Account.Keys
	serviceSigAlgoHex := "ECDSA_P256"
	NFTContractAddress := configuration.Contractaddresses.NonFungibleToken
	VeoletContractAddress := configuration.Contractaddresses.Veolet

	// [16]
	gasLimit := uint64(100)

	// [17]
	txID := CreateAccount(node, pubKey, sigAlgoName, hashAlgoName, nil, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, gasLimit) // statt nil -> string(code)

	// [19]
	address := GetAddress(node, txID)
	fmt.Println("New Account Address: " + address)

	//Setup Veolet wallet for the new created account
	// 1. Read transaction script
	setupcode, err := ioutil.ReadFile("cadence/transactions/SetupAccount.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	// 2. Replace placeholder addresses
	setupcode = ReplaceAddressPlaceholders(setupcode, NFTContractAddress, VeoletContractAddress, "", "")

	result := SendTransaction(node, address, privKey, serviceSigAlgoHex, setupcode, nil, address, privKey)
	if result.Error != nil {
		panic("Setup account failed")
	}

	return address, pubKey, privKey
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

	var address flow.Address
	for true {
		result, err := c.GetTransactionResult(ctx, txID)
		if err != nil {
			panic("failed to get transaction result")
		}
		if result.Status == flow.TransactionStatusSealed {
			for _, event := range result.Events {
				if event.Type == flow.EventAccountCreated {
					accountCreatedEvent := flow.AccountCreatedEvent(event)
					address = accountCreatedEvent.Address()
				}
			}
			break
		} else {
			time.Sleep(time.Second)
		}
	}

	// [14]
	return address.Hex()
}
