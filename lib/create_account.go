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

type Account struct {
	Address flow.Address
	Privkey crypto.PrivateKey
}

// cryptographically secure random number generator (CSPRNG)
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// [2]
func GenerateKeys(sigAlgoName string) (string, string) {
	seed, err := GenerateRandomBytes(64)
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

func CreateNewAccount(configuration Configuration) (flow.Address, crypto.PrivateKey) {
	sigAlgoName := "ECDSA_P256"
	hashAlgoName := "SHA3_256"
	pubKey, privKey := GenerateKeys("ECDSA_P256")

	node := configuration.Network.Host

	// [16]
	gasLimit := uint64(100)

	// [17]
	txID := CreateAccount(configuration, pubKey, sigAlgoName, hashAlgoName, nil, gasLimit)

	// [19]
	address := GetAddress(node, txID)
	fmt.Println("New Account Address: " + address.Hex())

	//Setup Veolet wallet for the new created account
	PrivKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, privKey)
	if err != nil {
		panic("Cant decode private key")
	}

	SetupAccount(configuration, Account{Address: address, Privkey: PrivKey})

	return address, PrivKey
}
func CreateAccount(config Configuration,
	publicKeyHex string,
	sigAlgoName string,
	hashAlgoName string,
	code *string,
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
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	serviceAddress, serviceAccountKey, serviceSigner := GetServiceAccount(config)

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
func GetAddress(node string, txIDHex string) flow.Address {
	ctx := context.Background()
	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	// [12]
	txID := flow.HexToID(txIDHex)

	var address flow.Address
	for {
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
	return address
}

func SetupAccount(config Configuration, targetAccount Account) {
	// Function to create a Veolet collection for given account
	{
		ctx := context.Background()
		c, err := client.New(config.Network.Host, grpc.WithInsecure())
		if err != nil {
			panic("Failed to establish connection with Access API")
		}
		defer c.Close()
		// Read transaction script
		setupcode, err := ioutil.ReadFile("cadence/transactions/SetupAccount.cdc")
		if err != nil {
			panic("Cannot read script file")
		}
		// Replace placeholder addresses
		setupcode = ReplaceAddressPlaceholders(setupcode, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")

		// Get service account
		serviceAddress, serviceKey, serviceSigner := GetServiceAccount(config)

		// Get target Account
		flowTargetAccount, err := c.GetAccountAtLatestBlock(ctx, targetAccount.Address)
		if err != nil {
			fmt.Println(err)
			panic("Could not get target account")
		}
		// Create target account signer
		targetSigner := crypto.NewInMemorySigner(targetAccount.Privkey, flowTargetAccount.Keys[0].HashAlgo)

		// Send transaction with targetAccount as authorizer/proposer and service account as payer
		latestBlock, err := c.GetLatestBlockHeader(ctx, true)
		if err != nil {
			panic("failed to fetch latest block")
		}
		tx := flow.NewTransaction().
			SetScript(setupcode).
			SetProposalKey(flowTargetAccount.Address, flowTargetAccount.Keys[0].Index, flowTargetAccount.Keys[0].SequenceNumber).
			SetPayer(serviceAddress).
			SetReferenceBlockID(latestBlock.ID).
			AddAuthorizer(flowTargetAccount.Address)

		// Sign payload and envelope
		err = tx.SignPayload(targetAccount.Address, flowTargetAccount.Keys[0].Index, targetSigner)
		if err != nil {
			panic("Could not sign payload")
		}

		err = tx.SignEnvelope(serviceAddress, serviceKey.Index, serviceSigner)
		if err != nil {
			panic("Could not sign envelope")
		}
		// Send transaction
		err = c.SendTransaction(ctx, *tx)
		if err != nil {
			fmt.Print(err)
			panic("Failed to submit transaction")
		}

		// Get transaction result
		result, err := c.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			fmt.Print(err)
			panic("Transaction failed")
		}

		for result.Status != flow.TransactionStatusSealed {
			time.Sleep(time.Second)
			result, err = c.GetTransactionResult(ctx, tx.ID())
			if err != nil {
				fmt.Print(err)
				panic("Transaction failed")
			}
		}
	}
}
