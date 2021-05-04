package veolet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/onflow/cadence"
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

type testaccountjson struct {
	Address string `json:"address"`
	PrivKey string `json:"privkey"`
}

// Function to read testnet configuration
func setup(t *testing.T) *Configuration {
	config := GetConfig("testnet")
	return config
}

func getServiceAccountTestnet(t *testing.T, config Configuration) testaccount {
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
	return testaccount{Address: serviceAddress, Key: serviceAccount.Keys[0], Signer: serviceSigner}
}
func readAccountfile(t *testing.T) []testaccountjson {
	// function to read and umnarshal the accounts json
	// Read test account file and append new created account to the list
	jsonfile, err := ioutil.ReadFile("testaccounts.json")
	if err != nil {
		t.Error("Could not read testaccounts.json file")
	}
	testaccountdata := []testaccountjson{}
	json.Unmarshal(jsonfile, &testaccountdata)
	return testaccountdata
}

// Function to create a new account on the testnet using the Veolet service account as the payer/creator
func createAccountTestnet(t *testing.T, config Configuration) testaccount {
	ctx := context.Background()

	accountKey, signer, privKey := createAccountCreds(t)
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	defer c.Close()

	//Get service account
	serviceAccount := getServiceAccountTestnet(t, config)
	// Use the templates package to create a new account creation transaction
	tx := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, serviceAccount.Address)

	// Set the transaction payer and proposal key
	tx.SetPayer(serviceAccount.Address)
	tx.SetProposalKey(
		serviceAccount.Address,
		serviceAccount.Key.Index,
		serviceAccount.Key.SequenceNumber,
	)

	// Get the latest sealed block to use as a reference block
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		t.Error("failed to fetch latest block")
	}
	tx.SetReferenceBlockID(latestBlock.ID)

	// Sign and submit the transaction
	err = tx.SignEnvelope(serviceAccount.Address, serviceAccount.Key.Index, serviceAccount.Signer)
	if err != nil {
		panic("failed to sign transaction envelope")
	}

	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		panic("failed to send transaction to network")
	}

	var address flow.Address
	for {
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
	//serviceAccount.Key.SequenceNumber++

	// unmarshal
	testaccountdata := readAccountfile(t)

	// Append new account
	newAccount := &testaccountjson{
		Address: address.Hex(),
		PrivKey: hex.EncodeToString(privKey.Encode()),
	}
	testaccountdata = append(testaccountdata, *newAccount)
	// Preparing the data to be marshalled and written.
	dataBytes, err := json.MarshalIndent(testaccountdata, "", "   ")
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("testaccounts.json", dataBytes, 0644)
	if err != nil {
		t.Error(err)
	}

	return testaccount{Address: address, Key: accountKey, Signer: signer}
}

// Function to create a Veolet collection for given account
func setupAccountTestnet(t *testing.T, config Configuration, targetAccount testaccount) {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}

	// Read transaction script
	setupcode, err := CadenceCode.ReadFile("cadence/transactions/SetupAccount.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	// Replace placeholder addresses
	setupcode = ReplaceAddressPlaceholders(setupcode, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")

	// Get service account
	serviceAccount := getServiceAccountTestnet(t, config)

	// Send transaction with targetAccount as authorizer/proposer and service account as payer
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		t.Error("failed to fetch latest block")
	}
	tx := flow.NewTransaction().
		SetScript(setupcode).
		SetProposalKey(targetAccount.Address, targetAccount.Key.Index, targetAccount.Key.SequenceNumber).
		SetPayer(serviceAccount.Address).
		SetReferenceBlockID(latestBlock.ID).
		AddAuthorizer(targetAccount.Address)

	// Sign payload and envelope
	err = tx.SignPayload(targetAccount.Address, targetAccount.Key.Index, targetAccount.Signer)
	if err != nil {
		t.Error("Could not sign payload")
	}

	err = tx.SignEnvelope(serviceAccount.Address, serviceAccount.Key.Index, serviceAccount.Signer)
	if err != nil {
		t.Error("Could not sign envelope")
	}
	// Send transaction
	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		t.Error("Failed to submit transaction: ", err)
	}

	// Get transaction result
	result, err := c.GetTransactionResult(ctx, tx.ID())
	if err != nil {
		t.Error("Transaction failed:", err)
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			t.Error("Transaction failed:", err)
		}
	}
}

