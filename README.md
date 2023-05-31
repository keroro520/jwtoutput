

### Install

```shell
go install github.com/keroro520/jwtoutput
```


## Usage


### Example

```
export JWT_SECRET=<...>
kubectl exec \
    -it \
    --namespace=bsc-qa-l2-opt-qanet-ha \
    qanet-ha-optimism-bedrock-bridge-l2-1 \
    -- time curl http://127.0.0.1:8551 \
          -X POST \
          --data '{
                "jsonrpc":"2.0",
                "method":"engine_forkchoiceUpdatedV1",
                "params":[{
                    "head_block_hash": "0x1000000000000000000000000000000000000000000000000000000000000000",
                    "safe_block_hash": "0x1000000000000000000000000000000000000000000000000000000000000000",
                    "finalized_block_hash": "0x1000000000000000000000000000000000000000000000000000000000000000"}],
                "id":74
                }' \
          -H 'Content-Type: application/json' \
          -H "$(jwtoutput --jwt-secret $JWT_SECRET)"
```
