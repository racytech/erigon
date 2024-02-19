package engineapi_el

import (
	"math/big"

	"github.com/ledgerwatch/erigon-lib/common/hexutil"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	"github.com/ledgerwatch/erigon/core/types"
)

//go:generate gencodec -type ExecutionPayload -field-override executionPayloadMarshaling -out gen_ep.go

// ExecutionPayload represents an execution payload (aka block)
type ExecutionPayload struct {
	ParentHash    libcommon.Hash      `json:"parentHash"    gencodec:"required"`
	FeeRecipient  libcommon.Address   `json:"feeRecipient"  gencodec:"required"`
	StateRoot     libcommon.Hash      `json:"stateRoot"     gencodec:"required"`
	ReceiptsRoot  libcommon.Hash      `json:"receiptsRoot"  gencodec:"required"`
	LogsBloom     []byte              `json:"logsBloom"     gencodec:"required"`
	PrevRandao    libcommon.Hash      `json:"prevRandao"    gencodec:"required"`
	BlockNumber   uint64              `json:"blockNumber"   gencodec:"required"`
	GasLimit      uint64              `json:"gasLimit"      gencodec:"required"`
	GasUsed       uint64              `json:"gasUsed"       gencodec:"required"`
	Timestamp     uint64              `json:"timestamp"     gencodec:"required"`
	ExtraData     []byte              `json:"extraData"     gencodec:"required"`
	BaseFeePerGas *big.Int            `json:"baseFeePerGas" gencodec:"required"`
	BlockHash     libcommon.Hash      `json:"blockHash"     gencodec:"required"`
	Transactions  [][]byte            `json:"transactions"  gencodec:"required"`
	Withdrawals   []*types.Withdrawal `json:"withdrawals"`
	BlobGasUsed   *uint64             `json:"blobGasUsed"`
	ExcessBlobGas *uint64             `json:"excessBlobGas"`
}

type executionPayloadMarshaling struct {
	LogsBloom     hexutility.Bytes
	PrevRandao    libcommon.Hash
	BlockNumber   hexutil.Uint64
	GasLimit      hexutil.Uint64
	GasUsed       hexutil.Uint64
	Timestamp     hexutil.Uint64
	ExtraData     hexutility.Bytes
	BaseFeePerGas *hexutil.Big
	Transactions  []hexutility.Bytes
}

// PayloadAttributes represent the attributes required to start assembling a payload
type ForkChoiceState struct {
	HeadBlockHash      libcommon.Hash `json:"headBlockHash"             gencodec:"required"`
	SafeBlockHash      libcommon.Hash `json:"safeBlockHash"             gencodec:"required"`
	FinalizedBlockHash libcommon.Hash `json:"finalizedBlockHash"        gencodec:"required"`
}

// PayloadAttributes represent the attributes required to start assembling a payload
type PayloadAttributes struct {
	Timestamp             hexutil.Uint64      `json:"timestamp"             gencodec:"required"`
	PrevRandao            libcommon.Hash      `json:"prevRandao"            gencodec:"required"`
	SuggestedFeeRecipient libcommon.Address   `json:"suggestedFeeRecipient" gencodec:"required"`
	Withdrawals           []*types.Withdrawal `json:"withdrawals"`
	ParentBeaconBlockRoot *libcommon.Hash     `json:"parentBeaconBlockRoot"`
}

// TransitionConfiguration represents the correct configurations of the CL and the EL
type TransitionConfiguration struct {
	TerminalTotalDifficulty *hexutil.Big   `json:"terminalTotalDifficulty" gencodec:"required"`
	TerminalBlockHash       libcommon.Hash `json:"terminalBlockHash"       gencodec:"required"`
	TerminalBlockNumber     *hexutil.Big   `json:"terminalBlockNumber"     gencodec:"required"`
}

// BlobsBundleV1 holds the blobs of an execution payload
type BlobsBundleV1 struct {
	Commitments []hexutility.Bytes `json:"commitments" gencodec:"required"`
	Proofs      []hexutility.Bytes `json:"proofs"      gencodec:"required"`
	Blobs       []hexutility.Bytes `json:"blobs"       gencodec:"required"`
}

type ExecutionPayloadBodyV1 struct {
	Transactions []hexutility.Bytes  `json:"transactions" gencodec:"required"`
	Withdrawals  []*types.Withdrawal `json:"withdrawals"  gencodec:"required"`
}

type PayloadStatus struct {
	Status          string         `json:"status" gencodec:"required"`
	ValidationError *string        `json:"validationError"`
	LatestValidHash libcommon.Hash `json:"latestValidHash"`
	CriticalError   error
}

type ForkChoiceUpdatedResponse struct {
	PayloadID     *hexutility.Bytes `json:"payloadId"` // We need to reformat the uint64 so this makes more sense.
	PayloadStatus PayloadStatus     `json:"payloadStatus"`
}

//go:generate gencodec -type GetPayloadResponse -field-override getPayloadResponseMarshaling -out gen_gpr.go

type GetPayloadResponse struct {
	ExecutionPayload      ExecutionPayload `json:"executionPayload" gencodec:"required"`
	BlockValue            *big.Int         `json:"blockValue"`
	BlobsBundle           *BlobsBundleV1   `json:"blobsBundle"`
	ShouldOverrideBuilder bool             `json:"shouldOverrideBuilder"`
}

type getPayloadResponseMarshaling struct {
	BlockValue *hexutil.Big
}
