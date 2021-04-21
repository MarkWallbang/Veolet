package test

import (
	"io/ioutil"
	"testing"
	"veolet/lib"

	//"time"

	//"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	//"github.com/onflow/flow-go-sdk/test"
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

func setupTestaccount(b *emulator.Blockchain, t *testing.T, nftacc testaccount, veoletacc testaccount) {
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
		SetProposalKey(nftacc.address, nftacc.key.Index, nftacc.key.SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		//SetReferenceBlockID(latestblock.ID).
		AddAuthorizer(nftacc.address)

	err = tx.SignPayload(nftacc.address, nftacc.key.Index, nftacc.signer)
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
	_, err = b.ExecuteNextTransaction()
	if err != nil {
		t.Error("Execution of transaction failed")
	}
	// Commit block
	_, err = b.CommitBlock()
	if err != nil {
		t.Error("CommitBlock failed")
	}

}

func deployVeoletContracts(b *emulator.Blockchain, t *testing.T) (testaccount, testaccount) {
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
	//AccountKey, Signer := accountKeys.NewWithSigner()
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

	// Setup account for nftAddress holder to hold Veolet NFTs
	nftacc := testaccount{address: nftAddr, key: nil, signer: nil}
	veoletacc := testaccount{address: VeoletAddr, key: AccountKey, signer: Signer}

	t.Log(nftacc)
	t.Log(veoletacc)
	return nftacc, veoletacc
}

func TestDeployVeoletContract(t *testing.T) {
	// Should be able to deploy Veolet contract (and the NFT core contract if runtimeenv = emulator)
	t.Log("Start DeployContract test")

	b := NewEmulator()
	deployVeoletContracts(b, t)
}

/*func TestSetupAccount(t *testing.T) {
	// Should be able to create new Flow account and Veolet vault for the new account
	t.Log("Start SetupAccount test")

	b := NewEmulator()
	// Create new accounts and deploy contracts
	nftacc, veoletacc := deployVeoletContracts(b, t)
	// set up the test account to be able to store Veolet tokens
	setupTestaccount(b, t, nftacc, veoletacc)
}

func TestMintToken(t *testing.T) {
	// Should be able to mint new token
	t.Log("Start MintToken test")

	b := NewEmulator()
	// Create new accounts and deploy contracts
	nftacc, veoletacc := deployVeoletContracts(b, t)

	// Use Veolet account to mint tokens
	//Read script file
	transactioncode, err := ioutil.ReadFile("../../../transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = lib.ReplaceAddressPlaceholders(transactioncode, nftacc.address.Hex(), veoletacc.address.Hex(), "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(nftacc.address))
	arguments = append(arguments, cadence.NewString("testNFT.com/test"))
	arguments = append(arguments, cadence.NewString("creatorName"))
	arguments = append(arguments, cadence.NewAddress(nftacc.address))
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

	err = tx.SignPayload(nftacc.address, nftacc.key.Index, nftacc.signer)
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
	_, err = b.ExecuteNextTransaction()
	if err != nil {
		t.Error("Execution of transaction failed")
	}
	// Commit block
	_, err = b.CommitBlock()
	if err != nil {
		t.Error("CommitBlock failed")
	}
}

func TestSendToken(t *testing.T) {
	// Should be able to send tokens from one adress to another
	// TODO It should be tested to send tokens to adresses with and without a valid Veolet vault
}
*/
