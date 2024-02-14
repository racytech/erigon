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

func (chain *blockchain) readForkchoiceHead() (libcommon.Hash, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceHead: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()
	hash := rawdb.ReadForkchoiceHead(tx)

	return hash, nil
}

func (chain *blockchain) readForkchoiceFinalized() (libcommon.Hash, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceFinalized: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	hash := rawdb.ReadForkchoiceFinalized(tx)

	return hash, nil
}

func (chain *blockchain) readForkchoiceSafe() (libcommon.Hash, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceSafe: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	hash := rawdb.ReadForkchoiceSafe(tx)

	return hash, nil
}

func (chain *blockchain) blockByHash(hash libcommon.Hash) (*types.Block, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: blockByHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	if chain.blockReader == nil {
		return nil, fmt.Errorf("EnginAPI: blockByHash: blockreader is nil")
	}

	if chain.blockReader != nil {
		return chain.blockReader.BlockByHash(chain.ctx, tx, hash)
	}

	blockNumber := rawdb.ReadHeaderNumber(tx, hash)
	if blockNumber == nil {
		return nil, fmt.Errorf("rawdb.ReadHeaderNumber: Could not retrieve block number: hash=%v", hash)
	}
	block := rawdb.ReadBlock(tx, hash, *blockNumber)
	return block, nil
}

func (chain *blockchain) headerByHash(hash libcommon.Hash) (*types.Header, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: headerByHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	if chain.blockReader != nil {
		return chain.blockReader.HeaderByHash(chain.ctx, tx, hash)
	}

	return rawdb.ReadHeaderByHash(tx, hash)
}

func (chain *blockchain) headerNumber(hash libcommon.Hash) (*uint64, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: headerNumber: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	return rawdb.ReadHeaderNumber(tx, hash), nil
}

func (chain *blockchain) getTotalDifficulty(hash libcommon.Hash, number uint64) (*big.Int, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: getTD: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	return rawdb.ReadTd(tx, hash, number)
}

func (chain *blockchain) canonicalHash(hash libcommon.Hash) (libcommon.Hash, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return libcommon.Hash{}, fmt.Errorf("EnginAPI: canonicalHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	blockNumber := rawdb.ReadHeaderNumber(tx, hash)
	if blockNumber == nil {
		return libcommon.Hash{}, nil
	}

	if chain.blockReader != nil {
		return chain.blockReader.CanonicalHash(chain.ctx, tx, *blockNumber)
	}

	return rawdb.ReadCanonicalHash(tx, *blockNumber)
}

func (chain *blockchain) isCanonicalHash(hash libcommon.Hash) (bool, error) {
	_hash, err := chain.canonicalHash(hash)
	if err != nil {
		return false, nil
	}

	return _hash == hash, nil
}

func (chain *blockchain) currentHead() (*types.Header, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: currentHeader: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	hash := rawdb.ReadHeadHeaderHash(tx)
	if chain.blockReader != nil {
		fmt.Println("USING BLOCK READER")
		number := rawdb.ReadHeaderNumber(tx, hash)
		return chain.blockReader.Header(chain.ctx, tx, hash, *number)
	}

	return rawdb.ReadHeaderByHash(tx, hash)
}
