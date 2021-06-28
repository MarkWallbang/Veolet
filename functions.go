package veolet

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
)

//go:embed flow.json
var Configjson []byte

//go:embed cadence
var CadenceCode embed.FS

// Account type
type Account struct {
	Address flow.Address
	Privkey crypto.PrivateKey
}

func GetFS() embed.FS {
	return CadenceCode
}

// structure of flow.json config
type FlowConfiguration struct {
	Emulators struct {
		Default struct {
			Port           int    `json:"port"`
			ServiceAccount string `json:"serviceAccount"`
		} `json:"default"`
	} `json:"emulators"`

	Contracts struct {
		NonFungibleToken struct {
			Source  string `json:"source"`
			Aliases struct {
				Testnet string `json:"testnet"`
			} `json:"aliases"`
		} `json:"NonFungibleToken"`
		Veolet string `json:"Veolet"`
	} `json:"contracts"`

	Networks struct {
		Emulator struct {
			Host  string `json:"host"`
			Chain string `json:"chain"`
		} `json:"emulator"`

		Testnet struct {
			Host  string `json:"host"`
			Chain string `json:"chain"`
		} `json:"testnet"`
	} `json:"networks"`

	Accounts struct {
		Emulator_account struct {
			Address string `json:"address"`
			Keys    string `json:"keys"`
			Chain   string `json:"chain"`
		} `json:"emulator-account"`

		Testnet_account struct {
			Address string `json:"address"`
			Keys    string `json:"keys"`
			Chain   string `json:"chain"`
		} `json:"testnet-account"`
	} `json:"accounts"`

	Deployments struct {
		Emulator struct {
			Emulator_account []string `json:"emulator-account"`
		} `json:"emulator"`

		Testnet struct {
			Testnet_account []string `json:"testnet-account"`
		} `json:"testnet"`
	} `json:"deployments"`

	Contractaddresses struct {
		Emulator struct {
			NonFungibleToken string `json:"NonFungibleToken"`
			Veolet           string `json:"Veolet"`
			FungibleToken    string `json:"FungibleToken"`
			FlowToken        string `json:"FlowToken"`
		} `json:"emulator"`
		Testnet struct {
			NonFungibleToken string `json:"NonFungibleToken"`
			Veolet           string `json:"Veolet"`
			FungibleToken    string `json:"FungibleToken"`
			FlowToken        string `json:"FlowToken"`
		} `json:"testnet"`
	} `json:"contractaddresses"`
}

// Own defined config that will adapt based on runtime env
type Configuration struct {
	Contractaddresses struct {
		NonFungibleToken string
		Veolet           string
		FungibleToken    string
		FlowToken        string
	}
	Account struct {
		Address string
		Keys    string
		Chain   string
	}
	Network struct {
		Host  string
		Chain string
	}
	Deployments struct {
		Account []string
	}
}

