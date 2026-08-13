package rpc

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture shape from upstream gagliardetto/solana-go rpc/types_test.go
// (TestParsedTransactionMeta_Decode), with non-empty base64 content.
const returnDataFixture = `{"programId":"11111111111111111111111111111111","data":["aGVsbG8td29ybGQ=","base64"]}`

func TestReturnDataJSON(t *testing.T) {
	ret := jsonRoundTrip[ReturnData](t, []byte(returnDataFixture))

	if ret.ProgramId.String() != "11111111111111111111111111111111" {
		t.Fatalf("ProgramId = %s", ret.ProgramId)
	}
	if ret.Data.Encoding != solana.EncodingBase64 {
		t.Fatalf("Encoding = %q", ret.Data.Encoding)
	}
	if !bytes.Equal(ret.Data.GetBinary(), []byte("hello-world")) {
		t.Fatalf("Data = %q", ret.Data.GetBinary())
	}
}

func TestReturnDataEmpty(t *testing.T) {
	ret := jsonRoundTrip[ReturnData](t, []byte(`{"programId":"11111111111111111111111111111111","data":["","base64"]}`))
	if len(ret.Data.GetBinary()) != 0 {
		t.Fatalf("Data = %q", ret.Data.GetBinary())
	}
}
