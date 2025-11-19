package node

import (
	"context"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type mockCometRPC struct {
	client.CometRPC
	earliestBlockHeight int64
}

func (m mockCometRPC) Status(ctx context.Context) (*coretypes.ResultStatus, error) {
	return &coretypes.ResultStatus{
		SyncInfo: coretypes.SyncInfo{
			EarliestBlockHeight: m.earliestBlockHeight,
		},
	}, nil
}

func TestServiceServer_Config(t *testing.T) {
	defaultCfg := config.DefaultConfig()
	defaultCfg.PruningKeepRecent = "2000"
	defaultCfg.PruningInterval = "10"
	defaultCfg.HaltHeight = 100
	svr := NewQueryServer(client.Context{}, *defaultCfg)
	ctx := sdk.Context{}.WithMinGasPrices(sdk.NewDecCoins(sdk.NewInt64DecCoin("stake", 15)))

	resp, err := svr.Config(ctx, &ConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, ctx.MinGasPrices().String(), resp.MinimumGasPrice)
	require.Equal(t, defaultCfg.PruningKeepRecent, resp.PruningKeepRecent)
	require.Equal(t, defaultCfg.PruningInterval, resp.PruningInterval)
	require.Equal(t, defaultCfg.HaltHeight, resp.HaltHeight)
}

func TestServiceServer_Status(t *testing.T) {
	blockTime := time.Now().UTC()
	blockHeight := int64(100)
	appHash := []byte("apphash")
	validatorHash := []byte("validatorhash")

	testCases := []struct {
		name                string
		clientCtx           client.Context
		expectedEarliest    uint64
		expectedHeight      uint64
	}{
		{
			name: "with comet client - pruned node",
			clientCtx: client.Context{
				Client: mockCometRPC{earliestBlockHeight: 50},
			},
			expectedEarliest: 50,
			expectedHeight:   100,
		},
		{
			name: "with comet client - full node",
			clientCtx: client.Context{
				Client: mockCometRPC{earliestBlockHeight: 1},
			},
			expectedEarliest: 1,
			expectedHeight:   100,
		},
		{
			name:             "without comet client",
			clientCtx:        client.Context{},
			expectedEarliest: 0,
			expectedHeight:   100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaultCfg := config.DefaultConfig()
			svr := NewQueryServer(tc.clientCtx, *defaultCfg)

			ctx := sdk.Context{}.
				WithBlockHeader(cmtproto.Header{
					Height:             blockHeight,
					Time:               blockTime,
					AppHash:            appHash,
					NextValidatorsHash: validatorHash,
				})

			resp, err := svr.Status(ctx, &StatusRequest{})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.expectedEarliest, resp.EarliestStoreHeight)
			require.Equal(t, tc.expectedHeight, resp.Height)
			require.Equal(t, blockTime, *resp.Timestamp)
			require.Equal(t, appHash, resp.AppHash)
			require.Equal(t, validatorHash, resp.ValidatorHash)
		})
	}
}
