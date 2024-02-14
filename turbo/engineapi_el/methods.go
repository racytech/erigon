package engineapi_el

import (
	"context"
	"fmt"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/turbo/stages/headerdownload"
)

type payloadVersion byte

var (
	paris    payloadVersion = 0x1
	shanghai payloadVersion = 0x2
	cancun   payloadVersion = 0x3
)

/* ----------------- ForkchoiceUpdated V1, V2, V3 ----------------- */
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

func (api *EngineAPI) forkchoiceUpdated(update *ForkChoiceState, payloadAttributes *PayloadAttributes, version payloadVersion) (*ForkChoiceUpdatedResponse, error) {
	api.lock.Lock()
	defer api.lock.Unlock()

	// TODO: check that we are in PoS chain

	// headBlockHash - block hash of the head of the canonical chain
	// safeBlockHash - the "safe" block hash of the canonical chain under certain synchrony and honesty assumptions. This value MUST be either equal to or an ancestor of headBlockHash
	// finalizedBlockHash - block hash of the most recent finalized block

	// head hash received from CL
	clHeadHash := update.HeadBlockHash

	logPrefix := fmt.Sprintf("[ForkchoiceUpdatedV%v]", version)
	msg := fmt.Sprintf("%v Started processing", logPrefix)
	api._info(msg, []interface{}{"hash", clHeadHash}...)

	// If we are syncing there is no reason to go any ferther
	// return SYNCING status right away
	if api.hd.PosStatus() == headerdownload.Syncing {
		msg := fmt.Sprintf("%v Execution layer is syncing, cannot process forkchoice", logPrefix)
		api._info(msg, []interface{}{"hash", clHeadHash}...)
		return makeFCUresponce(SYNCING, libcommon.Hash{}, nil, nil), nil
	}

	block, err := api.chain.blockByHash(clHeadHash)
	if err != nil {
		return nil, makeError(SERVER_ERROR, err.Error())
	}

	if block == nil {
		msg := fmt.Sprintf("%v Unknown hash received: block is nil", logPrefix)
		api._warn(msg, []interface{}{"hash", clHeadHash}...)
		return &STATUS_SYNCING, nil
	}

	isCanonical, err := api.chain.isCanonicalHash(clHeadHash)
	if err != nil {
		return nil, makeError(SERVER_ERROR, err.Error())
	}
	td, err := api.chain.getTotalDifficulty(block.Hash(), block.NumberU64())
	if err != nil {
		return nil, makeError(SERVER_ERROR, err.Error())
	}
	__assert_true(td != nil, "api.chain.getTotalDifficulty: There is an error somewhere up the call stack: td == nil")

	if td.Cmp(api.config.TerminalTotalDifficulty) >= 0 { // Reached PoS

		if !isCanonical {
			// update block is not canonical
			fmt.Println("---------------------> BLOCK IS NOT CANONICAL")
		}

		currentHead, err := api.chain.currentHead()
		if err != nil {
			return nil, makeError(SERVER_ERROR, err.Error())
		}
		if currentHead.Hash() == clHeadHash {
			// msg := fmt.Sprintf("%v Skipping update to old hash", logPrefix)
			// api._info(msg, []interface{}{"hash", clHeadHash}...)
			// return makeFCUresponce(VALID, clHeadHash, nil, nil), nil
			fmt.Println("---------------------> currentHead.Hash() == clHeadHash")
		}
		if isCanonical {
			// return makeFCUresponce(VALID, update.HeadBlockHash, nil, nil), nil
			fmt.Println("---------------------> BLOCK IS CANONICAL")
		}

		if payloadAttributes != nil {
			timestamp := uint64(payloadAttributes.Timestamp)
			if block.Time() >= timestamp {
				return nil, makeError(INVALID_PAYLOAD_ATTRIBUTES, "Payload timestamp is greater or equal to block time")
			}
			args := core.BlockBuilderParameters{
				ParentHash:            clHeadHash,
				Timestamp:             timestamp,
				PrevRandao:            payloadAttributes.PrevRandao,
				SuggestedFeeRecipient: payloadAttributes.SuggestedFeeRecipient,
			}

			if version >= shanghai {

			}
			if version >= cancun {

			}

			id := api.builder.startPayloadBuild(&args)

			return makeFCUresponce(VALID, clHeadHash, nil, convertPayloadId(id)), nil
		}

	} else {
		// Process PoW blocks here (optional)
		// TODO(racytech): make sure we're not missing anything here

		msg := fmt.Errorf("%v Head hash refers to PoW block", logPrefix)
		api._warn(msg.Error(), []interface{}{"hash", clHeadHash}...)
		str := msg.Error()
		return makeFCUresponce(INVALID, libcommon.Hash{}, &str, nil), nil
	}

	return nil, nil
}

