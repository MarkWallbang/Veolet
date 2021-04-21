package test

import (
	"io/ioutil"
	"testing"
	"time"
	"veolet/lib"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
)

/*
	This file tests all core functions of the Veolet contract and its transactions/scripts. The test functions will use the package "lib", which
	defines functions to send transactions, create and setup accounts, mint tokens etc.
*/

type testaccount struct {
	address flow.Address
	key     *flow.AccountKey
	signer  crypto.Signer
}

// newEmulator returns a emulator object for testing
func NewEmulator() *emulator.Blockchain {
	b, err := emulator.NewBlockchain()
	if err != nil {
		panic(err)
	}
	return b
}

func setupTestaccount(b *emulator.Blockchain, t *testing.T, nftacc testaccount, veoletacc testaccount, useracc testaccount) {
	// Setup Veolet wallet for the nft account (the veolet account has it set up from contract init() function)
	// 1. Read transaction script
	setupcode, err := ioutil.ReadFile("../../../transactions/SetupAccount.cdc")
	if err != nil {
		t.Error("Cannot read script file")
	}
	// 2. Replace placeholder addresses
	setupcode = lib.ReplaceAddressPlaceholders(setupcode, nftacc.address.Hex(), veoletacc.address.Hex(), "", "")

	// 3. Create Transaction
	//latestblock, err := b.GetLatestBlock()
	tx := flow.NewTransaction().
		SetScript(setupcode).
		SetProposalKey(useracc.address, useracc.key.Index, useracc.key.SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		//SetReferenceBlockID(latestblock.ID).
		AddAuthorizer(useracc.address)

	err = tx.SignPayload(useracc.address, useracc.key.Index, useracc.signer)
	if err != nil {
		t.Error("Could not sign payload")
	}
	err = tx.SignEnvelope(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().Signer())
	if err != nil {
		t.Error("Could not sign envelope with service account")
	}
	// Add transaction to block
	err = b.AddTransaction(*tx)
	if err != nil {
		t.Error("Adding of transaction failed")
	}
	// Execute transaction
	result, err := b.ExecuteNextTransaction()
	if err != nil {
		t.Error("Execution of transaction failed")
	}
	if result.Error != nil {
		t.Error("Transaction Reverted")
	}
	// Commit block
	_, err = b.CommitBlock()
	if err != nil {
		t.Error("CommitBlock failed")
	}

}
func createAccountCreds(t *testing.T) (*flow.AccountKey, crypto.Signer) {
	pubKeyHex, privKeyHex := lib.GenerateKeys("ECDSA_P256")
	sigAlgo := crypto.StringToSignatureAlgorithm("ECDSA_P256")
	publicKey, err := crypto.DecodePublicKeyHex(sigAlgo, pubKeyHex)
	if err != nil {
		t.Error(err)
	}
	hashAlgo := crypto.StringToHashAlgorithm("SHA3_256")

	AccountKey := flow.NewAccountKey().
		SetPublicKey(publicKey).
		SetSigAlgo(sigAlgo).
		SetHashAlgo(hashAlgo).
		SetWeight(flow.AccountKeyWeightThreshold)
	privKey, err := crypto.DecodePrivateKeyHex(sigAlgo, privKeyHex)
	Signer := crypto.NewInMemorySigner(privKey, hashAlgo)
	return AccountKey, Signer
}

/*
Function to deploy necessary contracts, and create Accounts for them. Returns 3 accounts, one for the NFT core contract (no keys),
one for the Veolet contract, and a user account without a contract.
*/
func deployVeoletContracts(b *emulator.Blockchain, t *testing.T) (testaccount, testaccount, testaccount) {

	// Create credentials for Veolet Contract account
	AccountKey, Signer := createAccountCreds(t)

	// Should be able to deploy a contract as a new account with no keys.
	code, err := ioutil.ReadFile("../../../contracts/NonFungibleToken.cdc")
	if err != nil {
		t.Error(err)
	}
	nftAddr, err := b.CreateAccount(
		nil,
		[]templates.Contract{
			{
				Name:   "NonFungibleToken",
				Source: string(code),
			},
		})
	if err != nil {
		t.Error(err)
	}
	_, err = b.CommitBlock()
	if err != nil {
		t.Error(err)
	}

	// Should be able to deploy Veolet contract as a new account with one key.
	veoletcode, err := ioutil.ReadFile("../../../contracts/Veolet.cdc")
	if err != nil {
		t.Error(err)
	}
	// Replace address placeholder
	veoletcode = lib.ReplaceAddressPlaceholders(veoletcode, nftAddr.String(), "", "", "")

	VeoletAddr, err := b.CreateAccount(
		[]*flow.AccountKey{AccountKey},
		[]templates.Contract{
			{
				Name:   "Veolet",
				Source: string(veoletcode),
			},
		})
	if err != nil {
		t.Error(err)
	}
	_, err = b.CommitBlock()
	if err != nil {
		t.Error(err)
	}

	// Should be able to create account without contract (user)
	UserAccountKey, UserSigner := createAccountCreds(t)
	UserAddr, err := b.CreateAccount([]*flow.AccountKey{UserAccountKey}, nil)
	if err != nil {
		t.Error(err)
	}
	_, err = b.CommitBlock()
	if err != nil {
		t.Error(err)
	}

	nftacc := testaccount{address: nftAddr, key: nil, signer: nil}
	veoletacc := testaccount{address: VeoletAddr, key: AccountKey, signer: Signer}
	useracc := testaccount{address: UserAddr, key: UserAccountKey, signer: UserSigner}
	t.Log("NFT Account", nftacc)
	t.Log("Veolet Account", veoletacc)
	t.Log("User Account", useracc)
	return nftacc, veoletacc, useracc
}

func mintToken(t *testing.T, b *emulator.Blockchain, receiver testaccount, veoletacc *testaccount, nftacc testaccount) {
	//Read script file
	transactioncode, err := ioutil.ReadFile("../../../transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = lib.ReplaceAddressPlaceholders(transactioncode, nftacc.address.Hex(), veoletacc.address.Hex(), "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(receiver.address))
	arguments = append(arguments, cadence.NewString("testNFT.com/test"))
	arguments = append(arguments, cadence.NewString("creatorName"))
	arguments = append(arguments, cadence.NewAddress(receiver.address))
	arguments = append(arguments, cadence.NewUInt64(uint64(time.Now().Unix())))
	arguments = append(arguments, cadence.NewString("caption"))
	arguments = append(arguments, cadence.NewString("hash"))
	arguments = append(arguments, cadence.NewUInt16(1))

	// Send transaction
	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(veoletacc.address, veoletacc.key.Index, veoletacc.key.SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		//SetReferenceBlockID(latestblock.ID).
		AddAuthorizer(veoletacc.address)

	for i := 0; i < len(arguments); i++ {
		err = tx.AddArgument(arguments[i])
		if err != nil {
			t.Error("Can't add argument")
		}
	}
	// Veolet address signs payload as minter (authorizer/proposer)
	err = tx.SignPayload(veoletacc.address, veoletacc.key.Index, veoletacc.signer)
	if err != nil {
		t.Error("Could not sign payload")
	}

	// Service account signs envelope as payer
	err = tx.SignEnvelope(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().Signer())
	if err != nil {
		t.Error("Could not sign envelope with service account")
	}
	// Add transaction to block
	err = b.AddTransaction(*tx)
	if err != nil {
		t.Error("Adding of transaction failed")
	}
	// Execute transaction
	result, err := b.ExecuteNextTransaction()
	if err != nil {
		t.Error("Execution of transaction failed")
	}
	if result.Error != nil {
		t.Log("Transaction Reverted")
		t.Error(result.Error)
	}

	// Commit block
	_, err = b.CommitBlock()
	if err != nil {
		t.Error("CommitBlock failed")
	}

	// Increment Sequence number of veolet key
	veoletacc.key.SequenceNumber++
}

// Function to transfer Veolet NFT from one account to the other
func transferToken(t *testing.T, b *emulator.Blockchain, recipient *testaccount, sender *testaccount, tokenID uint64, nftacc testaccount, veoletacc testaccount, shouldRevert bool) {

	transfercode, err := ioutil.ReadFile("../../../transactions/Transfer.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}
	transfercode = lib.ReplaceAddressPlaceholders(transfercode, nftacc.address.Hex(), veoletacc.address.Hex(), "", "")
	tx := flow.NewTransaction().
		SetScript(transfercode).
		SetProposalKey(sender.address, sender.key.Index, sender.key.SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		//SetReferenceBlockID(latestblock.ID).
		AddAuthorizer(sender.address)
	tx.AddArgument(cadence.NewAddress(recipient.address)) // Add recipient argument
	tx.AddArgument(cadence.NewUInt64(tokenID))            // Add tokenID argument

	// Sender signs payload (authorizer/proposer)
	err = tx.SignPayload(sender.address, sender.key.Index, sender.signer)
	if err != nil {
		t.Error("Could not sign payload")
	}
	// Service account signs envelope as payer
	err = tx.SignEnvelope(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().Signer())
	if err != nil {
		t.Error("Could not sign envelope with service account")
	}
	// Add transaction to block
	err = b.AddTransaction(*tx)
	if err != nil {
		t.Error("Adding of Transfer transaction failed")
	}
	// Execute transaction
	result, err := b.ExecuteNextTransaction()
	if err != nil {
		t.Error("Execution of Transfer transaction failed")
	}
	if shouldRevert {
		if result.Error == nil {
			t.Log("Transfer transaction not Reverted, but expected revert")
		}

	} else {
		if result.Error != nil {
			t.Log("Transfer transaction Reverted, expected success")
			t.Error(result.Error)
		}

	}

	// Commit block
	_, err = b.CommitBlock()
	if err != nil {
		t.Error("CommitBlock failed")
	}

	sender.key.SequenceNumber++
}
func fetchCollection(t *testing.T, b *emulator.Blockchain, target testaccount, nftacc testaccount) cadence.Value {
	fetchscript, err := ioutil.ReadFile("../../../scripts/FetchCollection.cdc")
	if err != nil {
		t.Error("Could not read script file")
	}
	fetchscript = lib.ReplaceAddressPlaceholders(fetchscript, nftacc.address.Hex(), "", "", "")
	result, err := b.ExecuteScript(fetchscript, [][]byte{jsoncdc.MustEncode(cadence.NewAddress(target.address))})
	if err != nil {
		t.Error("Could not execute script", err)
	}
	return result.Value
}

func TestDeployVeoletContract(t *testing.T) {
	// Should be able to deploy Veolet contract (and the NFT core contract if runtimeenv = emulator)
	t.Log("Start DeployContract test")

	b := NewEmulator()
	deployVeoletContracts(b, t)
}

func TestSetupAccount(t *testing.T) {
	// Should be able to create new Flow account and Veolet vault for the new account
	t.Log("Start SetupAccount test")

	b := NewEmulator()
	// Create new accounts and deploy contracts
	nftacc, veoletacc, useracc := deployVeoletContracts(b, t)
	// set up the test account to be able to store Veolet tokens
	setupTestaccount(b, t, nftacc, veoletacc, useracc)
}

func TestMintToken(t *testing.T) {
	// Should be able to mint new token
	t.Log("Start MintToken test")

	b := NewEmulator()
	// Create new accounts and deploy contracts
	nftacc, veoletacc, useracc := deployVeoletContracts(b, t)
	setupTestaccount(b, t, nftacc, veoletacc, useracc)

	// Use Veolet account to mint token into its own account
	mintToken(t, b, veoletacc, &veoletacc, nftacc)
	// Assert that the collection of the receiver has been updated
	veoletcollection := fetchCollection(t, b, veoletacc, nftacc)
	t.Log("FetchCollection result: ", veoletcollection)

	// Use Veolet account to mint token into other users account
	mintToken(t, b, useracc, &veoletacc, nftacc)
	// Assert that the collection of the receiver has been updated
	usercollection := fetchCollection(t, b, useracc, nftacc)
	t.Log("FetchCollection result: ", usercollection)
}

func TestSendToken(t *testing.T) {
	// Should be able to send tokens from one adress to another
	// TODO It should be tested to send tokens to adresses with and without a valid Veolet vault
	t.Log("Start SendToken test")

	b := NewEmulator()
	// Create new accounts and deploy contracts
	nftacc, veoletacc, useracc := deployVeoletContracts(b, t)
	// Setup User account
	setupTestaccount(b, t, nftacc, veoletacc, useracc)
	// Use Veolet account to mint token into its own account
	mintToken(t, b, veoletacc, &veoletacc, nftacc)
	// Assert that the collection of the receiver has been updated
	veoletcollection := fetchCollection(t, b, veoletacc, nftacc)
	t.Log("FetchCollection result: ", veoletcollection)

	// Send created token from Veolet account into User account
	transferToken(t, b, &useracc, &veoletacc, 0, nftacc, veoletacc, false)
	// Assert that token has been transferred by fetching collections
	veoletcollection = fetchCollection(t, b, veoletacc, nftacc)
	usercollection := fetchCollection(t, b, useracc, nftacc)
	t.Log("FetchCollection Veolet account result: ", veoletcollection)
	t.Log("FetchCollection User account result: ", usercollection)

	/*
		Send created token from User Account into NFT Contract Account. This transaction is
		supposed to be reverted as the NFT Contract Account does not have a Veolet Vault
	*/
	transferToken(t, b, &nftacc, &useracc, 0, nftacc, veoletacc, true) // Should revert
	// Assert that User Account still holds NFT, and NFT Contract Account does not hold NFT
	nftacccollection := fetchCollection(t, b, nftacc, nftacc)
	usercollection = fetchCollection(t, b, useracc, nftacc)
	t.Log("FetchCollection User account result: ", usercollection)
	t.Log("FetchCollection NFT account result: ", nftacccollection)
}

func TestGetVeoletInformation(t *testing.T) {
	// Should be able to fetch all fields of Veolet token
}

func TestSetMediaURL(t *testing.T) {
	// Should be able to make use of the "setMediaURL" method of the Veolet Collection
}
