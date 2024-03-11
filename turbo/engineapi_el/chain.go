package engineapi_el

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	mrand "math/rand"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/chain"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/ledgerwatch/erigon-lib/wrap"
	"github.com/ledgerwatch/erigon/consensus"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/eth/stagedsync"
	"github.com/ledgerwatch/erigon/turbo/services"
	"github.com/ledgerwatch/erigon/turbo/trie"
	"github.com/ledgerwatch/log/v3"
)

// type execFunc func(wrap.TxContainer, *core.ChainPack, uint64) error

type blockchain struct {
	ctx context.Context
	// blockReader services.FullBlockReader
	chainDB    kv.RwDB
	logger     log.Logger
	config     *chain.Config
	engine     consensus.Engine
	stagedSync *stagedsync.Sync
	rand       *mrand.Rand
}

func newBlockChain(
	ctx context.Context,
	blockReader services.FullBlockReader,
	memDB kv.RwDB,
	chainDB kv.RwDB,
	logger log.Logger,
	cfg *chain.Config,
	engine consensus.Engine,
	stagedSync *stagedsync.Sync,
) *blockchain {
	return &blockchain{
		ctx: ctx,
		// blockReader: blockReader,
		chainDB:    chainDB,
		logger:     logger,
		config:     cfg,
		engine:     engine,
		stagedSync: stagedSync,
	}
}

// func (chain *blockchain) readForkchoiceHead() (libcommon.Hash, error) {
// 	tx, err := chain.chainDB.BeginRo(chain.ctx)
// 	if err != nil {
// 		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceHead: Failed to BeginRo: %s", err)
// 	}
// 	defer tx.Rollback()
// 	hash := rawdb.ReadForkchoiceHead(tx)

// 	return hash, nil
// }

// func (chain *blockchain) readForkchoiceFinalized() (libcommon.Hash, error) {
// 	tx, err := chain.chainDB.BeginRo(chain.ctx)
// 	if err != nil {
// 		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceFinalized: Failed to BeginRo: %s", err)
// 	}
// 	defer tx.Rollback()

// 	hash := rawdb.ReadForkchoiceFinalized(tx)

// 	return hash, nil
// }

// unused
// func (chain *blockchain) readForkchoiceSafe() (libcommon.Hash, error) {
// 	tx, err := chain.chainDB.BeginRo(chain.ctx)
// 	if err != nil {
// 		return libcommon.Hash{}, fmt.Errorf("EnginAPI: readForkchoiceSafe: Failed to BeginRo: %s", err)
// 	}
// 	defer tx.Rollback()

// 	hash := rawdb.ReadForkchoiceSafe(tx)

// 	return hash, nil
// }

func (chain *blockchain) blockByHash(hash libcommon.Hash) (*types.Block, error) {
	tx, err := chain.chainDB.BeginRo(chain.ctx)
	if err != nil {
		return nil, fmt.Errorf("EnginAPI: blockByHash: Failed to BeginRo: %s", err)
	}
	defer tx.Rollback()

	// if chain.blockReader == nil {
	// 	return nil, fmt.Errorf("EnginAPI: blockByHash: blockreader is nil")
	// }

	// if chain.blockReader != nil {
	// 	return chain.blockReader.BlockByHash(chain.ctx, tx, hash)
	// }

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

	// if chain.blockReader != nil {
	// 	return chain.blockReader.HeaderByHash(chain.ctx, tx, hash)
	// }

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
	// if blockNumber == nil {
	// 	return libcommon.Hash{}, nil
	// }

	// if chain.blockReader != nil {
	// 	return chain.blockReader.CanonicalHash(chain.ctx, tx, *blockNumber)
	// }

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
	// if chain.blockReader != nil {
	// 	fmt.Println("USING BLOCK READER")
	// 	number := rawdb.ReadHeaderNumber(tx, hash)
	// 	return chain.blockReader.Header(chain.ctx, tx, hash, *number)
	// }

	return rawdb.ReadHeaderByHash(tx, hash)
}

