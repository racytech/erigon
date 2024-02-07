package engineapi_el

import (
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/common"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	"github.com/ledgerwatch/erigon-lib/gointerfaces"
	"github.com/ledgerwatch/erigon-lib/gointerfaces/execution"
)

type payloadVersion byte

var (
	paris    payloadVersion = 0x1
	shanghai payloadVersion = 0x2
	cancun   payloadVersion = 0x3
)

/* ----------------- ForkchoiceUpdated V1, V2, V3 ----------------- */
// ForkchoiceUpdatedV1(context.Context, *ForkChoiceState, *PayloadAttributes) (*ForkChoiceUpdatedResponse, error)
func (api *EngineAPI) ForkchoiceUpdatedV1(ctx context.Context, update *ForkChoiceState, attributes *PayloadAttributes) (*ForkChoiceUpdatedResponse, error) {
	if attributes != nil {
		if attributes.Withdrawals != nil || attributes.ParentBeaconBlockRoot != nil {
			// withdrawals and beacon root not supported in V1
		}
		if api.config.IsShanghai(attributes.Timestamp.Uint64()) {
			// forkChoiceUpdateV1 called post-shanghai
		}
	}
	return api.forkchoiceUpdated(update, attributes, paris)
}

func (api *EngineAPI) ForkchoiceUpdatedV2(ctx context.Context, update *ForkChoiceState, attributes *PayloadAttributes) (*ForkChoiceUpdatedResponse, error) {
	if attributes != nil {
		time := attributes.Timestamp.Uint64()
		if api.config.IsShanghai(time) && attributes.Withdrawals == nil {
			// missing withdrawals list
		}

		if !api.config.IsShanghai(time) && attributes.Withdrawals != nil {
			// withdrawals before shanghai
		}

		if attributes.ParentBeaconBlockRoot != nil {
			// unexpected beacon root
		}
	}
	return api.forkchoiceUpdated(update, attributes, shanghai)
}

func (api *EngineAPI) ForkchoiceUpdatedV3(ctx context.Context, update *ForkChoiceState, attributes *PayloadAttributes) (*ForkChoiceUpdatedResponse, error) {
	return api.forkchoiceUpdated(update, attributes, cancun)
}

func (api *EngineAPI) forkchoiceUpdated(update *ForkChoiceState, attributes *PayloadAttributes, version payloadVersion) (*ForkChoiceUpdatedResponse, error) {
	api.lock.Lock()
	defer api.lock.Unlock()

	// TODO: check that we are in PoS chain

	clHeadHash := update.HeadBlockHash

	logPrefix := fmt.Sprintf("[ForkchoiceUpdatedV%v]", version)
	msg := fmt.Sprintf("%v Started processing", logPrefix)
	api._log(msg, []interface{}{"head_hash", clHeadHash})

	// headHash, err := api.chain.readForkchoiceHead()
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// finalizedHash, err := api.chain.readForkchoiceFinalized()
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// safeHash, err := api.chain.readForkchoiceSafe()
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// header, err := api.chain.headerByHash(update.HeadBlockHash)
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// var parent *types.Header

	// check if we've seen this block and marked it as bad
	bad, lastValidHash := api.hd.IsBadHeaderPoS(clHeadHash)
	if bad {
		errMsg := fmt.Errorf("links to previously rejected block")
		return makeFCUresponce(INVALID, lastValidHash, errMsg, nil), nil
	}

	block, err := api.chain.blockByHash(clHeadHash)
	fmt.Println("block.hash == clHeadHash: ", block.Hash() == clHeadHash)
	if err != nil {
		return nil, makeError(SERVER_ERROR, err.Error())
	}

	if block == nil {
		msg := fmt.Sprintf("%v Request for unknown hash", logPrefix)
		api._warn(msg, []interface{}{"head_hash", clHeadHash})

		return &STATUS_SYNCING, nil
	}

	td, err := api.chain.getTotalDifficulty(clHeadHash, block.NumberU64())
	if err != nil {
		return nil, makeError(SERVER_ERROR, err.Error())
	}

	if td != nil && td.Cmp(api.config.TerminalTotalDifficulty) < 0 {
		msg := fmt.Errorf("%v Beacon Chain request before TTD", logPrefix)
		api._warn(msg.Error(), "hash", clHeadHash)
		return makeFCUresponce(INVALID, libcommon.Hash{}, msg, nil), nil
	}

	// no need to start block build process
	if attributes == nil {
		// TODO:
		return makeFCUresponce(VALID, libcommon.Hash{}, nil, nil), nil
	}

	timestamp := uint64(attributes.Timestamp)
	if block.Time() >= timestamp {
		return nil, &INVALID_PAYLOAD_ATTRIBUTES_ERR
	}

	req := &execution.AssembleBlockRequest{
		ParentHash:            gointerfaces.ConvertHashToH256(clHeadHash),
		Timestamp:             timestamp,
		PrevRandao:            gointerfaces.ConvertHashToH256(attributes.PrevRandao),
		SuggestedFeeRecipient: gointerfaces.ConvertAddressToH160(attributes.SuggestedFeeRecipient),
	}

	if version >= shanghai {
		req.Withdrawals = ConvertWithdrawalsToRpc(attributes.Withdrawals)
	}

	if version >= cancun {
		req.ParentBeaconBlockRoot = gointerfaces.ConvertHashToH256(*attributes.ParentBeaconBlockRoot)
	}

	// canonicalHash, err := api.chain.canonicalHash(clHeadHash)
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// if canonicalHash != clHeadHash {

	// }

	fmt.Println("HEADER: ", block.NumberU64())

	return makeFCUresponce(VALID, lastValidHash, nil, nil), nil
}

