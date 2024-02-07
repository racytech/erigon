package engineapi_el

import (
	"context"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
)

// CL <-> EL communication interface
type IEngineAPI interface {
	NewPayloadV1(context.Context, *ExecutionPayload) (*PayloadStatus, error)
	NewPayloadV2(context.Context, *ExecutionPayload) (*PayloadStatus, error)
	NewPayloadV3(context.Context, *ExecutionPayload, []common.Hash, *common.Hash) (*PayloadStatus, error)

	ForkchoiceUpdatedV1(context.Context, *ForkChoiceState, *PayloadAttributes) (*ForkChoiceUpdatedResponse, error)
	ForkchoiceUpdatedV2(context.Context, *ForkChoiceState, *PayloadAttributes) (*ForkChoiceUpdatedResponse, error)
	ForkchoiceUpdatedV3(context.Context, *ForkChoiceState, *PayloadAttributes) (*ForkChoiceUpdatedResponse, error)

	GetPayloadV1(context.Context, hexutility.Bytes) (*ExecutionPayload, error)
	GetPayloadV2(context.Context, hexutility.Bytes) (*GetPayloadResponse, error)
	GetPayloadV3(context.Context, hexutility.Bytes) (*GetPayloadResponse, error)

	ExchangeTransitionConfigurationV1(context.Context, *TransitionConfiguration) (*TransitionConfiguration, error)

	GetPayloadBodiesByHashV1(context.Context, []common.Hash) ([]*ExecutionPayloadBodyV1, error)

	GetPayloadBodiesByRangeV1(context.Context, hexutil.Uint64, hexutil.Uint64) ([]*ExecutionPayloadBodyV1, error)
}
