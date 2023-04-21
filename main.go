package main

import (
	"context"
	"math/big"

	"github.com/ethereum-optimism/optimism/opnode"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO --rpc-url
// TODO --target-number
// TODO --reorg-step

func main() {
	// create a new OP-Node client
	opNode, err := opnode.Dial("http://optimism-qanet-master-op-node.bk.nodereal.cc")
	if err != nil {
		panic(err)
	}

	conn, err := ethclient.Dial("https://optimism-qanet-master.bk.nodereal.cc")
	if err != nil {
		panic(err)
	}

	latestNumber, err := conn.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	targetNumber := latestNumber - 10
	// reorgStep := uint64(2)

	targetBlock, err := conn.BlockByNumber(context.Background(), big.NewInt(targetNumber))
	if err != nil {
		panic(err)
	}

	c.ForkchoiceUpdate()

	// for blockNumber := uint64(1208728); blockNumber > 0; blockNumber -= 1 {
	// for blockNumber := uint64(1208728); blockNumber > 0; blockNumber += 100 {
	//for blockNumber := uint64(1209128); blockNumber > 0; blockNumber += 1 {
	//	if err != nil {
	//		panic(err)
	//	}
	//}
}