/* ----------------- NewPayload V1, V2, V3 ----------------- */

func (api *EngineAPI) NewPayloadV1(ctx context.Context, payload *ExecutionPayload) (*PayloadStatus, error) {
	return api.newPayload(payload, nil, nil, paris)
}

func (api *EngineAPI) NewPayloadV2(ctx context.Context, payload *ExecutionPayload) (*PayloadStatus, error) {
	return api.newPayload(payload, nil, nil, shanghai)
}

func (api *EngineAPI) NewPayloadV3(ctx context.Context, payload *ExecutionPayload, expectedBlobHashes []common.Hash, parentBeaconBlockRoot *common.Hash) (*PayloadStatus, error) {
	return api.newPayload(payload, expectedBlobHashes, parentBeaconBlockRoot, cancun)
}

func (api *EngineAPI) newPayload(payload *ExecutionPayload, expectedBlobHashes []common.Hash, parentBeaconBlockRoot *common.Hash, version payloadVersion) (*PayloadStatus, error) {
	logPrefix := fmt.Sprintf("NewPayloadV%v", version)

	errMsg := fmt.Sprintf("%v: Reached End", logPrefix)
	return nil, makeError(SERVER_ERROR, errMsg)
}

/* ----------------- GetPayload V1, V2, V3 ----------------- */

func (api *EngineAPI) GetPayloadV1(ctx context.Context, payloadID hexutility.Bytes) (*ExecutionPayload, error) {

	return nil, makeError(SERVER_ERROR, "GetPayloadV1: Unimplemented")
	// return ExecutionPayload{}, nil
}

func (api *EngineAPI) GetPayloadV2(ctx context.Context, payloadID hexutility.Bytes) (*GetPayloadResponse, error) {
	return api.getPayload(ctx, payloadID, shanghai)
}

func (api *EngineAPI) GetPayloadV3(ctx context.Context, payloadID hexutility.Bytes) (*GetPayloadResponse, error) {
	return api.getPayload(ctx, payloadID, cancun)
}

func (api *EngineAPI) getPayload(ctx context.Context, payloadID hexutility.Bytes, version payloadVersion) (*GetPayloadResponse, error) {
	logPrefix := fmt.Sprintf("GetPayloadV%v", version)

	errMsg := fmt.Sprintf("%v: Reached End", logPrefix)
	return nil, makeError(SERVER_ERROR, errMsg)
}

/* ----------------- ExchangeTransitionConfigurationV1 ----------------- */

func (api *EngineAPI) ExchangeTransitionConfigurationV1(ctx context.Context, transitionConfiguration *TransitionConfiguration) (*TransitionConfiguration, error) {
	return &TransitionConfiguration{}, makeError(SERVER_ERROR, "ExchangeTransitionConfigurationV1: Unimplemented")
}

/* ----------------- GetPayloadBodiesByHashV1 ----------------- */

func (api *EngineAPI) GetPayloadBodiesByHashV1(ctx context.Context, hashes []common.Hash) ([]*ExecutionPayloadBodyV1, error) {
	return nil, makeError(SERVER_ERROR, "GetPayloadBodiesByHashV1: Unimplemented")
}

/* ----------------- GetPayloadBodiesByRangeV1 ----------------- */

func (api *EngineAPI) GetPayloadBodiesByRangeV1(ctx context.Context, start, count hexutil.Uint64) ([]*ExecutionPayloadBodyV1, error) {
	return nil, makeError(SERVER_ERROR, "GetPayloadBodiesByRangeV1: Unimplemented")
}