func mintTokenTestnet(t *testing.T, receiver testaccount, config Configuration) {

	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}

	//Read script file
	transactioncode, err := CadenceCode.ReadFile("cadence/transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	serviceaccount := getServiceAccountTestnet(t, config)
	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(receiver.Address))
	arguments = append(arguments, cadence.NewString("testNFT.com/test"))
	arguments = append(arguments, cadence.NewString("creatorName"))
	arguments = append(arguments, cadence.NewAddress(receiver.Address))
	arguments = append(arguments, cadence.NewUInt64(uint64(time.Now().Unix())))
	arguments = append(arguments, cadence.NewString("caption"))
	arguments = append(arguments, cadence.NewString("hash"))
	arguments = append(arguments, cadence.NewUInt16(1))

	// Get the latest sealed block to use as a reference block
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		panic("failed to fetch latest block")
	}
	// Send transaction
	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(serviceaccount.Address, serviceaccount.Key.Index, serviceaccount.Key.SequenceNumber).
		SetPayer(serviceaccount.Address).
		SetReferenceBlockID(latestBlock.ID).
		AddAuthorizer(serviceaccount.Address)

	for i := 0; i < len(arguments); i++ {
		err = tx.AddArgument(arguments[i])
		if err != nil {
			t.Error("Can't add argument")
		}
	}
	// Veolet account (Service account) signs envelope as minter (authorizer, proposer & payer)
	err = tx.SignEnvelope(serviceaccount.Address, serviceaccount.Key.Index, serviceaccount.Signer)
	if err != nil {
		t.Error("Could not sign envelope with service account")
	}
	// Send transaction
	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		t.Error("Failed to submit transaction: ", err)
	}

	// Get transaction result
	result, err := c.GetTransactionResult(ctx, tx.ID())
	if err != nil {
		t.Error("Transaction failed:", err)
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			t.Error("Transaction failed:", err)
		}
	}
}

func transferTokenTestnet(t *testing.T, config Configuration, sender testaccount, recipient testaccount, tokenID uint64) {
	// Function to transfer NFT from one account to another

	// Connect to Flow network
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	// Read script file
	transfercode, err := CadenceCode.ReadFile("cadence/transactions/Transfer.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}

	// Get service account as payer
	serviceAccount := getServiceAccountTestnet(t, config)

	// Get sender account (for current sequence number)
	senderAccount, err := c.GetAccount(ctx, sender.Address)
	if err != nil {
		t.Error("Could not get sender account", err)
	}

	transfercode = ReplaceAddressPlaceholders(transfercode, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")

	// Get the latest sealed block to use as a reference block
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		panic("failed to fetch latest block")
	}
	tx := flow.NewTransaction().
		SetScript(transfercode).
		SetProposalKey(senderAccount.Address, senderAccount.Keys[0].Index, senderAccount.Keys[0].SequenceNumber).
		SetPayer(serviceAccount.Address).
		SetReferenceBlockID(latestBlock.ID).
		AddAuthorizer(senderAccount.Address)
	tx.AddArgument(cadence.NewAddress(recipient.Address)) // Add recipient argument
	tx.AddArgument(cadence.NewUInt64(tokenID))            // Add tokenID argument

	// Sender signs payload (authorizer/proposer)
	err = tx.SignPayload(senderAccount.Address, senderAccount.Keys[0].Index, sender.Signer)
	if err != nil {
		t.Error("Could not sign payload")
	}
	// Service account signs envelope as payer
	err = tx.SignEnvelope(serviceAccount.Address, serviceAccount.Key.Index, serviceAccount.Signer)
	if err != nil {
		t.Error("Could not sign envelope with service account")
	}
	// Send transaction
	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		t.Error("Failed to submit transaction: ", err)
	}

	// Get transaction result
	result, err := c.GetTransactionResult(ctx, tx.ID())
	if err != nil {
		t.Error("Transaction failed:", err)
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			t.Error("Transaction failed:", err)
		}
	}

}