/* ----------------- NewPayload V1, V2, V3 ----------------- */

func (api *EngineAPI) NewPayloadV1(ctx context.Context, payload *ExecutionPayload) (*PayloadStatus, error) {
	return api.newPayload(payload, nil, nil, paris)
}

func (api *EngineAPI) NewPayloadV2(ctx context.Context, payload *ExecutionPayload) (*PayloadStatus, error) {
	return api.newPayload(payload, nil, nil, shanghai)
}

func (api *EngineAPI) NewPayloadV3(ctx context.Context, payload *ExecutionPayload, expectedBlobHashes []libcommon.Hash, parentBeaconBlockRoot *libcommon.Hash) (*PayloadStatus, error) {
	return api.newPayload(payload, expectedBlobHashes, parentBeaconBlockRoot, cancun)
}

func (api *EngineAPI) newPayload(payload *ExecutionPayload, expectedBlobHashes []libcommon.Hash, parentBeaconBlockRoot *libcommon.Hash, version payloadVersion) (*PayloadStatus, error) {

	api.lock.Lock()
	defer api.lock.Unlock()

	logPrefix := fmt.Sprintf("[NewPayloadV%v]", version)
	// msg := fmt.Sprintf("%v Started", logPrefix)

	errMsg := fmt.Sprintf("%v: Reached End", logPrefix)
	return nil, makeError(SERVER_ERROR, errMsg)
}

/* ----------------- GetPayload V1, V2, V3 ----------------- */

func (api *EngineAPI) GetPayloadV1(ctx context.Context, payloadID hexutility.Bytes) (*ExecutionPayload, error) {
	payload, err := api.getPayload(ctx, payloadID, paris)
	if err != nil {
		return nil, err
	}
	return &payload.ExecutionPayload, nil
}

func (api *EngineAPI) GetPayloadV2(ctx context.Context, payloadID hexutility.Bytes) (*GetPayloadResponse, error) {
	return api.getPayload(ctx, payloadID, shanghai)
}

func (api *EngineAPI) GetPayloadV3(ctx context.Context, payloadID hexutility.Bytes) (*GetPayloadResponse, error) {
	return api.getPayload(ctx, payloadID, cancun)
}

func (api *EngineAPI) getPayload(ctx context.Context, payloadID hexutility.Bytes, version payloadVersion) (*GetPayloadResponse, error) {
	api.lock.Lock()
	defer api.lock.Unlock()

	id := payloadIDtoUint64(payloadID)

	logPrefix := fmt.Sprintf("[GetPayloadV%v]", version)
	msg := fmt.Sprintf("%v Started", logPrefix)
	api._info(msg, []interface{}{"payloadID", id}...)

	res, err := api.builder.extractPayload(id)
	if err != nil {
		return nil, err
	}

	msg = fmt.Sprintf("%v Finished", logPrefix)
	api._info(msg, []interface{}{"parent_hash", res.ExecutionPayload.ParentHash}...)

	return res, nil
}

/* ----------------- ExchangeTransitionConfigurationV1 ----------------- */

func (api *EngineAPI) ExchangeTransitionConfigurationV1(ctx context.Context, transitionConfiguration *TransitionConfiguration) (*TransitionConfiguration, error) {
	return &TransitionConfiguration{}, makeError(SERVER_ERROR, "ExchangeTransitionConfigurationV1: Unimplemented")
}

/* ----------------- GetPayloadBodiesByHashV1 ----------------- */

func (api *EngineAPI) GetPayloadBodiesByHashV1(ctx context.Context, hashes []libcommon.Hash) ([]*ExecutionPayloadBodyV1, error) {
	return nil, makeError(SERVER_ERROR, "GetPayloadBodiesByHashV1: Unimplemented")
}

/* ----------------- GetPayloadBodiesByRangeV1 ----------------- */

func (api *EngineAPI) GetPayloadBodiesByRangeV1(ctx context.Context, start, count hexutil.Uint64) ([]*ExecutionPayloadBodyV1, error) {
	return nil, makeError(SERVER_ERROR, "GetPayloadBodiesByRangeV1: Unimplemented")
}

/* ----------------- HELPERS ----------------- */

func __assert_true(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
