package engineapi_el

import (
	"context"
	"sync"

	"github.com/ledgerwatch/erigon-lib/chain"
	"github.com/ledgerwatch/erigon-lib/gointerfaces/txpool"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/kvcache"
	"github.com/ledgerwatch/erigon-lib/state"
	"github.com/ledgerwatch/erigon-lib/wrap"
	"github.com/ledgerwatch/erigon/cmd/rpcdaemon/cli"
	"github.com/ledgerwatch/erigon/cmd/rpcdaemon/cli/httpcfg"
	"github.com/ledgerwatch/erigon/consensus"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/rpc"
	"github.com/ledgerwatch/erigon/turbo/builder"
	"github.com/ledgerwatch/erigon/turbo/jsonrpc"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
	"github.com/ledgerwatch/erigon/turbo/services"
	"github.com/ledgerwatch/erigon/turbo/shards"
	"github.com/ledgerwatch/erigon/turbo/stages/headerdownload"
	"github.com/ledgerwatch/log/v3"
)

type stagesFunc = func(txc wrap.TxContainer, header *types.Header, body *types.RawBody, unwindPoint uint64, headersChain []*types.Header, bodiesChain []*types.RawBody,
	notifications *shards.Notifications) error

type EngineAPI struct {
	hd *headerdownload.HeaderDownload

	ctx    context.Context
	config *chain.Config
	chain  *blockchain
	logger log.Logger

	builder *payloadBuilder

	// builderFunc builder.BlockBuilderFunc
	// builders    map[uint64]*builder.BlockBuilder

	lock sync.Mutex
}

func NewEngineAPI(
	ctx context.Context,
	logger log.Logger,
	config *chain.Config,
	blockReader services.FullBlockReader,
	chainDB kv.RwDB,
	hd *headerdownload.HeaderDownload,
	builderFunc builder.BlockBuilderFunc,
	engine consensus.Engine,
	execFunc stagesFunc,
) *EngineAPI {

	chain := newBlockChain(ctx, blockReader, chainDB, logger, config, engine, execFunc)
	builder := newPayloadBuilder(builderFunc)
	engineAPI := EngineAPI{
		hd:      hd,
		ctx:     ctx,
		config:  config,
		chain:   chain,
		logger:  logger,
		builder: builder,
	}

	return &engineAPI
}

func (api *EngineAPI) Start(httpConfig *httpcfg.HttpCfg, db kv.RoDB, blockReader services.FullBlockReader,
	filters *rpchelper.Filters, stateCache kvcache.Cache, agg *state.AggregatorV3, engineReader consensus.EngineReader,
	eth rpchelper.ApiBackend, txPool txpool.TxpoolClient, mining txpool.MiningClient) {
	base := jsonrpc.NewBaseApi(filters, stateCache, blockReader, agg, httpConfig.WithDatadir, httpConfig.EvmCallTimeout, engineReader, httpConfig.Dirs)

	ethImpl := jsonrpc.NewEthAPI(
		base, db, eth, txPool, mining,
		httpConfig.Gascap,
		httpConfig.ReturnDataLimit,
		httpConfig.AllowUnprotectedTxs,
		httpConfig.MaxGetProofRewindBlockCount,
		api.logger,
	)

	apiList := []rpc.API{
		{
			Namespace: "eth",
			Public:    true,
			Service:   jsonrpc.EthAPI(ethImpl),
			Version:   "1.0",
		}, {
			Namespace: "engine",
			Public:    true,
			Service:   IEngineAPI(api),
			Version:   "1.0",
		}}

	if err := cli.StartRpcServerWithJwtAuthentication(api.ctx, httpConfig, apiList, api.logger); err != nil {
		api.logger.Error(err.Error())
	}
}

var capabilities = []string{
	"engine_forkchoiceUpdatedV1",
	"engine_forkchoiceUpdatedV2",
	"engine_forkchoiceUpdatedV3",
	"engine_newPayloadV1",
	"engine_newPayloadV2",
	"engine_newPayloadV3",
	"engine_getPayloadV1",
	"engine_getPayloadV2",
	"engine_getPayloadV3",
	"engine_exchangeTransitionConfigurationV1",
	"engine_getPayloadBodiesByHashV1",
	"engine_getPayloadBodiesByRangeV1",
}

func (e *EngineAPI) ExchangeCapabilities(fromCl []string) []string {
	missingOurs := compareCapabilities(fromCl, capabilities)
	missingCl := compareCapabilities(capabilities, fromCl)

	if len(missingCl) > 0 || len(missingOurs) > 0 {
		e.logger.Debug("ExchangeCapabilities mismatches", "cl_unsupported", missingCl, "erigon_unsupported", missingOurs)
	}

	return capabilities
}

func compareCapabilities(from []string, to []string) []string {
	result := make([]string, 0)
	for _, f := range from {
		found := false
		for _, t := range to {
			if f == t {
				found = true
				break
			}
		}
		if !found {
			result = append(result, f)
		}
	}

	return result
}
