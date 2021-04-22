package lib

import (
	"context"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func FetchContracts(config Configuration, address flow.Address) map[string][]byte {
	ctx := context.Background()
	c, err := client.New(config.Network.Host, grpc.WithInsecure())

	account, err := c.GetAccount(ctx, address)

	if err != nil {
		panic("failed to fetch account")
	}
	return account.Contracts
}
