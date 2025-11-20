package node

import (
	context "context"
	"regexp"
	"strconv"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterNodeService registers the node gRPC service on the provided gRPC router.
func RegisterNodeService(clientCtx client.Context, server gogogrpc.Server, cfg config.Config) {
	RegisterServiceServer(server, NewQueryServer(clientCtx, cfg))
}

// RegisterGRPCGatewayRoutes mounts the node gRPC service's GRPC-gateway routes
// on the given mux object.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	_ = RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientConn))
}

var _ ServiceServer = queryServer{}

type queryServer struct {
	clientCtx client.Context
	cfg       config.Config
}

func NewQueryServer(clientCtx client.Context, cfg config.Config) ServiceServer {
	return queryServer{
		clientCtx: clientCtx,
		cfg:       cfg,
	}
}

func (s queryServer) Config(ctx context.Context, _ *ConfigRequest) (*ConfigResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &ConfigResponse{
		MinimumGasPrice:   sdkCtx.MinGasPrices().String(),
		PruningKeepRecent: s.cfg.PruningKeepRecent,
		PruningInterval:   s.cfg.PruningInterval,
		HaltHeight:        s.cfg.HaltHeight,
	}, nil
}

func (s queryServer) Status(ctx context.Context, _ *StatusRequest) (*StatusResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	blockTime := sdkCtx.BlockTime()

	var earliestHeight uint64
	if s.clientCtx.Client != nil {
		if node, err := s.clientCtx.GetNode(); err == nil {
			// Try to query block at height 1 to determine earliest available height
			height := int64(1)
			_, err := node.Block(ctx, &height)
			if err == nil {
				// Block at height 1 exists, so earliest height is 1
				earliestHeight = 1
			} else {
				// Parse error message to extract lowest available height
				// Expected format: "height X is not available, lowest height is Y"
				re := regexp.MustCompile(`lowest height is (\d+)`)
				matches := re.FindStringSubmatch(err.Error())
				if len(matches) > 1 {
					if parsed, parseErr := strconv.ParseUint(matches[1], 10, 64); parseErr == nil {
						earliestHeight = parsed
					}
				}
			}
		}
	}

	return &StatusResponse{
		EarliestStoreHeight: earliestHeight,
		Height:              uint64(sdkCtx.BlockHeight()),
		Timestamp:           &blockTime,
		AppHash:             sdkCtx.BlockHeader().AppHash,
		ValidatorHash:       sdkCtx.BlockHeader().NextValidatorsHash,
	}, nil
}
