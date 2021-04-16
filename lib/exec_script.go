package lib

// [1]
import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

// [2]
func ExecuteScript(node string, script []byte, args []cadence.Value) cadence.Value {
	ctx := context.Background()
	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	// [3]
	result, err := c.ExecuteScriptAtLatestBlock(ctx, script, args)
	if err != nil {
		panic(err)
	}

	return result
}