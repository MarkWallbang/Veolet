package lib

// [1]
import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

// [2]
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
