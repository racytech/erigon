package engineapi_el

import (
	"context"
	"fmt"
	"math/big"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/turbo/services"
)

type blockchain struct {
	ctx         context.Context
	blockReader services.FullBlockReader
	chainDB     kv.RwDB
}

func newBlockChain(ctx context.Context, blockReader services.FullBlockReader, chainDB kv.RwDB) *blockchain {
	return &blockchain{ctx: ctx, blockReader: blockReader, chainDB: chainDB}
}

func (bc *blockchain) readForkchoiceHead() (libcommon.Hash, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceHead: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()
	hash := rawdb.ReadForkchoiceHead(tx)

	return hash, nil
}

func (bc *blockchain) readForkchoiceFinalized() (libcommon.Hash, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceFinalized: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	hash := rawdb.ReadForkchoiceFinalized(tx)

	return hash, nil
}

func (bc *blockchain) readForkchoiceSafe() (libcommon.Hash, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceSafe: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	hash := rawdb.ReadForkchoiceSafe(tx)

	return hash, nil
}

func (bc *blockchain) blockByHash(hash libcommon.Hash) (*types.Block, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: blockByHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	if bc.blockReader == nil {
		return nil, fmt.Errorf("EnginAPI: blockByHash: blockreader is nil")
	}

	return bc.blockReader.BlockByHash(bc.ctx, tx, hash)
}

func (bc *blockchain) headerByHash(hash libcommon.Hash) (*types.Header, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: headerByHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	if bc.blockReader != nil {
		return bc.blockReader.HeaderByHash(bc.ctx, tx, hash)
	}

	return rawdb.ReadHeaderByHash(tx, hash)
}

func (bc *blockchain) headerNumber(hash libcommon.Hash) (*uint64, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: headerNumber: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	return rawdb.ReadHeaderNumber(tx, hash), nil
}

func (bc *blockchain) getTotalDifficulty(hash libcommon.Hash, number uint64) (*big.Int, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: getTD: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	return rawdb.ReadTd(tx, hash, number)
}

func (bc *blockchain) canonicalHash(hash libcommon.Hash) (libcommon.Hash, error) {
	tx, err := bc.chainDB.BeginRo(bc.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: isCanonicalHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	blockNumber := rawdb.ReadHeaderNumber(tx, hash)
	if blockNumber == nil {
		return libcommon.Hash{}, nil
	}

	if bc.blockReader != nil {
		return bc.blockReader.CanonicalHash(bc.ctx, tx, *blockNumber)
	}

	return rawdb.ReadCanonicalHash(tx, *blockNumber)
}