func GetConfig(runtimeenv string) *Configuration {

	var configuration FlowConfiguration
	json.Unmarshal(Configjson, &configuration)

	config := new(Configuration)
	if runtimeenv == "emulator" {
		config.Contractaddresses.NonFungibleToken = configuration.Contractaddresses.Emulator.NonFungibleToken
		config.Contractaddresses.Veolet = configuration.Contractaddresses.Emulator.Veolet
		config.Contractaddresses.FungibleToken = configuration.Contractaddresses.Emulator.FungibleToken
		config.Contractaddresses.FlowToken = configuration.Contractaddresses.Emulator.FlowToken

		config.Account.Address = configuration.Accounts.Emulator_account.Address
		config.Account.Keys = configuration.Accounts.Emulator_account.Keys
		config.Account.Chain = configuration.Accounts.Emulator_account.Chain

		config.Network.Host = configuration.Networks.Emulator.Host
		config.Network.Chain = configuration.Networks.Emulator.Chain

		config.Deployments.Account = configuration.Deployments.Emulator.Emulator_account

	} else if runtimeenv == "testnet" {
		config.Contractaddresses.NonFungibleToken = configuration.Contractaddresses.Testnet.NonFungibleToken
		config.Contractaddresses.Veolet = configuration.Contractaddresses.Testnet.Veolet
		config.Contractaddresses.FungibleToken = configuration.Contractaddresses.Testnet.FungibleToken
		config.Contractaddresses.FlowToken = configuration.Contractaddresses.Testnet.FlowToken

		config.Account.Address = configuration.Accounts.Testnet_account.Address
		config.Account.Keys = configuration.Accounts.Testnet_account.Keys
		config.Account.Chain = configuration.Accounts.Testnet_account.Chain

		config.Network.Host = configuration.Networks.Testnet.Host
		config.Network.Chain = configuration.Networks.Testnet.Chain

		config.Deployments.Account = configuration.Deployments.Testnet.Testnet_account
	} else if runtimeenv == "mainnet" {
		//Todo
	} else {
		panic("Invalid runtimeenv")
	}

	return config
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
		setupcode, err := CadenceCode.ReadFile("cadence/transactions/SetupAccount.cdc")
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

func DeployContract(configuration Configuration, target Account, runtimeenv string, contract templates.Contract) *flow.TransactionResult {

	code, err := CadenceCode.ReadFile("cadence/transactions/AddContract.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewString(contract.Name))
	arguments = append(arguments, cadence.NewString(contract.SourceHex()))

	result := SendTransaction(configuration, target, code, arguments)
	return result
}

// Execute a script on the given network
func ExecuteScript(node string, script []byte, script_panik_flag bool, args []cadence.Value) (cadence.Value, error) {
	ctx := context.Background()
	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	// [3]
	result, err := c.ExecuteScriptAtLatestBlock(ctx, script, args)
	if err != nil && script_panik_flag {
		panic(err)
	}

	return result, err
}

// Send a transaction to the network
func SendTransaction(config Configuration, proposer Account, transactioncode []byte, arguments []cadence.Value) *flow.TransactionResult {
	ctx := context.Background()
	flowClient, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("Could not connect to client")
	}

	acc, err := flowClient.GetAccount(ctx, proposer.Address)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(proposer.Privkey, accountKey.HashAlgo)

	// Get service account (as payer)
	serviceAddress, serviceKey, serviceSigner := GetServiceAccount(config)

	block, err := flowClient.GetLatestBlock(ctx, true)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}

	tx := flow.NewTransaction().
		SetScript(transactioncode).
		SetProposalKey(proposer.Address, accountKey.Index, accountKey.SequenceNumber).
		SetReferenceBlockID(block.ID).
		SetPayer(serviceAddress).
		AddAuthorizer(proposer.Address)

	if arguments != nil {
		for i := 0; i < len(arguments); i++ {
			err = tx.AddArgument(arguments[i])
			if err != nil {
				fmt.Println("err:", err.Error())
				panic(err)
			}
		}
	}

	// Check if proposer is the service account. If thats the case, only the envelope is to be signed.
	if serviceAddress.Hex() != proposer.Address.Hex() {
		err = tx.SignPayload(proposer.Address, accountKey.Index, signer)
		if err != nil {
			fmt.Println("err:", err.Error())
			panic(err)
		}
	}

	err = tx.SignEnvelope(serviceAddress, serviceKey.Index, serviceSigner)
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

//Execute transaction to mint new token into service account
func MintToken(configuration Configuration, recipient flow.Address, URL string, creatorName string, creatorAddress flow.Address, caption string, hash string, edition uint16) *flow.TransactionResult {

	// creation timestamp
	var timestamp = uint64(time.Now().Unix())

	//Read script file
	transactioncode, err := CadenceCode.ReadFile("cadence/transactions/MintToken.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, configuration.Contractaddresses.NonFungibleToken, configuration.Contractaddresses.Veolet, "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(recipient))
	arguments = append(arguments, cadence.NewString(URL))
	arguments = append(arguments, cadence.NewString(creatorName))
	arguments = append(arguments, cadence.NewAddress(creatorAddress))
	arguments = append(arguments, cadence.NewUInt64(timestamp))
	arguments = append(arguments, cadence.NewString(caption))
	arguments = append(arguments, cadence.NewString(hash))
	arguments = append(arguments, cadence.NewUInt16(edition))

	// Get service account (as payer)
	serviceAddress := flow.HexToAddress(configuration.Account.Address)
	servicePrivKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, configuration.Account.Keys)
	if err != nil {
		panic("Could not decode priv key.")
	}

	// Send transaction
	result := SendTransaction(configuration, Account{Address: serviceAddress, Privkey: servicePrivKey}, transactioncode, arguments)
	return result
}

