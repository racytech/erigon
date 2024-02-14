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