package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/golang-jwt/jwt/v4"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO --rpc-url
// TODO --target-number
// TODO --reorg-step

func JWTAuthToken(jwtSecret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": &jwt.NumericDate{Time: time.Now()},
	})
	signedToken, err := token.SignedString(jwtSecret[:])
	if err != nil {
		return "", err
	}

	// Include the signed JWT as a Bearer token in the Authorization header
	return fmt.Sprintf("Bearer %s", signedToken), nil
}

func main() {
	ethRpcUrlFlag := flag.String("eth-rpc-url", "http://127.0.0.1:8545", "ETH RPC URL")
	engineRpcUrlFlag := flag.String("engine-rpc-url", "http://127.0.0.1:8551", "Engine RPC URL")
	jwtSecretFlag := flag.String("jwt-secret", "", "JWT secret")
	flag.Parse()

	fmt.Printf("eth-rpc-url: %s", *ethRpcUrlFlag)
	fmt.Println()
	fmt.Printf("engine-rpc-url: %s", *engineRpcUrlFlag)
	fmt.Println()
	fmt.Printf("jwt-secret: %s", *jwtSecretFlag)
	fmt.Println()

	rpcClient, err := rpc.DialContext(context.Background(), *ethRpcUrlFlag)
	if err != nil {
		panic(err)
	}

	ethClient := ethclient.NewClient(rpcClient)

	latestNumber, err := ethClient.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	latestHeader, err := ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(latestNumber)))
	if err != nil {
		panic(err)
	}

	targetNumber := latestNumber - 1
	// reorgStep := uint64(2)

	targetHeader, err := ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(targetNumber)))
	if err != nil {
		panic(err)
	}

	engineRpc, err := rpc.DialContext(context.Background(), *engineRpcUrlFlag)
	if err != nil {
		panic(err)
	}

	jwtSecret := common.FromHex(*jwtSecretFlag)
	jwtToken, err := JWTAuthToken(jwtSecret)
	if err != nil {
		panic(err)
	}

	var result eth.ForkchoiceUpdatedResult
	var fc = eth.ForkchoiceState{
		FinalizedBlockHash: common.Hash{},
		SafeBlockHash:      targetHeader.Hash(),
		HeadBlockHash:      targetHeader.Hash(),
	}

	engineRpc.SetHeader("Authorization", jwtToken)

	var attributes = eth.PayloadAttributes{
		Timestamp:             eth.Uint64Quantity(targetHeader.Time + 1),
		PrevRandao:            eth.Bytes32{0x1},
		SuggestedFeeRecipient: common.Address{},
		Transactions:          nil,
		NoTxPool:              false,
		GasLimit:              (*eth.Uint64Quantity)(&targetHeader.GasLimit),
	}
	err = engineRpc.CallContext(context.Background(), &result, "engine_forkchoiceUpdatedV1", fc, attributes)
	if err != nil {
		// fmt.Printf("engine_forkchoiceUpdatedV1 error, %s", err)
		panic(err)
	}

	fmt.Printf("%+v", result)
	fmt.Println()

	fmt.Printf("fc: %+v", fc)
	fmt.Println()
	fmt.Println("===== Before =====")
	fmt.Println("Before: ", latestHeader.Number, latestHeader.Hash())
	newLatestNumber, err := ethClient.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	newLatestHeader, err := ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(newLatestNumber)))
	if err != nil {
		panic(err)
	}
	fmt.Println("After:  ", newLatestHeader.Number, newLatestHeader.Hash())

	fmt.Println("Target: ", targetHeader.Number, targetHeader.Hash())
}
