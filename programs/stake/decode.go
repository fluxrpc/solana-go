package stake

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// DecodeInstruction decodes one Stake Program instruction. The supplied
// account slice and decoded seed strings alias the inputs. As with native
// bincode deserialization, trailing data is allowed.
func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("stake instruction type: %w", err)
	}
	out := DecodedInstruction{Type: typ}
	switch typ {
	case InitializeInstruction:
		out.Initialize = &Initialize{instruction: instruction{AccountMetaSlice: accounts}}
		out.Initialize.Authorized.decode(dec)
		out.Initialize.Lockup.decode(dec)
	case AuthorizeInstruction:
		out.Authorize = &Authorize{
			NewAuthorized:  dec.ReadPublicKey(),
			StakeAuthorize: StakeAuthorize(dec.ReadUint32()),
			instruction:    instruction{AccountMetaSlice: accounts},
		}
	case DelegateStakeInstruction:
		out.DelegateStake = &DelegateStake{instruction{accounts}}
	case SplitInstruction:
		out.Split = &Split{Lamports: dec.ReadUint64(), instruction: instruction{AccountMetaSlice: accounts}}
	case WithdrawInstruction:
		out.Withdraw = &Withdraw{Lamports: dec.ReadUint64(), instruction: instruction{AccountMetaSlice: accounts}}
	case DeactivateInstruction:
		out.Deactivate = &Deactivate{instruction{accounts}}
	case SetLockupInstruction:
		out.SetLockup = &SetLockup{instruction: instruction{AccountMetaSlice: accounts}}
		out.SetLockup.Args.decode(dec)
	case MergeInstruction:
		out.Merge = &Merge{instruction{accounts}}
	case AuthorizeWithSeedInstruction:
		out.AuthorizeWithSeed = &AuthorizeWithSeed{
			Args: AuthorizeWithSeedArgs{
				NewAuthorized:  dec.ReadPublicKey(),
				StakeAuthorize: StakeAuthorize(dec.ReadUint32()),
				AuthoritySeed:  dec.ReadBincodeString(),
				AuthorityOwner: dec.ReadPublicKey(),
			},
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case InitializeCheckedInstruction:
		out.InitializeChecked = &InitializeChecked{instruction{accounts}}
	case AuthorizeCheckedInstruction:
		out.AuthorizeChecked = &AuthorizeChecked{
			StakeAuthorize: StakeAuthorize(dec.ReadUint32()),
			instruction:    instruction{AccountMetaSlice: accounts},
		}
	case AuthorizeCheckedWithSeedInstruction:
		out.AuthorizeCheckedWithSeed = &AuthorizeCheckedWithSeed{
			Args: AuthorizeCheckedWithSeedArgs{
				StakeAuthorize: StakeAuthorize(dec.ReadUint32()),
				AuthoritySeed:  dec.ReadBincodeString(),
				AuthorityOwner: dec.ReadPublicKey(),
			},
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case SetLockupCheckedInstruction:
		out.SetLockupChecked = &SetLockupChecked{instruction: instruction{AccountMetaSlice: accounts}}
		out.SetLockupChecked.Args.decode(dec)
	case GetMinimumDelegationInstruction:
		out.GetMinimumDelegation = &GetMinimumDelegation{instruction{accounts}}
	case DeactivateDelinquentInstruction:
		out.DeactivateDelinquent = &DeactivateDelinquent{instruction{accounts}}
	case RedelegateInstruction:
		out.Redelegate = &Redelegate{instruction{accounts}}
	case MoveStakeInstruction:
		out.MoveStake = &MoveStake{Lamports: dec.ReadUint64(), instruction: instruction{AccountMetaSlice: accounts}}
	case MoveLamportsInstruction:
		out.MoveLamports = &MoveLamports{Lamports: dec.ReadUint64(), instruction: instruction{AccountMetaSlice: accounts}}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint32(typ))
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode stake %s: %w", typ, err)
	}
	if err := out.validate(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode stake %s: %w", typ, err)
	}
	return out, nil
}
