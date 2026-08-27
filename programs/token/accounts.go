package token

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type Mint struct {
	MintAuthority   *solana.PublicKey
	Supply          uint64
	Decimals        uint8
	IsInitialized   bool
	FreezeAuthority *solana.PublicKey
}

func DecodeMint(data []byte) (Mint, error) {
	if len(data) != MintSize {
		return Mint{}, fmt.Errorf("%w: mint: %d", ErrInvalidAccountSize, len(data))
	}
	dec := bin.NewDecoder(data)
	var mint Mint
	hasMintAuthority := dec.ReadCOption()
	authority := dec.ReadPublicKey()
	if hasMintAuthority {
		mint.MintAuthority = &authority
	}
	mint.Supply = dec.ReadUint64()
	mint.Decimals = dec.ReadUint8()
	mint.IsInitialized = dec.ReadBool()
	hasFreezeAuthority := dec.ReadCOption()
	authority = dec.ReadPublicKey()
	if hasFreezeAuthority {
		mint.FreezeAuthority = &authority
	}
	if err := dec.Err(); err != nil {
		return Mint{}, fmt.Errorf("decode token mint: %w", err)
	}
	return mint, nil
}

type Account struct {
	Mint            solana.PublicKey
	Owner           solana.PublicKey
	Amount          uint64
	Delegate        *solana.PublicKey
	State           AccountState
	IsNative        *uint64
	DelegatedAmount uint64
	CloseAuthority  *solana.PublicKey
}

func DecodeAccount(data []byte) (Account, error) {
	if len(data) != AccountSize {
		return Account{}, fmt.Errorf("%w: account: %d", ErrInvalidAccountSize, len(data))
	}
	dec := bin.NewDecoder(data)
	account := Account{Mint: dec.ReadPublicKey(), Owner: dec.ReadPublicKey(), Amount: dec.ReadUint64()}
	hasDelegate := dec.ReadCOption()
	delegate := dec.ReadPublicKey()
	if hasDelegate {
		account.Delegate = &delegate
	}
	account.State = AccountState(dec.ReadUint8())
	isNative := dec.ReadCOption()
	reserve := dec.ReadUint64()
	if isNative {
		account.IsNative = &reserve
	}
	account.DelegatedAmount = dec.ReadUint64()
	hasCloseAuthority := dec.ReadCOption()
	authority := dec.ReadPublicKey()
	if hasCloseAuthority {
		account.CloseAuthority = &authority
	}
	if err := dec.Err(); err != nil {
		return Account{}, fmt.Errorf("decode token account: %w", err)
	}
	return account, nil
}

type Multisig struct {
	M             uint8
	N             uint8
	IsInitialized bool
	Signers       [MaxSigners]solana.PublicKey
}

func DecodeMultisig(data []byte) (Multisig, error) {
	if len(data) != MultisigSize {
		return Multisig{}, fmt.Errorf("%w: multisig: %d", ErrInvalidAccountSize, len(data))
	}
	dec := bin.NewDecoder(data)
	multisig := Multisig{M: dec.ReadUint8(), N: dec.ReadUint8(), IsInitialized: dec.ReadBool()}
	for i := range multisig.Signers {
		multisig.Signers[i] = dec.ReadPublicKey()
	}
	if err := dec.Err(); err != nil {
		return Multisig{}, fmt.Errorf("decode token multisig: %w", err)
	}
	return multisig, nil
}
