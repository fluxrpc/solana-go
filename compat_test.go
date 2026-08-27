package solana_go

import "testing"

func TestCompatibilityHelpers(t *testing.T) {
	if got := Meta(pdaWallet); got.PublicKey != pdaWallet {
		t.Fatalf("Meta() public key = %s, want %s", got.PublicKey, pdaWallet)
	}
	if !IsAnyOfEncodingType(EncodingBase64, EncodingBase58, EncodingBase64) {
		t.Fatal("IsAnyOfEncodingType() did not find allowed encoding")
	}
	if IsAnyOfEncodingType(EncodingBase64, EncodingBase58) {
		t.Fatal("IsAnyOfEncodingType() found disallowed encoding")
	}

	curveKey := testPrivateKey(t).PublicKey()
	if got, want := IsOnCurve(curveKey[:]), curveKey.IsOnCurve(); got != want {
		t.Fatalf("IsOnCurve() = %v, want %v", got, want)
	}
	if IsOnCurve([]byte{1, 2, 3}) {
		t.Fatal("IsOnCurve() accepted a short key")
	}

	wantKey, wantErr := pdaWallet.CreateWithSeed("compat", SystemProgramID)
	gotKey, gotErr := CreateWithSeed(pdaWallet, "compat", SystemProgramID)
	assertCompatKeyResult(t, "CreateWithSeed", gotKey, gotErr, wantKey, wantErr)

	seeds := [][]byte{[]byte("compat"), {255}}
	wantKey, wantErr = TokenProgramID.CreateProgramAddress(seeds)
	gotKey, gotErr = CreateProgramAddress(seeds, TokenProgramID)
	assertCompatKeyResult(t, "CreateProgramAddress", gotKey, gotErr, wantKey, wantErr)

	findSeeds := [][]byte{[]byte("compat")}
	wantKey, wantBump, wantErr := TokenProgramID.FindProgramAddress(findSeeds)
	gotKey, gotBump, gotErr := FindProgramAddress(findSeeds, TokenProgramID)
	assertCompatPDAResult(t, "FindProgramAddress", gotKey, gotBump, gotErr, wantKey, wantBump, wantErr)

	wantKey, wantBump, wantErr = pdaWallet.FindAssociatedTokenAddress(pdaMint)
	gotKey, gotBump, gotErr = FindAssociatedTokenAddress(pdaWallet, pdaMint)
	assertCompatPDAResult(t, "FindAssociatedTokenAddress", gotKey, gotBump, gotErr, wantKey, wantBump, wantErr)

	wantKey, wantBump, wantErr = pdaWallet.FindAssociatedTokenAddressWithProgram(pdaMint, Token2022ProgramID)
	gotKey, gotBump, gotErr = FindAssociatedTokenAddressWithProgram(pdaWallet, pdaMint, Token2022ProgramID)
	assertCompatPDAResult(t, "FindAssociatedTokenAddressWithProgram", gotKey, gotBump, gotErr, wantKey, wantBump, wantErr)

	wantKey, wantBump, wantErr = pdaMint.FindTokenMetadataAddress()
	gotKey, gotBump, gotErr = FindTokenMetadataAddress(pdaMint)
	assertCompatPDAResult(t, "FindTokenMetadataAddress", gotKey, gotBump, gotErr, wantKey, wantBump, wantErr)
}

func assertCompatKeyResult(t *testing.T, name string, got PublicKey, gotErr error, want PublicKey, wantErr error) {
	t.Helper()
	if got != want || gotErr != wantErr {
		t.Fatalf("%s() = %s, %v; want %s, %v", name, got, gotErr, want, wantErr)
	}
}

func assertCompatPDAResult(t *testing.T, name string, got PublicKey, gotBump uint8, gotErr error, want PublicKey, wantBump uint8, wantErr error) {
	t.Helper()
	if got != want || gotBump != wantBump || gotErr != wantErr {
		t.Fatalf("%s() = %s, %d, %v; want %s, %d, %v", name, got, gotBump, gotErr, want, wantBump, wantErr)
	}
}

func BenchmarkFindProgramAddressCompatibility(b *testing.B) {
	seeds := [][]byte{pdaWallet[:], TokenProgramID[:], pdaMint[:]}
	b.Run("method", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkPDA, benchmarkPDABump, err = SPLAssociatedTokenAccountProgramID.FindProgramAddress(seeds)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("package-helper", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkPDA, benchmarkPDABump, err = FindProgramAddress(seeds, SPLAssociatedTokenAccountProgramID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
