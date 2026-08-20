package solana_go

import "testing"

func TestAccountMetaBuilders(t *testing.T) {
	key := testPublicKey()

	meta := key.Meta()
	if meta.PublicKey != key || meta.IsWritable || meta.IsSigner {
		t.Fatalf("Meta() = %+v", meta)
	}

	meta.WRITE().SIGNER()
	if !meta.IsWritable || !meta.IsSigner {
		t.Fatalf("WRITE().SIGNER() = %+v", meta)
	}

	got := NewAccountMeta(key, true, false)
	if got.PublicKey != key || !got.IsWritable || got.IsSigner {
		t.Fatalf("NewAccountMeta() = %+v", got)
	}
}

func TestAccountMetaSlice(t *testing.T) {
	signer := NewAccountMeta(testPublicKey(), false, true)
	writable := NewAccountMeta(PublicKey{1}, true, false)

	var slice AccountMetaSlice
	slice.Append(signer)
	slice.Append(writable)

	if slice.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", slice.Len())
	}
	if slice.Get(0) != signer || slice.Get(1) != writable {
		t.Fatal("Get returned wrong element")
	}
	if slice.Get(2) != nil || slice.Get(-1) != nil {
		t.Fatal("Get out of range did not return nil")
	}

	signers := slice.GetSigners()
	if len(signers) != 1 || signers[0] != signer {
		t.Fatalf("GetSigners() = %v", signers)
	}

	keys := slice.GetKeys()
	if len(keys) != 2 || keys[0] != signer.PublicKey || keys[1] != writable.PublicKey {
		t.Fatalf("GetKeys() = %v", keys)
	}
}
