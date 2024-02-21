package engineapi_el

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/chain"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/ledgerwatch/erigon-lib/kv/membatch"
	"github.com/ledgerwatch/erigon/consensus"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/core/vm"
	"github.com/ledgerwatch/erigon/core/vm/evmtypes"
	"github.com/ledgerwatch/erigon/eth/calltracer"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/turbo/services"
	"github.com/ledgerwatch/erigon/turbo/trie"
	"github.com/ledgerwatch/log/v3"
)

// type execFunc func(wrap.TxContainer, *core.ChainPack, uint64) error

type blockchain struct {
	ctx         context.Context
	blockReader services.FullBlockReader
	chainDB     kv.RwDB
	logger      log.Logger
	config      *chain.Config
	memChain    []*types.Block // TODO(racytech): see how much max bytes takes one block
	memChainIdx int

	engine consensus.Engine
}

func newBlockChain(
	ctx context.Context,
	blockReader services.FullBlockReader,
	chainDB kv.RwDB,
	logger log.Logger,
	cfg *chain.Config,
	engine consensus.Engine,
) *blockchain {
	return &blockchain{
		ctx:         ctx,
		blockReader: blockReader,
		chainDB:     chainDB,
		logger:      logger,
		config:      cfg,
		engine:      engine,
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

var VMcfg = &vm.Config{
	Tracer:     calltracer.NewCallTracer(),
	ReadOnly:   true,
	NoReceipts: true,
}

// func newVmCfg() *vm.Config {
// 	return &vm.Config{
// 		Tracer:     calltracer.NewCallTracer(),
// 		ReadOnly:   true,
// 		NoReceipts: true,
// 	}
// }

func (chain *blockchain) insertBlock(block, parent *types.Block) error {
	tx, err := chain.chainDB.BeginRw(chain.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stateRootBegin, err := calculateStateRoot(tx)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	// chain.execBlock(txc, block.Header(), block.RawBody(), 0)

	// MakePreState()

	// we need to execute block here and compare root hashes

	quit := chain.ctx.Done()

	// state is stored through ethdb batches
	batch := membatch.NewHashBatch(tx, quit, "/tmp/state_rw", chain.logger)
	// avoids stacking defers within the loop
	defer func() {
		batch.Close()
	}()

	stateR := state.NewPlainStateReader(batch)
	stateW := state.NewPlainStateWriter(batch, nil, block.NumberU64())

	getHeader := func(hash libcommon.Hash, number uint64) *types.Header {
		h, _ := chain.blockReader.Header(context.Background(), tx, hash, number)
		return h
	}

	blockHashFunc := core.GetHashFn(block.Header(), getHeader)

	ibs := state.New(stateR)
	header := block.Header()

	gp := new(core.GasPool)
	gp.AddGas(block.GasLimit()).AddBlobGas(chain.config.GetMaxBlobGasPerBlock())

	var rejectedTxs []*core.RejectedTx
	includedTxs := make(types.Transactions, 0, block.Transactions().Len())

	blockContext := core.NewEVMBlockContext(header, blockHashFunc, chain.engine, nil)
	evm := vm.NewEVM(blockContext, evmtypes.TxContext{}, ibs, chain.config, *VMcfg)
	rules := evm.ChainRules()

	chain.logger.Info("Block execution started", "block", block.NumberU64())
	for i, tx := range block.Transactions() {

		msg, err := tx.AsMessage(*types.MakeSigner(chain.config, header.Number.Uint64(), header.Time), header.BaseFee, rules)
		if err != nil {
			rejectedTxs = append(rejectedTxs, &core.RejectedTx{Index: i, Err: err.Error()})
		}

		if _, err = core.ApplyMessage(evm, msg, gp, false /* refunds */, false /* gasBailout */); err != nil {
			chain.logger.Error("Error applying message", "error", err.Error())
			continue
		}

		if err = ibs.FinalizeTx(rules, stateW); err != nil {
			chain.logger.Error("Error finalizing transaction", "error", err.Error())
			continue
		}

		includedTxs = append(includedTxs, tx)
	}

	if err = ibs.CommitBlock(rules, stateW); err != nil {
		chain.logger.Error("Error commiting block", "error", err.Error())
		return err
	}

	txRoot := types.DeriveSha(includedTxs)
	if txRoot != block.TxHash() {
		return fmt.Errorf("transaction root mismatch: expected: %v, got: %v", block.TxHash(), txRoot)
	}
	fmt.Printf("expected txRoot: %v\n, got: %v\n", block.TxHash(), txRoot)

	stateRoot, err := calculateStateRoot(tx)
	if err != nil {
		return fmt.Errorf("error calculating state root: %w", err)
	}
	fmt.Printf("expected stateRoot: %v\n, got: %v\n, state root begin: %v", block.Root(), stateRoot, stateRootBegin)

	if len(rejectedTxs) > 0 {
		chain.logger.Warn("Some transactions were rejected", "count", len(rejectedTxs))
		for _, rtx := range rejectedTxs {
			chain.logger.Error("Transaction was rejected", "tx_number", rtx.Index, "error", rtx.Err)
		}
	}

	return nil
}

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
