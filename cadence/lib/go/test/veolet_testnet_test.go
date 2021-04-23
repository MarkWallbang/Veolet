package test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"
	"veolet/lib"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
)

/*
	This test file holds all automated tests for the testnet deplyoment. The deployment of the contract will not be tested (have a look at veolet_test.go).
	All functions of the contract are tested here. For an instruction on how to deploy the Veolet contract, have a look at the Readme.md of teh repository.
*/

// Function to read testnet configuration
func setup(t *testing.T) *lib.Configuration {
	file, _ := os.Open("../../../../flow.json")
	defer file.Close()
	byteFile, _ := ioutil.ReadAll(file)
	var configuration lib.FlowConfiguration
	json.Unmarshal(byteFile, &configuration)
	config := lib.GetConfig(configuration, "emulator")
	return config
}

func getServiceAccount(t *testing.T, config lib.Configuration) testaccount {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	defer c.Close()

	serviceAddress := flow.HexToAddress(config.Account.Address)
	serviceAccount, err := c.GetAccountAtLatestBlock(ctx, serviceAddress)
	if err != nil {
		t.Error("Failed to get Veolet testnet account")
	}
	servicePrivKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, config.Account.Keys)
	if err != nil {
		t.Error("Failed to decode Veolet private key")
	}
	serviceSigner := crypto.NewInMemorySigner(servicePrivKey, serviceAccount.Keys[0].HashAlgo)
	return testaccount{address: serviceAddress, key: serviceAccount.Keys[0], signer: serviceSigner}
}

// Function to create a new account on the testnet using the Veolet service account as the payer/creator
func createAccountTestnet(t *testing.T, config lib.Configuration) testaccount {
	ctx := context.Background()

	accountKey, signer := createAccountCreds(t)
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	defer c.Close()

	//Get service account
	serviceAccount := getServiceAccount(t, config)
	// Use the templates package to create a new account creation transaction
	tx := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, serviceAccount.address)

	// Set the transaction payer and proposal key
	tx.SetPayer(serviceAccount.address)
	tx.SetProposalKey(
		serviceAccount.address,
		serviceAccount.key.Index,
		serviceAccount.key.SequenceNumber,
	)

	// Get the latest sealed block to use as a reference block
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		t.Error("failed to fetch latest block")
	}
	tx.SetReferenceBlockID(latestBlock.ID)

	// Sign and submit the transaction
	err = tx.SignEnvelope(serviceAccount.address, serviceAccount.key.Index, serviceAccount.signer)
	if err != nil {
		panic("failed to sign transaction envelope")
	}

	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		panic("failed to send transaction to network")
	}

	var address flow.Address
	for true {
		result, err := c.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			t.Error("Failed to get transaction result")
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
	serviceAccount.key.SequenceNumber++
	return testaccount{address: address, key: accountKey, signer: signer}
}

// Function to create a Veolet collection for given account
func setupAccountTestnet(t *testing.T, config lib.Configuration, targetAccount testaccount) {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}

	// Read transaction script
	setupcode, err := ioutil.ReadFile("../../../transactions/SetupAccount.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	// Replace placeholder addresses
	setupcode = lib.ReplaceAddressPlaceholders(setupcode, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")

	// Get service account
	serviceAccount := getServiceAccount(t, config)

	// Send transaction with targetAccount as authorizer/proposer and service account as payer
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		t.Error("failed to fetch latest block")
	}
	tx := flow.NewTransaction().
		SetScript(setupcode).
		SetProposalKey(targetAccount.address, targetAccount.key.Index, targetAccount.key.SequenceNumber).
		SetPayer(serviceAccount.address).
		SetReferenceBlockID(latestBlock.ID).
		AddAuthorizer(targetAccount.address)

	// Sign payload and envelope
	err = tx.SignPayload(targetAccount.address, targetAccount.key.Index, targetAccount.signer)
	if err != nil {
		t.Error("Could not sign payload")
	}

	err = tx.SignEnvelope(serviceAccount.address, serviceAccount.key.Index, serviceAccount.signer)
	if err != nil {
		t.Error("Could not sign envelope")
	}
}

func TestCreateAndSetupAccountTestnet(t *testing.T) {
	// Should be able to create a new account and set it up to hold a Veolet collection
	// TODO: The created account information should be saved to a file
	t.Log("Start CreateAndSetupAccountTestnet test")

	// Read config
	config := setup(t)

	// Create new account on testnet using Veolet service account
	newAccount := createAccountTestnet(t, *config)

	// Setup Veolet collection for the new Flow account
	setupAccountTestnet(t, *config, newAccount)

}

func TestMintTokenTestnet(t *testing.T) {
	// Should be able to mint new token
}

func TestTransferTokenTestnet(t *testing.T) {
	// Should be able to send tokens
}

func TestGetVeoletInformationTestnet(t *testing.T) {
	// Should be able to get information about collection
}

func TestSetMediaURLTestnet(t *testing.T) {
	// Should be able to get all fields of specific Veolet NFT
}
