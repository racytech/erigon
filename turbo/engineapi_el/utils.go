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
	"github.com/ledgerwatch/log/v3"
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

	var withdrawalsHash *libcommon.Hash
	if payload.Withdrawals != nil {
		wh := types.DeriveSha(types.Withdrawals(payload.Withdrawals))
		withdrawalsHash = &wh
	}

	header := types.Header{
		ParentHash:            payload.ParentHash,
		UncleHash:             types.EmptyUncleHash,
		Coinbase:              payload.FeeRecipient,
		Root:                  payload.StateRoot,
		TxHash:                types.DeriveSha(types.BinaryTransactions(txs)),
		ReceiptHash:           payload.ReceiptsRoot,
		Bloom:                 types.BytesToBloom(payload.LogsBloom),
		Difficulty:            merge.ProofOfStakeDifficulty,
		Number:                new(big.Int).SetUint64(payload.BlockNumber),
		GasLimit:              payload.GasLimit,
		GasUsed:               payload.GasUsed,
		Time:                  payload.Timestamp,
		BaseFee:               payload.BaseFeePerGas,
		Extra:                 payload.ExtraData,
		MixDigest:             payload.PrevRandao,
		WithdrawalsHash:       withdrawalsHash,
		ExcessBlobGas:         payload.ExcessBlobGas,
		BlobGasUsed:           payload.BlobGasUsed,
		ParentBeaconBlockRoot: beaconRoot,
	}

	block := types.NewBlockWithHeader(&header).WithTransactions(transactions).WithWithdrawals(payload.Withdrawals)
	// cmpHeaders(block.Header(), &header, payload)
	if block.Hash() != payload.BlockHash {
		return nil, fmt.Errorf("block hashes does not match: expected: %v, got :%v", payload.BlockHash, block.Hash())
	}
	return block, nil
}

func cmpHeaders(a, b *types.Header, c *ExecutionPayload) {
	fmt.Println("ParentHash", a.ParentHash == b.ParentHash)
	fmt.Println("UncleHash", a.UncleHash == b.UncleHash)
	fmt.Println("Coinbase", a.Coinbase == b.Coinbase)
	fmt.Println("Root", a.Root == b.Root)
	fmt.Println("TxHash", a.TxHash == b.TxHash)
	fmt.Printf("ReceiptHash a: %v, b: %v, c: %v\n", a.ReceiptHash, b.ReceiptHash, c.ReceiptsRoot)
	fmt.Println("Bloom", a.Bloom == b.Bloom)
	fmt.Printf("Difficulty a: %v, b: %v\n", a.Difficulty.Uint64(), b.Difficulty.Uint64())
	fmt.Printf("Number a: %v, b: %v\n", a.Number.Uint64(), b.Number.Uint64())
	fmt.Println("GasLimi", a.GasLimit == b.GasLimit)
	fmt.Println("GasUsed", a.GasUsed == b.GasUsed)
	fmt.Println("Time", a.Time == b.Time)
	fmt.Printf("BaseFee a: %v, b :%v\n", a.BaseFee.Uint64(), b.BaseFee.Uint64())
	// fmt.Println(a.Extra == b.Extra)
	fmt.Println("MixDigest", a.MixDigest == b.MixDigest)
	fmt.Println("WithdrawalsHash", a.WithdrawalsHash == b.WithdrawalsHash)
	fmt.Println("ExcessBlobGas a: %v, b: %v, original: %v", a.ExcessBlobGas, b.ExcessBlobGas, c.ExcessBlobGas)
	fmt.Printf("BlobGasUsed a: %v, b: %v, original: %v", a.BlobGasUsed, b.BlobGasUsed, c.BlobGasUsed)
	fmt.Println("ParentBeaconBlockRoot ", a.ParentBeaconBlockRoot == b.ParentBeaconBlockRoot)
}

/* Logging shortcuts */

func log_info(logger log.Logger, msg string, ctx ...interface{}) {
	logger.Info(msg, ctx...)
}

func log_warn(logger log.Logger, msg string, ctx ...interface{}) {
	logger.Warn(msg, ctx...)
}

func log_error(logger log.Logger, msg string, ctx ...interface{}) {
	logger.Error(msg, ctx...)
}

func log_debug(logger log.Logger, msg string, ctx ...interface{}) {
	logger.Debug(msg, ctx...)
}

func log_trace(logger log.Logger, msg string, ctx ...interface{}) {
	logger.Trace(msg, ctx...)
}
