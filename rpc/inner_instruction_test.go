package rpc

import (
	"reflect"
	"testing"

	"github.com/fluxrpc/base58"
)

// Fixture: compiled (non-parsed encoding) inner instruction; instruction
// shape from upstream gagliardetto/solana-go rpc/client_test.go vote
// transaction fixtures ("data":"3yZe7d","programIdIndex":4), with the
// stackHeight the RPC reports for nested invocations.
const innerInstructionFixture = `{"index":2,"instructions":[{"programIdIndex":4,"accounts":[1,2,3,0],"data":"3yZe7d","stackHeight":2}]}`

func TestInnerInstructionJSON(t *testing.T) {
	inner := jsonRoundTrip[InnerInstruction](t, []byte(innerInstructionFixture))

	if inner.Index != 2 {
		t.Fatalf("Index = %d", inner.Index)
	}
	if len(inner.Instructions) != 1 {
		t.Fatalf("Instructions = %+v", inner.Instructions)
	}
	ix := inner.Instructions[0]
	if ix.ProgramIDIndex != 4 || ix.StackHeight != 2 {
		t.Fatalf("got %+v", ix)
	}
	if !reflect.DeepEqual(ix.Accounts, []uint16{1, 2, 3, 0}) {
		t.Fatalf("Accounts = %v", ix.Accounts)
	}
	wantData, err := base58.Decode("3yZe7d")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]byte(ix.Data), wantData) {
		t.Fatalf("Data = %x, want %x", ix.Data, wantData)
	}
}

func TestCompiledInstructionMissingStackHeight(t *testing.T) {
	// Top-level instructions and old responses have no stackHeight.
	ix := jsonRoundTrip[CompiledInstruction](t, []byte(`{"programIdIndex":1,"accounts":[0,1],"data":"3Bxs4ThLFRfx6J7z"}`))
	if ix.StackHeight != 0 {
		t.Fatalf("StackHeight = %d", ix.StackHeight)
	}
}
