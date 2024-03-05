package engineapi_el

import (
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/turbo/services"
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

func newEngineAPIMemDB(chainDB kv.RwDB, blockReader services.FullBlockReader) *engineAPIMemDB {
	// 1. get N_BLOCKS blocks from chainDB (if any || genesis only)
	// 2. insert them into newly created memDB so we have some reference when CL asks
	// 		- this also required for payload validation, block execution
	return nil
}

// initializes memDB. it expects up to N_BLOCKS blocks from persistant DB (or 1 genesis, in case syncing from scratch)
// so it can operate independently of main chainDB
