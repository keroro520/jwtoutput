package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/golang-jwt/jwt/v4"
	"math/big"
	"time"
)

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

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	jwtSecretFlag := flag.String("jwt-secret", "", "JWT secret")
	ethRpcUrlFlag := flag.String("eth-rpc-url", "http://127.0.0.1:8545", "ETH RPC URL")
	engineRpcUrlFlag := flag.String("engine-rpc-url", "http://127.0.0.1:8551", "Engine RPC URL")
	flag.Parse()

	jwtSecret := common.FromHex(*jwtSecretFlag)
	jwtToken, err := JWTAuthToken(jwtSecret)
	panicIfError(err)

	rpcClient, err := rpc.DialContext(context.Background(), *ethRpcUrlFlag)
	ethClient := ethclient.NewClient(rpcClient)
	panicIfError(err)
	engineRpc, err := rpc.DialContext(context.Background(), *engineRpcUrlFlag)
	panicIfError(err)
	engineRpc.SetHeader("Authorization", jwtToken)

	chainID, err := ethClient.ChainID(context.Background())
	panicIfError(err)
	latestNumber, err := ethClient.BlockNumber(context.Background())
	panicIfError(err)
	latestHeader, err := ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(latestNumber)))
	panicIfError(err)

	targetNumber := latestNumber - 20
	targetBlock, err := ethClient.BlockByNumber(context.Background(), big.NewInt(int64(targetNumber)))
	panicIfError(err)

	privkey, err := crypto.HexToECDSA("701b615bbdfb9de65240bc28bd21bbc0d996645a3dd57e7b12bc2bdf6f192c82")
	panicIfError(err)
	senderAddr := common.HexToAddress("0x71bE63f3384f5fb98995898A86B02Fb2426c5788")
	receipianAddr := common.HexToAddress("0x71bE63f3384f5fb98995898A86B02Fb2426c5788")
	value := big.NewInt(1000000000000000000) // 1 ETH in wei
	gasLimit := uint64(21000)
	nonce, err := ethClient.PendingNonceAt(context.Background(), senderAddr)
	panicIfError(err)
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	panicIfError(err)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &receipianAddr,
		GasTipCap: gasPrice,
		GasFeeCap: gasPrice,
		Gas:       gasLimit,
		Value:     value,
	})
	tx, err = types.SignTx(tx, types.NewLondonSigner(chainID), privkey)
	panicIfError(err)

	nextDepositTx, err := targetBlock.Transactions()[0].MarshalBinary()
	panicIfError(err)

	marshaled, err := tx.MarshalBinary()
	panicIfError(err)

	var transactions = []eth.Data{nextDepositTx, marshaled}
	var attributes = eth.PayloadAttributes{
		Timestamp:             eth.Uint64Quantity(targetBlock.Time()),
		PrevRandao:            eth.Bytes32{0x15},
		SuggestedFeeRecipient: common.Address{},
		Transactions:          transactions,
		NoTxPool:              false,
		GasLimit:              (*eth.Uint64Quantity)(&latestHeader.GasLimit),
	}
	var fc = eth.ForkchoiceState{
		FinalizedBlockHash: common.Hash{},
		SafeBlockHash:      targetBlock.ParentHash(),
		HeadBlockHash:      targetBlock.ParentHash(),
	}
	var result eth.ForkchoiceUpdatedResult
	err = engineRpc.CallContext(context.Background(), &result, "engine_forkchoiceUpdatedV1", fc, attributes)
	if err != nil {
		panic(err)
	}
	fmt.Println("### engine_forkchoiceUpdatedV1")
	fmt.Println()
	fmt.Printf("params.eth.ForkchoiceState: %+v", fc)
	fmt.Println()
	fmt.Println()
	fmt.Printf("params.eth.PayloadAttributes: %+v", attributes)
	fmt.Println()
	fmt.Println()
	fmt.Printf("result.ForkchoiceUpdatedResult: %+v", result)
	fmt.Println()
	fmt.Println()

	var executableData engine.ExecutableData
	err = engineRpc.CallContext(context.Background(), &executableData, "engine_getPayloadV1", result.PayloadID)
	panicIfError(err)
	fmt.Println("### engine_getPayloadV1")
	fmt.Printf("params.PayloadID: %s", result.PayloadID.String())
	fmt.Println()
	fmt.Printf("result.ExecutableData: %+v", executableData)
	fmt.Println()
	fmt.Println()

	var payloadStatus eth.PayloadStatusV1
	err = engineRpc.CallContext(context.Background(), &payloadStatus, "engine_newPayloadV1", executableData)
	panicIfError(err)
	fmt.Println("### engine_newPayloadV1")
	fmt.Println()
	fmt.Printf("params.executableData")
	fmt.Printf("result.PayloadStatusV1: %+v", payloadStatus)
	fmt.Println()
	fmt.Println()
	callAgain(engineRpc, *payloadStatus.LatestValidHash)
	fmt.Println("### Compare")

	newLatestNumber, err := ethClient.BlockNumber(context.Background())
	panicIfError(err)
	newLatestHeader, err := ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(newLatestNumber)))
	panicIfError(err)
	fmt.Println("Before Latest: ", latestHeader.Number, latestHeader.Hash())
	fmt.Println("After  Latest: ", newLatestHeader.Number, newLatestHeader.Hash())

	fmt.Println()
	fmt.Println()
	fmt.Println("Before Target: ", targetBlock.Number(), targetBlock.Hash())
	newTargetHeader, err := ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(targetBlock.NumberU64())))
	fmt.Println("After  Target: ", newTargetHeader.Number, newTargetHeader.Hash())
}

// headBlockHash is not canonical head block hash before, so we have to call engine_forkchoiceUpdatedV1 again,
// to trigger the headBlockHash becomes canonical head block hash
// https://github.com/ethereum-optimism/op-geth/blob/70234f9b682ac7fdc7f12fca6ec611c340ec9940/eth/catalyst/api.go#L296-L299
func callAgain(engineRpc *rpc.Client, headBlockHash common.Hash) {
	var attributes *eth.PayloadAttributes
	var fc = eth.ForkchoiceState{
		FinalizedBlockHash: common.Hash{},
		SafeBlockHash:      headBlockHash,
		HeadBlockHash:      headBlockHash,
	}
	var result eth.ForkchoiceUpdatedResult
	err := engineRpc.CallContext(context.Background(), &result, "engine_forkchoiceUpdatedV1", fc, attributes)
	fmt.Println("### engine_forkchoiceUpdatedV1")
	fmt.Println()
	fmt.Printf("params.ForkchoiceState: %+v", fc)
	fmt.Println()
	fmt.Println()
	fmt.Println("params.PayloadAttributes: nil")
	fmt.Println()
	panicIfError(err)
}
