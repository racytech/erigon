package engineapi_el

import (
	"context"
	"fmt"

	"github.com/c2h5oh/datasize"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/log/v3"
)

const (
	N_BLOCKS   = 128 // does it have to be this big?
	BATCH_SIZE = 256 // batch of blocks to be added into persistant DB (chainDB)
)

type engineAPIMemDB struct {
	// chain of hashes, blocks that could be safely (after making sure this chain not going to reorg, midified or else)
	// added into chainDB later when BATCH_SIZE is reached (see safeBlockHash in FCU)
	safeChain []*libcommon.Hash // h0 <- h1 <- h2 <- ... <- hn

	//
	memDB kv.RwDB

	// erigon's persistant chain DB.
	// we are going to insert blocks into it, when safe chain advenced to ~BATCH_SIZE number of blocks
	// the problem is that chainDB is always behind of a head for some blocks
	chainDB kv.RwDB
}

func __newMemDB() kv.RwDB {
	return mdbx.NewMDBX(log.New()).
		InMem("memDB_blocks").
		Label(kv.InMem).
		GrowthStep(16 * datasize.MB).
		MustOpen()
}

func newEngineAPIMemDB(chainDB kv.RwDB, ctx context.Context) (kv.RwDB, error) {
	// 1. get up to N_BLOCKS blocks from chainDB (if any or genesis only)
	// 2. insert them into newly created memDB so we have some reference when CL asks
	// 		- this also required for payload validation, block execution

	tx, err := chainDB.BeginRo(ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: currentHeader: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	hash := rawdb.ReadHeadBlockHash(tx)

	blockNumber := rawdb.ReadHeaderNumber(tx, hash)

	var blocks []*types.Header

	for blockN := int64(*blockNumber); blockN >= 0 || blockN > int64(*blockNumber)-N_BLOCKS; blockN -= 1 {
		header := rawdb.ReadHeaderByNumber(tx, uint64(blockN))
		if header == nil {
			break
		}
		blocks = append(blocks, header)
	}

	memDB := __newMemDB()
	mtx, err := memDB.BeginRw(ctx)
	if err != nil {
		return nil, fmt.Errorf("newEngineAPIMemDB: memDB BeginRw failed: %s", err)
	}
	// mtx.Rollback()

	for i := len(blocks) - 1; i >= 0; i-- {
		block := rawdb.ReadBlock(tx, blocks[i].Hash(), blocks[i].Number.Uint64())

		__assert_true(mtx != nil, "mtx == nil")
		__assert_true(block != nil, "block == nil")

		if err := rawdb.WriteBlock(mtx, block); err != nil {
			return nil, err
		}

		td, err := rawdb.ReadTd(tx, blocks[i].Hash(), blocks[i].Number.Uint64())
		if err != nil {
			return nil, err
		}
		fmt.Println("TD: ", td.Uint64())
		if err = rawdb.WriteTd(mtx, blocks[i].Hash(), blocks[i].Number.Uint64(), td); err != nil {
			return nil, err
		}
	}

	if err := mtx.Commit(); err != nil {
		return nil, err
	}
	fmt.Println("DONE MEM DB")
	return memDB, nil
}

// initializes memDB. it expects up to N_BLOCKS blocks from persistant DB (or 1 genesis, in case syncing from scratch)
// so it can operate independently of main chainDB
