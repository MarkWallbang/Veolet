package lib

import (
	"bytes"
	"context"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

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