func fetchCollectionTestnet(t *testing.T, config Configuration, target testaccount) cadence.Value {
	// Function to fetch the token ID's of target account
	// Read script file
	fetchscript, err := CadenceCode.ReadFile("cadence/scripts/FetchCollection.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}
	fetchscript = ReplaceAddressPlaceholders(fetchscript, config.Contractaddresses.NonFungibleToken, "", "", "")
	result, _ := ExecuteScript(config.Network.Host, fetchscript, true, []cadence.Value{cadence.NewAddress(target.Address)})
	if err != nil {
		t.Error("Could not execute script", err)
	}
	return result
}

func readRandomAccount(t *testing.T) testaccountjson {
	// function to read a random account from the testaccounts.txt to use for tests.
	testaccountsfile := readAccountfile(t)
	randidx := rand.Intn(len(testaccountsfile)-1) + 1
	return testaccountsfile[randidx]
}

func getTestingAccount(t *testing.T, config Configuration) testaccount {
	// function to get a random testing account thats not the service account
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	accounts := readAccountfile(t)
	var testingaccount testaccount
	if len(accounts) < 2 {
		testingaccount = createAccountTestnet(t, config)
		setupAccountTestnet(t, config, testingaccount)
	} else {
		jsontestaccount := readRandomAccount(t)
		sigAlgo := crypto.StringToSignatureAlgorithm("ECDSA_P256")
		privKey, err := crypto.DecodePrivateKeyHex(sigAlgo, jsontestaccount.PrivKey)
		if err != nil {
			t.Error("Could not decode private key")
		}
		hashAlgo := crypto.StringToHashAlgorithm("SHA3_256")
		signer := crypto.NewInMemorySigner(privKey, hashAlgo)
		flowacc, err := c.GetAccount(ctx, flow.HexToAddress(jsontestaccount.Address))
		if err != nil {
			t.Error("Could not fetch test account from FLOW testnet")
		}
		testingaccount = testaccount{Address: flowacc.Address, Key: flowacc.Keys[0], Signer: signer}
	}
	return testingaccount
}

func fetchNFTTestnet(t *testing.T, config Configuration, target testaccount, tokenID uint64) cadence.Value {
	// function to fetch information about a single NFT

	// Read script file
	fetchscript, err := CadenceCode.ReadFile("cadence/scripts/ReadNFT.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}
	fetchscript = ReplaceAddressPlaceholders(fetchscript, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")
	result, _ := ExecuteScript(config.Network.Host, fetchscript, true, []cadence.Value{cadence.NewAddress(target.Address), cadence.NewUInt64(tokenID)})
	if err != nil {
		t.Error("Could not execute script", err)
	}
	return result
}

func setMediaURLTestnet(t *testing.T, config Configuration, target testaccount, tokenID uint64, newurl string) {
	// Function to change the media URL of an NFT (originally set URL will remain)

	// Connect to Flow network
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	// Read script file
	code, err := CadenceCode.ReadFile("cadence/transactions/SetMediaURL.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}
	code = ReplaceAddressPlaceholders(code, "", config.Contractaddresses.Veolet, "", "")

	// Get service account (payer)
	serviceAccount := getServiceAccountTestnet(t, config)

	// Get the latest sealed block to use as a reference block
	latestBlock, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		panic("failed to fetch latest block")
	}
	tx := flow.NewTransaction().
		SetScript(code).
		SetProposalKey(target.Address, target.Key.Index, target.Key.SequenceNumber).
		SetPayer(serviceAccount.Address).
		SetReferenceBlockID(latestBlock.ID).
		AddAuthorizer(target.Address)

	tx.AddArgument(cadence.NewString(newurl))  // Add newURL argument
	tx.AddArgument(cadence.NewUInt64(tokenID)) // Add tokenID argument

	// Sender signs payload (authorizer/proposer)
	err = tx.SignPayload(target.Address, target.Key.Index, target.Signer)
	if err != nil {
		t.Error("Could not sign payload")
	}
	// Service account signs envelope as payer
	err = tx.SignEnvelope(serviceAccount.Address, serviceAccount.Key.Index, serviceAccount.Signer)
	if err != nil {
		t.Error("Could not sign envelope with service account")
	}
	// Send transaction
	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		t.Error("Failed to submit transaction: ", err)
	}

	// Get transaction result
	result, err := c.GetTransactionResult(ctx, tx.ID())
	if err != nil {
		t.Error("Transaction failed:", err)
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			t.Error("Transaction failed:", err)
		}
	}
}

func getStorageInfoTestnet(t *testing.T, config Configuration, target testaccount) cadence.Value {
	code, err := CadenceCode.ReadFile("cadence/scripts/StorageUsed.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}
	result, _ := ExecuteScript(config.Network.Host, code, true, []cadence.Value{cadence.NewAddress(target.Address)})
	if err != nil {
		t.Error("Could not execute script", err)
	}
	return result
}

func TestCreateAndSetupAccountTestnet(t *testing.T) {
	// Should be able to create a new account and set it up to hold a Veolet collection
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
	t.Log("Start MintTokenTestnet test")
	config := setup(t)

	// Get testnet account for testing
	testingaccount := getTestingAccount(t, *config)

	// Get service account that holds minter resource
	serviceAccount := getServiceAccountTestnet(t, *config)

	// Get collection length before minting
	servicecollection := fetchCollectionTestnet(t, *config, serviceAccount).(cadence.Array).Values

	// Use service account to mint token into its own account
	mintTokenTestnet(t, serviceAccount, *config)
	// Assert that the collection of the receiver has been updated
	servicecollection_new := fetchCollectionTestnet(t, *config, serviceAccount).(cadence.Array).Values
	if len(servicecollection_new)-len(servicecollection) != 1 {
		t.Error("Expected difference of 1, got ", len(servicecollection_new)-len(servicecollection))
	}

	// Get testing account collection length before minting
	usercollection := fetchCollectionTestnet(t, *config, testingaccount).(cadence.Array).Values

	// Get service account again (for current sequence number)
	serviceAccount = getServiceAccountTestnet(t, *config)
	// Use service account to mint token into testing account
	mintTokenTestnet(t, testingaccount, *config)
	// Assert that the collection of the receiver has been updated
	usercollection_new := fetchCollectionTestnet(t, *config, testingaccount).(cadence.Array).Values
	if len(usercollection_new)-len(usercollection) != 1 {
		t.Error("Expected difference of 1, got ", len(usercollection_new)-len(usercollection))
	}

}

func TestTransferTokenTestnet(t *testing.T) {
	// Should be able to send tokens from one adress to another
	t.Log("Start TransferTokenTestnet test")
	config := setup(t)

	// Get testnet account for testing
	testingaccount := getTestingAccount(t, *config)

	//Get service account for minting
	serviceAccount := getServiceAccountTestnet(t, *config)

	// Get length of collection before minting
	servicecollection := fetchCollectionTestnet(t, *config, serviceAccount).(cadence.Array).Values

	// Use Veolet account to mint token into its own account
	mintTokenTestnet(t, serviceAccount, *config)
	// Assert that the collection of the receiver has been updated
	servicecollection_new := fetchCollectionTestnet(t, *config, serviceAccount).(cadence.Array).Values
	if len(servicecollection_new)-len(servicecollection) != 1 {
		t.Error("Expected difference of 1, got ", len(servicecollection_new)-len(servicecollection))
	}

	//Get service account again for correct sequence number
	serviceAccount = getServiceAccountTestnet(t, *config)

	// Get collection before transfer
	usercollection := fetchCollectionTestnet(t, *config, testingaccount).(cadence.Array).Values

	// Send one random token from service account into testing account
	randomIndex := rand.Intn(len(servicecollection_new))
	transferTokenTestnet(t, *config, serviceAccount, testingaccount, uint64(servicecollection_new[randomIndex].(cadence.UInt64)))

	// Assert that token has been transferred by fetching collections
	servicecollection_new = fetchCollectionTestnet(t, *config, serviceAccount).(cadence.Array).Values
	usercollection_new := fetchCollectionTestnet(t, *config, testingaccount).(cadence.Array).Values
	if len(usercollection_new)-len(usercollection) != 1 {
		t.Error("Expected difference of 1, got ", len(usercollection_new)-len(usercollection))
	}
	if len(servicecollection_new) != len(servicecollection) {
		t.Error("Expected difference 0, got ", len(servicecollection_new) != len(servicecollection))
	}
}

func TestGetVeoletInformationTestnet(t *testing.T) {
	// Should be able to fetch all fields of Veolet token
	t.Log("Start GetVeoletInformationTestnet test")

	config := setup(t)
	// Get testnet account for testing
	testingaccount := getTestingAccount(t, *config)

	// Use Veolet account to mint token into testing account
	mintTokenTestnet(t, testingaccount, *config)

	// Get all NFT ID's of the account
	usercollection := fetchCollectionTestnet(t, *config, testingaccount).(cadence.Array).Values
	randomIndex := rand.Intn(len(usercollection))

	// Fetch the information about the minted NFT
	result := fetchNFTTestnet(t, *config, testingaccount, uint64(usercollection[randomIndex].(cadence.UInt64))).(cadence.Optional).Value.(cadence.Optional).Value.(cadence.Resource).Fields
	if len(result) != 10 {
		t.Error("Expected length 10, got", len(result))
	}
	// Check if the url is correct
	if strings.Trim(result[2].String(), "\"") != "testNFT.com/test" {
		t.Error("Expected \"testNFT.com/test\", got: ", result[2].String())
	}
}

func TestSetMediaURLTestnet(t *testing.T) {
	// Should be able to make use of the "setMediaURL" method of the Veolet Collection
	t.Log("Start SetMediaURL test")

	config := setup(t)
	serviceAccount := getServiceAccountTestnet(t, *config)

	// Use Veolet account to mint token into its own account
	mintTokenTestnet(t, serviceAccount, *config)

	// Get service account again for actual sequence number
	serviceAccount = getServiceAccountTestnet(t, *config)
	// Get random ID
	collection := fetchCollectionTestnet(t, *config, serviceAccount).(cadence.Array).Values
	randomIndex := rand.Intn(len(collection))
	randomID := uint64(collection[randomIndex].(cadence.UInt64))
	// Change the settable mediaURL of the minted Token using the private collection.
	setMediaURLTestnet(t, *config, serviceAccount, randomID, "newurl.com")

	// Assert that the URL has been changed
	token := fetchNFTTestnet(t, *config, serviceAccount, randomID).(cadence.Optional).Value.(cadence.Optional).Value.(cadence.Resource).Fields
	if strings.Trim(token[9].String(), "\"") != "newurl.com" {
		t.Error("Expected \"newurl.com\", got ", token[9])
	}
}

func TestGetStorageInfoTestnet(t *testing.T) {
	// Should be able to fetch storage capacity and usage of an account
	t.Log("Start GetStorageInfoTestnet test")

	config := setup(t)

	// Get user account
	testingaccount := getTestingAccount(t, *config)

	// Check storage usage and capacity of user account
	result := getStorageInfoTestnet(t, *config, testingaccount).(cadence.Array).Values
	cap := int(result[0].(cadence.UInt64))
	used := int(result[1].(cadence.UInt64))
	t.Log("Useraccount has storage cap: ", cap)
	t.Log("Useraccount has storage usage: ", used)

	// Use Veolet account to mint token into useraccount to test if storage changes
	mintTokenTestnet(t, testingaccount, *config)

	// Get new storage info
	newresult := getStorageInfoTestnet(t, *config, testingaccount).(cadence.Array).Values
	newcap := int(newresult[0].(cadence.UInt64))
	newused := int(newresult[1].(cadence.UInt64))
	t.Log("Useraccount has new storage cap: ", newcap)
	t.Log("Useraccount has new storage usage: ", newused)

	if newcap != cap {
		t.Error("Expected same capacity, got: ", cap, newcap)
	}
	if used == newused {
		t.Error("Expected different usage, got: ", used, newused)
	}
}

func TestGetBalanceTestnet(t *testing.T) {
	// Should be able to get accounts balance
	t.Log("Start GetBalanceTestnet test")

	config := setup(t)
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		t.Error("Failed to establish connection with Access API")
	}
	defer c.Close()

	testingaccount := getTestingAccount(t, *config)
	flowTestingaccount, err := c.GetAccount(ctx, testingaccount.Address)
	if err != nil {
		t.Error("Could not get Account")
	}
	balance := flowTestingaccount.Balance
	t.Log("Testaccounts balance: ", balance)
}
