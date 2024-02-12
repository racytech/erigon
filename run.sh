#!/usr/bin/bash

curl -L -o genesis.json https://github.com/ethereum/go-ethereum/raw/1010a79c7cbcdb4741e9f30e8cdc19c679ad7377/cmd/devp2p/internal/ethtest/testdata/genesis.json


curl -L -o chain.rlp https://github.com/ethereum/go-ethereum/raw/1010a79c7cbcdb4741e9f30e8cdc19c679ad7377/cmd/devp2p/internal/ethtest/testdata/chain.rlp

make erigon

rm -rf data

./build/bin/erigon init --datadir data genesis.json 

./build/bin/erigon --db.size.limit 2GB --datadir data --fakepow import chain.rlp 