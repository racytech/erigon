package engineapi_el

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/turbo/builder"
)

// type payloadBuildParams struct {
// 	PayloadId             uint64
// 	ParentHash            libcommon.Hash
// 	Timestamp             uint64
// 	PrevRandao            libcommon.Hash
// 	SuggestedFeeRecipient libcommon.Address
// 	Withdrawals           []*types.Withdrawal // added in Shapella (EIP-4895)
// 	ParentBeaconBlockRoot *libcommon.Hash     // added in Dencun (EIP-4788)
// }

const MAX_PAYLOAD_CYCLE = 1024

type payloadBuilder struct {
	payloadID   uint64
	payloadsMap map[uint64]*builder.BlockBuilder // TODO: make it array of 1024 cycled items 0..1023 -> 1024..2047 -> 2048..4097 -> and so on (n%1024)

	// payloadArr []int64

	builderFunc builder.BlockBuilderFunc

	lock sync.Mutex
}

func newPayloadBuilder(builderFunc builder.BlockBuilderFunc) *payloadBuilder {
	return &payloadBuilder{
		builderFunc: builderFunc,
		payloadsMap: make(map[uint64]*builder.BlockBuilder),
		// payloadArr:  make([]int64, MAX_PAYLOAD_CYCLE),
	}
}

func (pb *payloadBuilder) startPayloadBuild(params *core.BlockBuilderParameters) uint64 {
	pb.lock.Lock()
	defer pb.lock.Unlock()

	pb.payloadsMap[pb.payloadID] = builder.NewBlockBuilder(pb.builderFunc, params)

	id := pb.payloadID
	pb.payloadID++

	return id
}

func (pb *payloadBuilder) extractPayload(id uint64) (*GetPayloadResponse, error) {
	pb.lock.Lock()
	defer pb.lock.Unlock()

	builder, ok := pb.payloadsMap[id]
	if !ok {
		return nil, fmt.Errorf("can not find payload by given id=%d", id)
	}

	blockWithReceipts, err := builder.Stop()
	if err != nil {
		return nil, err
	}
	block := blockWithReceipts.Block

	encodedTransactions, err := types.MarshalTransactionsBinary(block.Transactions())
	if err != nil {
		return nil, err
	}

	transactions := make([][]byte, len(encodedTransactions))
	copy(transactions, encodedTransactions)

	payload := ExecutionPayload{
		BlockHash:     block.Hash(),
		ParentHash:    block.ParentHash(),
		FeeRecipient:  block.Coinbase(),
		StateRoot:     block.Root(),
		BlockNumber:   block.NumberU64(),
		GasLimit:      block.GasLimit(),
		GasUsed:       block.GasUsed(),
		BaseFeePerGas: block.BaseFee(),
		Timestamp:     block.Time(),
		ReceiptsRoot:  block.ReceiptHash(),
		LogsBloom:     block.Bloom().Bytes(),
		Transactions:  transactions,
		PrevRandao:    block.MixDigest(),
		ExtraData:     block.Extra(),
		Withdrawals:   block.Withdrawals(),
		BlobGasUsed:   block.Header().BlobGasUsed,
		ExcessBlobGas: block.Header().ExcessBlobGas,
	}

	baseFee := block.BaseFee()
	blockValue := blockValue(blockWithReceipts, baseFee)

	blobsBundle := &BlobsBundleV1{}
	for i, tx := range block.Transactions() {
		if tx.Type() != types.BlobTxType {
			continue
		}
		blobTx, ok := tx.(*types.BlobTxWrapper)
		if !ok {
			return nil, fmt.Errorf("expected blob transaction to be type BlobTxWrapper, got: %T", blobTx)
		}
		versionedHashes, commitments, proofs, blobs := blobTx.GetBlobHashes(), blobTx.Commitments, blobTx.Proofs, blobTx.Blobs
		lenCheck := len(versionedHashes)
		if lenCheck != len(commitments) || lenCheck != len(proofs) || lenCheck != len(blobs) {
			return nil, fmt.Errorf("tx %d in block %s has inconsistent commitments (%d) / proofs (%d) / blobs (%d) / "+
				"versioned hashes (%d)", i, block.Hash(), len(commitments), len(proofs), len(blobs), lenCheck)
		}
		for _, commitment := range commitments {
			c := types.KZGCommitment{}
			copy(c[:], commitment[:])
			blobsBundle.Commitments = append(blobsBundle.Commitments, c[:])
		}
		for _, proof := range proofs {
			p := types.KZGProof{}
			copy(p[:], proof[:])
			blobsBundle.Proofs = append(blobsBundle.Proofs, p[:])
		}
		for _, blob := range blobs {
			b := types.Blob{}
			copy(b[:], blob[:])
			blobsBundle.Blobs = append(blobsBundle.Blobs, b[:])
		}
	}

	return &GetPayloadResponse{
		ExecutionPayload: payload,
		BlockValue:       blockValue,
		BlobsBundle:      blobsBundle,
	}, nil
}

// The expected value to be received by the feeRecipient in wei
func blockValue(br *types.BlockWithReceipts, baseFee *big.Int) *big.Int {
	blockValue := uint256.NewInt(0)
	txs := br.Block.Transactions()
	uint256BaseFee := uint256.MustFromBig(baseFee)
	for i := range txs {
		gas := new(uint256.Int).SetUint64(br.Receipts[i].GasUsed)
		effectiveTip := txs[i].GetEffectiveGasTip(uint256BaseFee)
		txValue := new(uint256.Int).Mul(gas, effectiveTip)
		blockValue.Add(blockValue, txValue)
	}
	return blockValue.ToBig()
}
