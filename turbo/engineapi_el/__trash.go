	// headHash, err := api.chain.readForkchoiceHead()
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// finalizedHash, err := api.chain.readForkchoiceFinalized()
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// safeHash, err := api.chain.readForkchoiceSafe()
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// header, err := api.chain.headerByHash(update.HeadBlockHash)
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// var parent *types.Header

	// TODO:
	// check if we've seen this block and marked it as bad
	// bad, lastValidHash := api.hd.IsBadHeaderPoS(clHeadHash)
	// if bad {
	// 	errMsg := fmt.Errorf("links to previously rejected block")
	// 	return makeFCUresponce(INVALID, lastValidHash, errMsg, nil), nil
	// }

	// block, err := api.chain.blockByHash(clHeadHash)
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// if block == nil {
	// 	msg := fmt.Sprintf("%v Request for unknown hash", logPrefix)
	// 	api._warn(msg, []interface{}{"head_hash", clHeadHash}...)

	// 	return &STATUS_SYNCING, nil
	// }

	// td, err := api.chain.getTotalDifficulty(clHeadHash, block.NumberU64())
	// if err != nil {
	// 	return nil, makeError(SERVER_ERROR, err.Error())
	// }

	// if td != nil && td.Cmp(api.config.TerminalTotalDifficulty) < 0 {
	// 	msg := fmt.Errorf("%v Beacon Chain request before TTD", logPrefix)
	// 	api._warn(msg.Error(), []interface{}{"head_hash", clHeadHash}...)
	// 	return makeFCUresponce(INVALID, libcommon.Hash{}, msg, nil), nil
	// }

	// // no need to start block build process
	// if attributes == nil {
	// 	// TODO:
	// 	return makeFCUresponce(VALID, libcommon.Hash{}, nil, nil), nil
	// }

	// timestamp := uint64(attributes.Timestamp)
	// if block.Time() >= timestamp {
	// 	return nil, &INVALID_PAYLOAD_ATTRIBUTES_ERR
	// }

	// req := &execution.AssembleBlockRequest{
	// 	ParentHash:            gointerfaces.ConvertHashToH256(clHeadHash),
	// 	Timestamp:             timestamp,
	// 	PrevRandao:            gointerfaces.ConvertHashToH256(attributes.PrevRandao),
	// 	SuggestedFeeRecipient: gointerfaces.ConvertAddressToH160(attributes.SuggestedFeeRecipient),
	// }

	// if version >= shanghai {
	// 	req.Withdrawals = ConvertWithdrawalsToRpc(attributes.Withdrawals)
	// }

	// if version >= cancun {
	// 	req.ParentBeaconBlockRoot = gointerfaces.ConvertHashToH256(*attributes.ParentBeaconBlockRoot)
	// }

	// // canonicalHash, err := api.chain.canonicalHash(clHeadHash)
	// // if err != nil {
	// // 	return nil, makeError(SERVER_ERROR, err.Error())
	// // }

	// // if canonicalHash != clHeadHash {

	// // }

	// fmt.Println("HEADER: ", block.NumberU64())

	// return makeFCUresponce(VALID, lastValidHash, nil, nil), nil

	// stateRootBegin, err := calculateStateRoot(tx)
	// if err != nil {
	// 	fmt.Println("ERROR: ", err)
	// }
	// // chain.execBlock(txc, block.Header(), block.RawBody(), 0)

	// // MakePreState()

	// // we need to execute block here and compare root hashes

	// quit := chain.ctx.Done()

	// // state is stored through ethdb batches
	// batch := membatch.NewHashBatch(tx, quit, "/tmp/state_rw", chain.logger)
	// // avoids stacking defers within the loop
	// defer func() {
	// 	batch.Close()
	// }()

	// stateR := state.NewPlainStateReader(batch)
	// stateW := state.NewPlainStateWriter(batch, nil, block.NumberU64())

	// getHeader := func(hash libcommon.Hash, number uint64) *types.Header {
	// 	h, _ := chain.blockReader.Header(context.Background(), tx, hash, number)
	// 	return h
	// }

	// blockHashFunc := core.GetHashFn(block.Header(), getHeader)

	// ibs := state.New(stateR)
	// header := block.Header()

	// gp := new(core.GasPool)
	// gp.AddGas(block.GasLimit()).AddBlobGas(chain.config.GetMaxBlobGasPerBlock())

	// var rejectedTxs []*core.RejectedTx
	// includedTxs := make(types.Transactions, 0, block.Transactions().Len())

	// blockContext := core.NewEVMBlockContext(header, blockHashFunc, chain.engine, nil)
	// evm := vm.NewEVM(blockContext, evmtypes.TxContext{}, ibs, chain.config, *VMcfg)
	// rules := evm.ChainRules()

	// chain.logger.Info("Block execution started", "block", block.NumberU64())
	// for i, tx := range block.Transactions() {

	// 	msg, err := tx.AsMessage(*types.MakeSigner(chain.config, header.Number.Uint64(), header.Time), header.BaseFee, rules)
	// 	if err != nil {
	// 		rejectedTxs = append(rejectedTxs, &core.RejectedTx{Index: i, Err: err.Error()})
	// 	}

	// 	if _, err = core.ApplyMessage(evm, msg, gp, false /* refunds */, false /* gasBailout */); err != nil {
	// 		chain.logger.Error("Error applying message", "error", err.Error())
	// 		continue
	// 	}

	// 	if err = ibs.FinalizeTx(rules, stateW); err != nil {
	// 		chain.logger.Error("Error finalizing transaction", "error", err.Error())
	// 		continue
	// 	}

	// 	includedTxs = append(includedTxs, tx)
	// }

	// if err = ibs.CommitBlock(rules, stateW); err != nil {
	// 	chain.logger.Error("Error commiting block", "error", err.Error())
	// 	return err
	// }

	// txRoot := types.DeriveSha(includedTxs)
	// if txRoot != block.TxHash() {
	// 	return fmt.Errorf("transaction root mismatch: expected: %v, got: %v", block.TxHash(), txRoot)
	// }
	// fmt.Printf("expected txRoot: %v\n, got: %v\n", block.TxHash(), txRoot)

	// stateRoot, err := calculateStateRoot(tx)
	// if err != nil {
	// 	return fmt.Errorf("error calculating state root: %w", err)
	// }
	// fmt.Printf("expected stateRoot: %v\n, got: %v\n, state root begin: %v", block.Root(), stateRoot, stateRootBegin)

	// if len(rejectedTxs) > 0 {
	// 	chain.logger.Warn("Some transactions were rejected", "count", len(rejectedTxs))
	// 	for _, rtx := range rejectedTxs {
	// 		chain.logger.Error("Transaction was rejected", "tx_number", rtx.Index, "error", rtx.Err)
	// 	}
	// }