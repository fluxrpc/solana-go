package system

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// DecodeInstruction decodes one System Program instruction. The accounts are
// the instruction-local accounts, not the transaction message's full list.
// The decoded instruction retains the supplied account slice, and decoded
// string fields alias data; neither input may be mutated while the result is
// in use.
func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("system instruction type: %w", err)
	}

	out := DecodedInstruction{Type: typ}
	switch typ {
	case CreateAccountInstruction:
		out.CreateAccount = &CreateAccount{
			Lamports:    dec.ReadUint64(),
			Space:       dec.ReadUint64(),
			Owner:       dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case AssignInstruction:
		out.Assign = &Assign{
			Owner:       dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case TransferInstruction:
		out.Transfer = &Transfer{
			Lamports:    dec.ReadUint64(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case CreateAccountWithSeedInstruction:
		out.CreateAccountWithSeed = &CreateAccountWithSeed{
			Base:        dec.ReadPublicKey(),
			Seed:        dec.ReadBincodeString(),
			Lamports:    dec.ReadUint64(),
			Space:       dec.ReadUint64(),
			Owner:       dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case AdvanceNonceAccountInstruction:
		out.AdvanceNonceAccount = &AdvanceNonceAccount{instruction: instruction{AccountMetaSlice: accounts}}
	case WithdrawNonceAccountInstruction:
		out.WithdrawNonceAccount = &WithdrawNonceAccount{
			Lamports:    dec.ReadUint64(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case InitializeNonceAccountInstruction:
		out.InitializeNonceAccount = &InitializeNonceAccount{
			Authorized:  dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case AuthorizeNonceAccountInstruction:
		out.AuthorizeNonceAccount = &AuthorizeNonceAccount{
			Authorized:  dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case AllocateInstruction:
		out.Allocate = &Allocate{
			Space:       dec.ReadUint64(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case AllocateWithSeedInstruction:
		out.AllocateWithSeed = &AllocateWithSeed{
			Base:        dec.ReadPublicKey(),
			Seed:        dec.ReadBincodeString(),
			Space:       dec.ReadUint64(),
			Owner:       dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case AssignWithSeedInstruction:
		out.AssignWithSeed = &AssignWithSeed{
			Base:        dec.ReadPublicKey(),
			Seed:        dec.ReadBincodeString(),
			Owner:       dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case TransferWithSeedInstruction:
		out.TransferWithSeed = &TransferWithSeed{
			Lamports:    dec.ReadUint64(),
			FromSeed:    dec.ReadBincodeString(),
			FromOwner:   dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case UpgradeNonceAccountInstruction:
		out.UpgradeNonceAccount = &UpgradeNonceAccount{instruction: instruction{AccountMetaSlice: accounts}}
	case CreateAccountAllowPrefundInstruction:
		out.CreateAccountAllowPrefund = &CreateAccountAllowPrefund{
			Lamports:    dec.ReadUint64(),
			Space:       dec.ReadUint64(),
			Owner:       dec.ReadPublicKey(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint32(typ))
	}

	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode system %s: %w", typ, err)
	}
	return out, nil
}
