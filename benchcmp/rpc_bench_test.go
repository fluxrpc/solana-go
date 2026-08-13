package benchcmp

// Compares the rpc package response types through each SDK's canonical JSON
// path: flux uses bytedance/sonic, upstream uses goccy/go-json (what its rpc
// client uses internally).

import (
	"testing"

	"github.com/bytedance/sonic"

	fluxrpc "github.com/fluxrpc/solana-go/rpc"
	gaglrpc "github.com/gagliardetto/solana-go/rpc"
	gojson "github.com/goccy/go-json"
)

// Realistic transaction meta of a getBlock response; same fixture as
// rpc/transaction_meta_test.go (assembled from upstream client_test.go).
const rpcTransactionMetaFixture = `{
	"err": null,
	"fee": 5000,
	"innerInstructions": [
		{"index": 0, "instructions": [
			{"programIdIndex": 4, "accounts": [1, 2, 3, 0], "data": "3yZe7d", "stackHeight": 2}
		]}
	],
	"logMessages": [
		"Program Vote111111111111111111111111111111111111111 invoke [1]",
		"Program Vote111111111111111111111111111111111111111 success"
	],
	"preBalances": [441866068495, 40905918933763, 1, 1, 1],
	"postBalances": [441866063495, 40905918933763, 1, 1, 1],
	"preTokenBalances": [{"accountIndex": 4, "mint": "E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1", "owner": "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo", "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", "uiTokenAmount": {"amount": "47444666", "decimals": 6, "uiAmount": 47.444666, "uiAmountString": "47.444666"}}],
	"postTokenBalances": [{"accountIndex": 4, "mint": "E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1", "owner": "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo", "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", "uiTokenAmount": {"amount": "0", "decimals": 6, "uiAmount": null, "uiAmountString": "0"}}],
	"rewards": [{"pubkey": "5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq", "lamports": 1595000, "postBalance": 482032983798, "rewardType": "Fee"}],
	"loadedAddresses": {"readonly": ["11111111111111111111111111111111"], "writable": ["4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg"]},
	"returnData": {"programId": "11111111111111111111111111111111", "data": ["aGVsbG8=", "base64"]},
	"status": {"Ok": null},
	"computeUnitsConsumed": 2100
}`

// Realistic getAccountInfo value with the u64-max rentEpoch sentinel.
const rpcAccountFixture = `{"lamports":88849814690250,"owner":"11111111111111111111111111111111","data":["dGVzdCBkYXRh","base64"],"executable":false,"rentEpoch":18446744073709551615,"space":9}`

var (
	sinkFluxMeta fluxrpc.TransactionMeta
	sinkGaglMeta gaglrpc.TransactionMeta
	sinkFluxAcct fluxrpc.Account
	sinkGaglAcct gaglrpc.Account
)