func (chain *blockchain) writeBlockToDB(block, parent *types.Block) error {
	tx, err := chain.chainDB.BeginRw(chain.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	parentTd, err := rawdb.ReadTd(tx, parent.Hash(), parent.NumberU64())
	if err != nil {
		return fmt.Errorf("error getting parent TD: %v", err)
	}
	// Sum TDs.
	td := parentTd.Add(parentTd, block.Difficulty())
	if err := rawdb.WriteTd(tx, block.Hash(), block.NumberU64(), td); err != nil {
		return fmt.Errorf("error writeBlockToDB - WriteTd: %s", err)
	}

	if err := rawdb.WriteBlock(tx, block); err != nil {
		return fmt.Errorf("error writeBlockToDB - WriteBlock: %s", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error could not commit tx: %s", err)
	}
	return nil
}

func (chain *blockchain) insertBlock(block *types.Block) error {

	currentHead, err := chain.currentHead()
	if err != nil {
		return fmt.Errorf("error getting currentHead: %v", err)
	}

	parent, err := chain.blockByHash(block.ParentHash())
	if err != nil {
		return fmt.Errorf("error getting parent: %v", err)
	}

	reorg, err := chain.reorgNeeded(currentHead, block.Header())
	if err != nil {
		log.Warn("reorgNeeded", "error", err)
	}
	fmt.Println("Reorg needed: ", reorg)

	if reorg {

	}

	err = chain.writeBlockToDB(block, parent)
	if err != nil {
		return err
	}

	tx, err := chain.chainDB.BeginRw(chain.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var txc wrap.TxContainer
	txc.Tx = tx
	result, err := chain.stagedSync.Run(chain.chainDB, txc, false)
	if err != nil {
		return fmt.Errorf("error stagedSync Run: %v", err)
	}

	fmt.Println("RESULT: ", result)
	return nil
}

// ReorgNeeded returns whether the reorg should be applied
// based on the given external header and local canonical chain.
// In the td mode, the new head is chosen if the corresponding
// total difficulty is higher. In the extern mode, the trusted
// header is always selected as the head.
func (chain *blockchain) reorgNeeded(current *types.Header, extern *types.Header) (bool, error) {

	localTD, err := chain.getTotalDifficulty(current.Hash(), current.Number.Uint64())
	if err != nil {
		return false, fmt.Errorf("error getting totalDifficulty for current head: %v", err)
	}
	externTd, err := chain.getTotalDifficulty(extern.Hash(), extern.Number.Uint64())
	if err != nil {
		return false, fmt.Errorf("error getting totalDifficulty for extern head: %v", err)
	}

	if localTD == nil || externTd == nil {
		return false, errors.New("missing td")
	}
	// Accept the new header as the chain head if the transition
	// is already triggered. We assume all the headers after the
	// transition come from the trusted consensus layer.
	if ttd := chain.config.TerminalTotalDifficulty; ttd != nil && ttd.Cmp(externTd) <= 0 {
		return true, nil
	}

	// If the total difficulty is higher than our known, add it to the canonical chain
	if diff := externTd.Cmp(localTD); diff > 0 {
		return true, nil
	} else if diff < 0 {
		return false, nil
	}
	// Local and external difficulty is identical.
	// Second clause in the if statement reduces the vulnerability to selfish mining.
	// Please refer to http://www.cs.cornell.edu/~ie53/publications/btcProcFC.pdf
	reorg := false
	externNum, localNum := extern.Number.Uint64(), current.Number.Uint64()
	if externNum < localNum {
		reorg = true
	} else if externNum == localNum {
		var currentPreserve, externPreserve bool

		reorg = !currentPreserve && (externPreserve || chain.rand.Float64() < 0.5)
	}
	return reorg, nil
}

// // execute block
// vmConfig := &vm.Config{}

// stateWriter := state.NewPlainStateWriter(tx, nil, block.NumberU64()+1)
// stateReader := state.NewPlainStateReader(tx)
// ibs := state.New(stateReader)
// header := block.Header()

// gp := new(core.GasPool)
// gp.AddGas(block.GasLimit()).AddBlobGas(chain.config.GetMaxBlobGasPerBlock())

// getHeader := func(hash common.Hash, number uint64) *types.Header {
// 	h, _ := chain.blockReader.Header(context.Background(), tx, hash, number)
// 	return h
// }
// getHashFn := core.GetHashFn(block.Header(), getHeader)

// blockContext := core.NewEVMBlockContext(header, getHashFn, chain.engine, nil)
// evm := vm.NewEVM(blockContext, evmtypes.TxContext{}, ibs, chain.config, *vmConfig)

// rules := chain.config.Rules(block.NumberU64(), block.Time())
// for i, txn := range block.Transactions() {
// 	msg, err := txn.AsMessage(*types.MakeSigner(chain.config, block.NumberU64(), block.Time()), block.BaseFee(), rules)
// 	if err != nil {
// 		return fmt.Errorf("Error txn.AsMessage: %v", err)
// 	}

// 	txContext := core.NewEVMTxContext(msg)

// 	evm.Reset(txContext, ibs)

// 	_, err = core.ApplyMessage(evm, msg, gp, true, false)
// 	if err != nil {
// 		return fmt.Errorf("error ApplyMessage: %v", err)
// 	}
// 	// Update the state with pending changes
// 	if err = ibs.FinalizeTx(rules, stateWriter); err != nil {
// 		return fmt.Errorf("error FinalizeTx: %v", err)
// 	}

// 	fmt.Printf("TX %v: Done\n", i)
// }

// if err = ibs.CommitBlock(rules, stateWriter); err != nil {
// 	return fmt.Errorf("error ibs.CommitBlock: %v", err)
// }

// ibs.SetTxContext(txn.Hash(), block.Hash(), i)

// receipt, _, err := core.ApplyTransaction(chain.config, getHashFn, chain.engine, nil, gp, ibs, noop, header, txn, usedGas, usedBlobGas, *vmConfig)
// if err != nil {
// 	rejectedTxs = append(rejectedTxs, &core.RejectedTx{Index: i, Err: err.Error()})
// } else {
// 	receipts = append(receipts, receipt)
// }

// fmt.Printf("RECEIPT #%v: type = %v\n", i, receipt.Type)

// var txc wrap.TxContainer
// txc.Tx = tx

// var notes = &shards.Notifications{
// 	Events:      shards.NewEvents(),
// 	Accumulator: shards.NewAccumulator(),
// }

// chain.logger.Debug("Spawning Stages")
// err = chain.execFunc(txc, block.Header(), block.RawBody(), 0, nil, nil, notes)

func MakePreState(rules *chain.Rules, tx kv.RwTx, accounts types.GenesisAlloc, blockNr uint64) (*state.IntraBlockState, error) {
	r := state.NewPlainStateReader(tx)
	statedb := state.New(r)
	for addr, a := range accounts {
		statedb.SetCode(addr, a.Code)
		statedb.SetNonce(addr, a.Nonce)
		balance := uint256.NewInt(0)
		if a.Balance != nil {
			balance, _ = uint256.FromBig(a.Balance)
		}
		statedb.SetBalance(addr, balance)
		for k, v := range a.Storage {
			key := k
			val := uint256.NewInt(0).SetBytes(v.Bytes())
			statedb.SetState(addr, &key, *val)
		}

		if len(a.Code) > 0 || len(a.Storage) > 0 {
			statedb.SetIncarnation(addr, state.FirstContractIncarnation)

			var b [8]byte
			binary.BigEndian.PutUint64(b[:], state.FirstContractIncarnation)
			if err := tx.Put(kv.IncarnationMap, addr[:], b[:]); err != nil {
				return nil, err
			}
		}
	}

	var w state.StateWriter
	if ethconfig.EnableHistoryV4InTest {
		panic("implement me")
	} else {
		w = state.NewPlainStateWriter(tx, nil, blockNr+1)
	}
	// Commit and re-open to start with a clean state.
	if err := statedb.FinalizeTx(rules, w); err != nil {
		return nil, err
	}
	if err := statedb.CommitBlock(rules, w); err != nil {
		return nil, err
	}
	return statedb, nil
}

func calculateStateRoot(tx kv.RwTx) (*libcommon.Hash, error) {
	// Generate hashed state
	c, err := tx.RwCursor(kv.PlainState)
	if err != nil {
		return nil, err
	}
	h := libcommon.NewHasher()
	defer libcommon.ReturnHasherToPool(h)
	for k, v, err := c.First(); k != nil; k, v, err = c.Next() {
		if err != nil {
			return nil, fmt.Errorf("interate over plain state: %w", err)
		}
		var newK []byte
		if len(k) == length.Addr {
			newK = make([]byte, length.Hash)
		} else {
			newK = make([]byte, length.Hash*2+length.Incarnation)
		}
		h.Sha.Reset()
		//nolint:errcheck
		h.Sha.Write(k[:length.Addr])
		//nolint:errcheck
		h.Sha.Read(newK[:length.Hash])
		if len(k) > length.Addr {
			copy(newK[length.Hash:], k[length.Addr:length.Addr+length.Incarnation])
			h.Sha.Reset()
			//nolint:errcheck
			h.Sha.Write(k[length.Addr+length.Incarnation:])
			//nolint:errcheck
			h.Sha.Read(newK[length.Hash+length.Incarnation:])
			if err = tx.Put(kv.HashedStorage, newK, libcommon.CopyBytes(v)); err != nil {
				return nil, fmt.Errorf("insert hashed key: %w", err)
			}
		} else {
			if err = tx.Put(kv.HashedAccounts, newK, libcommon.CopyBytes(v)); err != nil {
				return nil, fmt.Errorf("insert hashed key: %w", err)
			}
		}
	}
	c.Close()
	root, err := trie.CalcRoot("", tx)
	if err != nil {
		return nil, err
	}

	return &root, nil
}

// newMemoryNodeDB creates a new in-memory node database without a persistent backend.
func newMemoryDB(ctx context.Context, logger log.Logger, tmpDir string) kv.RwDB {
	return mdbx.NewMDBX(log.New()).InMem(tmpDir).MustOpen()
}

// var assembleBlockPOS = func(param *core.BlockBuilderParameters, interrupt *int32) (*types.BlockWithReceipts, error) {
// 	miningStatePos := stagedsync.NewProposingState(&config.Miner)
// 	miningStatePos.MiningConfig.Etherbase = param.SuggestedFeeRecipient
// 	proposingSync := stagedsync.New(
// 		config.Sync,
// 		stagedsync.MiningStages(backend.sentryCtx,
// 			stagedsync.StageMiningCreateBlockCfg(backend.chainDB, miningStatePos, *backend.chainConfig, backend.engine, backend.txPoolDB, param, tmpdir, backend.blockReader),
// 			stagedsync.StageBorHeimdallCfg(backend.chainDB, snapDb, miningStatePos, *backend.chainConfig, heimdallClient, backend.blockReader, nil, nil, nil, recents, signatures),
// 			stagedsync.StageMiningExecCfg(backend.chainDB, miningStatePos, backend.notifications.Events, *backend.chainConfig, backend.engine, &vm.Config{}, tmpdir, interrupt, param.PayloadId, backend.txPool, backend.txPoolDB, blockReader),
// 			stagedsync.StageHashStateCfg(backend.chainDB, dirs, config.HistoryV3),
// 			stagedsync.StageTrieCfg(backend.chainDB, false, true, true, tmpdir, blockReader, nil, config.HistoryV3, backend.agg),
// 			stagedsync.StageMiningFinishCfg(backend.chainDB, *backend.chainConfig, backend.engine, miningStatePos, backend.miningSealingQuit, backend.blockReader, latestBlockBuiltStore),
// 		), stagedsync.MiningUnwindOrder, stagedsync.MiningPruneOrder,
// 		logger)
// 	// We start the mining step
// 	if err := stages2.MiningStep(ctx, backend.chainDB, proposingSync, tmpdir, logger); err != nil {
// 		return nil, err
// 	}
// 	block := <-miningStatePos.MiningResultPOSCh
// 	return block, nil
// }
