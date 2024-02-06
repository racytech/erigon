package engineapi_el

import (
	"context"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
)

// CL <-> EL communication interface
type IEngineAPI interface {
	NewPayloadV1(context.Context, ExecutionPayload) (PayloadStatus, error)
	NewPayloadV2(context.Context, ExecutionPayload) (PayloadStatus, error)
	NewPayloadV3(ctx context.Context, executionPayload ExecutionPayload, expectedBlobHashes []common.Hash, parentBeaconBlockRoot *common.Hash) (PayloadStatus, error)
	ForkchoiceUpdatedV1(ctx context.Context, forkChoiceState ForkChoiceState, payloadAttributes PayloadAttributes) (ForkChoiceUpdatedResponse, error)
	ForkchoiceUpdatedV2(ctx context.Context, forkChoiceState ForkChoiceState, payloadAttributes PayloadAttributes) (ForkChoiceUpdatedResponse, error)
	ForkchoiceUpdatedV3(ctx context.Context, forkChoiceState ForkChoiceState, payloadAttributes PayloadAttributes) (ForkChoiceUpdatedResponse, error)
	GetPayloadV1(ctx context.Context, payloadID hexutility.Bytes) (ExecutionPayload, error)
	GetPayloadV2(ctx context.Context, payloadID hexutility.Bytes) (GetPayloadResponse, error)
	GetPayloadV3(ctx context.Context, payloadID hexutility.Bytes) (GetPayloadResponse, error)
	ExchangeTransitionConfigurationV1(ctx context.Context, transitionConfiguration TransitionConfiguration) (TransitionConfiguration, error)
	GetPayloadBodiesByHashV1(ctx context.Context, hashes []common.Hash) ([]ExecutionPayloadBodyV1, error)
	GetPayloadBodiesByRangeV1(ctx context.Context, start, count hexutil.Uint64) ([]ExecutionPayloadBodyV1, error)
}
