package engineapi_el

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ledgerwatch/erigon-lib/common"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	"github.com/ledgerwatch/erigon/consensus/merge"
	"github.com/ledgerwatch/erigon/core/types"
)

func makeFCUresponce(status string, lastValidHash libcommon.Hash, errMsg *string, payloadID *hexutility.Bytes) *ForkChoiceUpdatedResponse {

	return &ForkChoiceUpdatedResponse{
		PayloadStatus: PayloadStatus{
			Status:          status,
			LatestValidHash: lastValidHash,
			ValidationError: errMsg,
		},
		PayloadID: payloadID,
	}
}

func convertPayloadId(payloadId uint64) *hexutility.Bytes {
	encodedPayloadId := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedPayloadId, payloadId)
	ret := hexutility.Bytes(encodedPayloadId)
	return &ret
}

func payloadIDtoUint64(payloadID hexutility.Bytes) uint64 {
	return binary.BigEndian.Uint64(payloadID)
}

func payloadToBlock(payload *ExecutionPayload, versionedHashes []common.Hash, beaconRoot *common.Hash) (*types.Block, error) {

	if len(payload.ExtraData) > 32 {
		return nil, fmt.Errorf("invalid extradata length: %v", len(payload.ExtraData))
	}
	if len(payload.LogsBloom) != 256 {
		return nil, fmt.Errorf("invalid logsBloom length: %v", len(payload.LogsBloom))
	}
	// Check that baseFeePerGas is not negative or too big
	if payload.BaseFeePerGas != nil && (payload.BaseFeePerGas.Sign() == -1 || payload.BaseFeePerGas.BitLen() > 256) {
		return nil, fmt.Errorf("invalid baseFeePerGas: %v", payload.BaseFeePerGas)
	}

	txs := [][]byte{}
	txs = append(txs, payload.Transactions...)

	transactions, err := types.DecodeTransactions(txs)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %v", err)
	}

	var blobHashes []common.Hash
	for _, tx := range transactions {
		blobHashes = append(blobHashes, tx.GetBlobHashes()...)
	}
	if len(blobHashes) != len(versionedHashes) {
		return nil, fmt.Errorf("invalid number of versionedHashes: %v blobHashes: %v", versionedHashes, blobHashes)
	}
	for i := 0; i < len(blobHashes); i++ {
		if blobHashes[i] != versionedHashes[i] {
			return nil, fmt.Errorf("invalid versionedHash at %v: %v blobHashes: %v", i, versionedHashes, blobHashes)
		}
	}

	var bloom types.Bloom
	copy(bloom[:], payload.LogsBloom)

	var withdrawalsHash *libcommon.Hash
	if payload.Withdrawals != nil {
		wh := types.DeriveSha(types.Withdrawals(payload.Withdrawals))
		withdrawalsHash = &wh
	}

	header := types.Header{
		ParentHash:            payload.ParentHash,
		Coinbase:              payload.FeeRecipient,
		Root:                  payload.StateRoot,
		Bloom:                 bloom,
		BaseFee:               payload.BaseFeePerGas,
		Extra:                 payload.ExtraData,
		Number:                new(big.Int).SetUint64(payload.BlockNumber),
		GasUsed:               payload.GasUsed,
		GasLimit:              payload.GasLimit,
		Time:                  payload.Timestamp,
		MixDigest:             payload.PrevRandao,
		UncleHash:             types.EmptyUncleHash,
		Difficulty:            merge.ProofOfStakeDifficulty,
		Nonce:                 merge.ProofOfStakeNonce,
		ReceiptHash:           payload.ReceiptsRoot,
		TxHash:                types.DeriveSha(types.BinaryTransactions(txs)),
		WithdrawalsHash:       withdrawalsHash,
		ExcessBlobGas:         payload.ExcessBlobGas,
		BlobGasUsed:           payload.BlobGasUsed,
		ParentBeaconBlockRoot: beaconRoot,
	}

	block := types.NewBlock(&header, transactions, nil, nil, payload.Withdrawals)
	if block.Hash() != payload.BlockHash {
		return nil, fmt.Errorf("block hashes does not match: expected: %v, got :%v", payload.BlockHash, block.Hash())
	}
	return nil, nil
}