func BenchmarkRpcTransactionMeta_UnmarshalJSON(b *testing.B) {
	data := []byte(rpcTransactionMetaFixture)
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := sonic.Unmarshal(data, &sinkFluxMeta); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := gojson.Unmarshal(data, &sinkGaglMeta); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcTransactionMeta_MarshalJSON(b *testing.B) {
	if err := sonic.Unmarshal([]byte(rpcTransactionMetaFixture), &sinkFluxMeta); err != nil {
		b.Fatal(err)
	}
	if err := gojson.Unmarshal([]byte(rpcTransactionMetaFixture), &sinkGaglMeta); err != nil {
		b.Fatal(err)
	}
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkBytes, err = sonic.Marshal(&sinkFluxMeta)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkBytes, err = gojson.Marshal(&sinkGaglMeta)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcAccount_UnmarshalJSON(b *testing.B) {
	data := []byte(rpcAccountFixture)
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := sonic.Unmarshal(data, &sinkFluxAcct); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := gojson.Unmarshal(data, &sinkGaglAcct); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcAccount_MarshalJSON(b *testing.B) {
	if err := sonic.Unmarshal([]byte(rpcAccountFixture), &sinkFluxAcct); err != nil {
		b.Fatal(err)
	}
	if err := gojson.Unmarshal([]byte(rpcAccountFixture), &sinkGaglAcct); err != nil {
		b.Fatal(err)
	}
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkBytes, err = sonic.Marshal(&sinkFluxAcct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkBytes, err = gojson.Marshal(&sinkGaglAcct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Fixtures shared with rpc/get_block_test.go, rpc/get_transaction_test.go
// and rpc/parsed_transaction_test.go (assembled from upstream client_test.go).
const rpcGetBlockResultFixture = `{"blockHeight":69213636,"blockTime":1625227950,"blockhash":"5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB","parentSlot":83987983,"previousBlockhash":"Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN","rewards":[{"lamports":1595000,"postBalance":482032983798,"pubkey":"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq","rewardType":"Fee"}],"transactions":[{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[441866063495,40905918933763,1,1,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1,1,1],"preTokenBalances":[],"rewards":[],"returnData":{"programId":"11111111111111111111111111111111","data":["aGVsbG8=","base64"]},"status":{"Ok":null}},"transaction":["AQp2TH1spzjBAVM3alvnpaePFx3YEo9dvRglDuSChZUoTMD\/\/2h0HY5+89LJjCdiGJ7Ph3+Fyvbeiz1uJF8gxw0BAAMFyH0KDkXtjL1xebUYflZxYGlpV+LvjazzZCb\/mF2T67xZmkOUM\/A0iDSEkFzD5m4Ol82vsojigvqxrmp7Z1vrQgan1RcZLwqvxvJl4\/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ\/wNo1OAAAAAAAMFYbeqrsxJ9\/vZxtOaFi3rT2w9RF5Xi4jsyu61f3t1AQQEAQIDAAR0ZXN0","base64"]},{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[334759887662,151357332545078,1,1,1],"postTokenBalances":[],"preBalances":[334759892662,151357332545078,1,1,1],"preTokenBalances":[],"rewards":[],"returnData":{"programId":"11111111111111111111111111111111","data":["aGVsbG8=","base64"]},"status":{"Ok":null}},"transaction":["ATA7DkBatbe2JB43QV+QRj2yoXSMXXttYFggDxZYOBfsRyYuGtzrbUevivclchxVccRIPlRP9PtS\/9NPXlwmhwwBAAMFSDrhjiNPuNqc4BWwitZz7xJ2NIXtv6XZtwtEOmgLj3n3NQ+OONLFlsu0LoUBSDsp40i9jOjZJBsliMtvTfdV+gan1RcZLwqvxvJl4\/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ\/wNo1OAAAAAAAKlcZMqS\/Oh0v+kOq2Ipg73NqbvKBRGQJDK8\/01K+MBAQQEAQIDAAR0ZXN0","base64"]}]}`

const rpcGetTransactionResultFixture = `{"blockTime":1625227950,"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[441866063495,40905918933763,1,1,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1,1,1],"preTokenBalances":[],"returnData":{"programId":"11111111111111111111111111111111","data":["","base64"]},"rewards":[],"status":{"Ok":null}},"slot":83987984,"transaction":["AQp2TH1spzjBAVM3alvnpaePFx3YEo9dvRglDuSChZUoTMD//2h0HY5+89LJjCdiGJ7Ph3+Fyvbeiz1uJF8gxw0BAAMFyH0KDkXtjL1xebUYflZxYGlpV+LvjazzZCb/mF2T67xZmkOUM/A0iDSEkFzD5m4Ol82vsojigvqxrmp7Z1vrQgan1RcZLwqvxvJl4/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ/wNo1OAAAAAAAMFYbeqrsxJ9/vZxtOaFi3rT2w9RF5Xi4jsyu61f3t1AQQEAQIDAAR0ZXN0","base64"],"version":"legacy"}`

const rpcParsedTransactionFixture = `{"message":{"accountKeys":[{"pubkey":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","signer":true,"writable":true}],"addressTableLookups":null,"instructions":[{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"},{"parsed":{"info":{"amount":"47444666","delegate":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","source":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly"},"type":"approve"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"accounts":["G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"],"data":"2dmnzvSCNoP8bNbUnUtk7FTYod5czhUfk4E7LSPNMtK4V1FHgQVYeQ2GnsEtCKZCyLLHXvnkReP","programId":"wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb"}],"recentBlockhash":"9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi"},"signatures":["2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89"]}`

var (
	sinkFluxBlock  fluxrpc.GetBlockResult
	sinkGaglBlock  gaglrpc.GetBlockResult
	sinkFluxTxRes  fluxrpc.GetTransactionResult
	sinkGaglTxRes  gaglrpc.GetTransactionResult
	sinkFluxParsed fluxrpc.ParsedTransaction
	sinkGaglParsed gaglrpc.ParsedTransaction
)

func BenchmarkRpcGetBlockResult_UnmarshalJSON(b *testing.B) {
	data := []byte(rpcGetBlockResultFixture)
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := sonic.Unmarshal(data, &sinkFluxBlock); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := gojson.Unmarshal(data, &sinkGaglBlock); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcGetBlockResult_MarshalJSON(b *testing.B) {
	if err := sonic.Unmarshal([]byte(rpcGetBlockResultFixture), &sinkFluxBlock); err != nil {
		b.Fatal(err)
	}
	if err := gojson.Unmarshal([]byte(rpcGetBlockResultFixture), &sinkGaglBlock); err != nil {
		b.Fatal(err)
	}
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkBytes, err = sonic.Marshal(&sinkFluxBlock)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkBytes, err = gojson.Marshal(&sinkGaglBlock)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcGetTransactionResult_UnmarshalJSON(b *testing.B) {
	data := []byte(rpcGetTransactionResultFixture)
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := sonic.Unmarshal(data, &sinkFluxTxRes); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := gojson.Unmarshal(data, &sinkGaglTxRes); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcParsedTransaction_UnmarshalJSON(b *testing.B) {
	data := []byte(rpcParsedTransactionFixture)
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := sonic.Unmarshal(data, &sinkFluxParsed); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if err := gojson.Unmarshal(data, &sinkGaglParsed); err != nil {
				b.Fatal(err)
			}
		}
	})
}