func FetchContracts(config Configuration, address flow.Address) map[string][]byte {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}
	account, err := c.GetAccount(ctx, address)

	if err != nil {
		panic("failed to fetch account")
	}
	return account.Contracts
}

func FetchStorageCapacity(config Configuration, address flow.Address) (int, int) {
	// TODO implement method to check if user needs more capacity

	//Read script file
	code, err := CadenceCode.ReadFile("cadence/scripts/StorageUsed.cdc")
	if err != nil {
		panic("Cannot read script file")
	}
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(address))

	result, _ := ExecuteScript(config.Network.Host, code, true, arguments)
	resultarr := result.(cadence.Array).Values
	return int(resultarr[0].(cadence.UInt64)), int(resultarr[1].(cadence.UInt64))
}

func FetchCollection(config Configuration, target flow.Address) cadence.Value {
	// Function to fetch the token ID's of target account
	// Read script file
	fetchscript, err := CadenceCode.ReadFile("cadence/scripts/FetchCollection.cdc")
	if err != nil {
		panic("Could not read script file")
	}
	fetchscript = ReplaceAddressPlaceholders(fetchscript, config.Contractaddresses.NonFungibleToken, "", "", "")
	result, _ := ExecuteScript(config.Network.Host, fetchscript, true, []cadence.Value{cadence.NewAddress(target)})
	if err != nil {
		fmt.Print(err)
		panic("Could not execute script")
	}
	return result
}
func FetchBalance(config Configuration, target flow.Address) uint64 {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("Failed to establish connection with Access API")
	}
	defer c.Close()

	flowaccount, err := c.GetAccount(ctx, target)
	if err != nil {
		panic("Could not get Account")
	}
	balance := flowaccount.Balance
	return balance
}
func FetchNFT(config Configuration, target flow.Address, tokenID uint64) cadence.Value {
	// function to fetch information about a single NFT

	// Read script file
	fetchscript, err := CadenceCode.ReadFile("cadence/scripts/ReadNFT.cdc")
	if err != nil {
		panic("Could not read script file")
	}
	fetchscript = ReplaceAddressPlaceholders(fetchscript, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")
	result, _ := ExecuteScript(config.Network.Host, fetchscript, true, []cadence.Value{cadence.NewAddress(target), cadence.NewUInt64(tokenID)})
	if err != nil {
		panic("Could not execute script")
	}
	return result
}

func FetchCollectionNFTs(config Configuration, target flow.Address) (cadence.Value, error) {
	// function to fetch information about a single NFT

	// Read script file
	fetchscript, err := CadenceCode.ReadFile("cadence/scripts/FetchCollectionNFTs.cdc")
	if err != nil {
		panic(err)
		//panic("Could not read script file")
	}
	fetchscript = ReplaceAddressPlaceholders(fetchscript, config.Contractaddresses.NonFungibleToken, config.Contractaddresses.Veolet, "", "")
	fmt.Println(config.Contractaddresses.NonFungibleToken)
	result, err := ExecuteScript(config.Network.Host, fetchscript, false, []cadence.Value{cadence.NewAddress(target)})
	/*if err != nil {
		panic("Could not execute script")
	}*/

	return result, err
}

// Function to transfer token from one account to the other
func TransferToken(configuration Configuration, sender Account, recipientAddress flow.Address, tokenID uint64) *flow.TransactionResult {

	//Read transaction code
	transactioncode, err := CadenceCode.ReadFile("cadence/transactions/Transfer.cdc")
	if err != nil {
		panic("Cannot read script file")
	}

	// Change placeholder address in script import
	transactioncode = ReplaceAddressPlaceholders(transactioncode, configuration.Contractaddresses.NonFungibleToken, configuration.Contractaddresses.Veolet, "", "")

	//define arguments
	var arguments []cadence.Value
	arguments = append(arguments, cadence.NewAddress(recipientAddress))
	arguments = append(arguments, cadence.NewUInt64(tokenID))

	// Send transaction
	result := SendTransaction(configuration, sender, transactioncode, arguments)
	return result
}

// replace all placeholder addresses with the correct ones
func ReplaceAddressPlaceholders(code []byte, nftAddress string, veoletAddress string, ftAddress string, flowAddress string) []byte {
	code = bytes.ReplaceAll(code, []byte("0xNONFUNGIBLETOKEN"), []byte("0x"+nftAddress))
	code = bytes.ReplaceAll(code, []byte("0xVEOLET"), []byte("0x"+veoletAddress))
	code = bytes.ReplaceAll(code, []byte("0xFUNGIBLETOKEN"), []byte("0x"+ftAddress))
	code = bytes.ReplaceAll(code, []byte("0xFLOW"), []byte("0x"+flowAddress))
	code = bytes.ReplaceAll(code, []byte("\"./NonFungibleToken.cdc\""), []byte("0x"+nftAddress))
	return code
}

func GetServiceAccount(config Configuration) (flow.Address, *flow.AccountKey, crypto.Signer) {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())
	if err != nil {
		panic("Failed to establish connection with Access API")
	}
	defer c.Close()

	serviceAddress := flow.HexToAddress(config.Account.Address)
	serviceAccount, err := c.GetAccountAtLatestBlock(ctx, serviceAddress)
	if err != nil {
		panic("Failed to get Veolet testnet account")
	}
	servicePrivKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, config.Account.Keys)
	if err != nil {
		panic("Failed to decode Veolet private key")
	}
	serviceSigner := crypto.NewInMemorySigner(servicePrivKey, serviceAccount.Keys[0].HashAlgo)
	return serviceAddress, serviceAccount.Keys[0], serviceSigner
}

//makes map of nft {fieldname:value}
func convertCadenceCompositeToMap(fields []cadence.Field, values []cadence.Value) map[string]interface{} {
	m := make(map[string]interface{})
	for i, field := range fields {
		value := values[i]
		if field.Type.ID() == "Address" {
			bytes := value.ToGoValue().([8]byte)
			m[field.Identifier] = flow.BytesToAddress(bytes[:])
		} else {
			m[field.Identifier] = value
		}
	}
	return m
}

//makes map of nfts {nfts:[fieldname:value]}
func ConvertNFTsToMap(input cadence.Value) map[string]interface{} {
	var nft_array []map[string]interface{}
	for _, nft := range input.(cadence.Array).Values {

		//get the field types and the values of a nft
		//pls forgive me this abomination
		fields := nft.(cadence.Optional).Value.(cadence.Optional).Value.(cadence.Resource).ResourceType.Fields
		values := nft.(cadence.Optional).Value.(cadence.Optional).Value.(cadence.Resource).Fields
		nft_struct := convertCadenceCompositeToMap(fields, values)

		nft_array = append(nft_array, nft_struct)

	}
	m := make(map[string]interface{})
	m["NFTs"] = nft_array
	fmt.Println(m)
	return m
}
