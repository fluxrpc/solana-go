package token2022

import (
	"fmt"

	bin "github.com/fluxrpc/solana-go/binary"
	token "github.com/fluxrpc/solana-go/programs/token"
)

type ExtensionType uint16

const (
	ExtensionUninitialized ExtensionType = iota
	ExtensionTransferFeeConfig
	ExtensionTransferFeeAmount
	ExtensionMintCloseAuthority
	ExtensionConfidentialTransferMint
	ExtensionConfidentialTransferAccount
	ExtensionDefaultAccountState
	ExtensionImmutableOwner
	ExtensionMemoTransfer
	ExtensionNonTransferable
	ExtensionInterestBearingConfig
	ExtensionCPIGuard
	ExtensionPermanentDelegate
	ExtensionNonTransferableAccount
	ExtensionTransferHook
	ExtensionTransferHookAccount
	ExtensionConfidentialTransferFeeConfig
	ExtensionConfidentialTransferFeeAmount
	ExtensionMetadataPointer
	ExtensionTokenMetadata
	ExtensionGroupPointer
	ExtensionTokenGroup
	ExtensionGroupMemberPointer
	ExtensionTokenGroupMember
	ExtensionConfidentialMintBurn
	ExtensionScaledUIAmount
	ExtensionPausable
	ExtensionPausableAccount
)

type Extension struct {
	Type ExtensionType
	Data []byte
}

type Extensions []Extension

type Mint struct {
	token.Mint
	Extensions Extensions
}

func DecodeMint(data []byte) (Mint, error) {
	if len(data) != token.MintSize && len(data) < token.AccountSize+1 {
		return Mint{}, fmt.Errorf("%w: mint: %d", token.ErrInvalidAccountSize, len(data))
	}
	base, err := token.DecodeMint(data[:token.MintSize])
	if err != nil {
		return Mint{}, err
	}
	mint := Mint{Mint: base}
	if len(data) == token.MintSize {
		return mint, nil
	}
	mint.Extensions, err = decodeExtensions(data[token.AccountSize+1:])
	return mint, err
}

type Account struct {
	token.Account
	Extensions Extensions
}

func DecodeAccount(data []byte) (Account, error) {
	if len(data) != token.AccountSize && len(data) < token.AccountSize+1 {
		return Account{}, fmt.Errorf("%w: account: %d", token.ErrInvalidAccountSize, len(data))
	}
	base, err := token.DecodeAccount(data[:token.AccountSize])
	if err != nil {
		return Account{}, err
	}
	account := Account{Account: base}
	if len(data) == token.AccountSize {
		return account, nil
	}
	account.Extensions, err = decodeExtensions(data[token.AccountSize+1:])
	return account, err
}

type Multisig = token.Multisig

func DecodeMultisig(data []byte) (Multisig, error) { return token.DecodeMultisig(data) }

func decodeExtensions(data []byte) (Extensions, error) {
	dec := bin.NewDecoder(data)
	extensions := make([]Extension, 0, 4)
	for dec.Remaining() >= 2 {
		typ := ExtensionType(dec.ReadUint16())
		if err := dec.Err(); err != nil {
			return nil, fmt.Errorf("decode token-2022 extension type: %w", err)
		}
		if typ == ExtensionUninitialized {
			break
		}
		payload := dec.ReadBytes(int(dec.ReadUint16()))
		if err := dec.Err(); err != nil {
			return nil, fmt.Errorf("decode token-2022 extension %d: %w", typ, err)
		}
		extensions = append(extensions, Extension{Type: typ, Data: payload})
	}
	return extensions, nil
}

func (extensions Extensions) Find(typ ExtensionType) []byte {
	for _, extension := range extensions {
		if extension.Type == typ {
			return extension.Data
		}
	}
	return nil
}
