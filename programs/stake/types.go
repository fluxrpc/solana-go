package stake

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type instruction struct{ solana.AccountMetaSlice }

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

// InstructionType is the little-endian uint32 bincode enum variant encoded at
// the start of every Stake Program instruction.
type InstructionType uint32

func (typ InstructionType) String() string {
	switch typ {
	case InitializeInstruction:
		return "Initialize"
	case AuthorizeInstruction:
		return "Authorize"
	case DelegateStakeInstruction:
		return "DelegateStake"
	case SplitInstruction:
		return "Split"
	case WithdrawInstruction:
		return "Withdraw"
	case DeactivateInstruction:
		return "Deactivate"
	case SetLockupInstruction:
		return "SetLockup"
	case MergeInstruction:
		return "Merge"
	case AuthorizeWithSeedInstruction:
		return "AuthorizeWithSeed"
	case InitializeCheckedInstruction:
		return "InitializeChecked"
	case AuthorizeCheckedInstruction:
		return "AuthorizeChecked"
	case AuthorizeCheckedWithSeedInstruction:
		return "AuthorizeCheckedWithSeed"
	case SetLockupCheckedInstruction:
		return "SetLockupChecked"
	case GetMinimumDelegationInstruction:
		return "GetMinimumDelegation"
	case DeactivateDelinquentInstruction:
		return "DeactivateDelinquent"
	case RedelegateInstruction:
		return "Redelegate"
	case MoveStakeInstruction:
		return "MoveStake"
	case MoveLamportsInstruction:
		return "MoveLamports"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint32(typ))
	}
}

// DecodedInstruction is the fully typed result of DecodeInstruction. Type
// identifies the one non-nil concrete field.
type DecodedInstruction struct {
	Type                     InstructionType
	Initialize               *Initialize
	Authorize                *Authorize
	DelegateStake            *DelegateStake
	Split                    *Split
	Withdraw                 *Withdraw
	Deactivate               *Deactivate
	SetLockup                *SetLockup
	Merge                    *Merge
	AuthorizeWithSeed        *AuthorizeWithSeed
	InitializeChecked        *InitializeChecked
	AuthorizeChecked         *AuthorizeChecked
	AuthorizeCheckedWithSeed *AuthorizeCheckedWithSeed
	SetLockupChecked         *SetLockupChecked
	GetMinimumDelegation     *GetMinimumDelegation
	DeactivateDelinquent     *DeactivateDelinquent
	Redelegate               *Redelegate
	MoveStake                *MoveStake
	MoveLamports             *MoveLamports
}

func (out DecodedInstruction) validate() error {
	switch out.Type {
	case AuthorizeInstruction:
		if !out.Authorize.StakeAuthorize.valid() {
			return ErrInvalidStakeAuthorize
		}
	case AuthorizeWithSeedInstruction:
		if !out.AuthorizeWithSeed.Args.StakeAuthorize.valid() {
			return ErrInvalidStakeAuthorize
		}
		if len(out.AuthorizeWithSeed.Args.AuthoritySeed) > solana.MaxSeedLength {
			return solana.ErrMaxSeedLengthExceeded
		}
	case AuthorizeCheckedInstruction:
		if !out.AuthorizeChecked.StakeAuthorize.valid() {
			return ErrInvalidStakeAuthorize
		}
	case AuthorizeCheckedWithSeedInstruction:
		if !out.AuthorizeCheckedWithSeed.Args.StakeAuthorize.valid() {
			return ErrInvalidStakeAuthorize
		}
		if len(out.AuthorizeCheckedWithSeed.Args.AuthoritySeed) > solana.MaxSeedLength {
			return solana.ErrMaxSeedLengthExceeded
		}
	}
	return nil
}

// StakeAuthorize identifies which stake-account authority is changed.
type StakeAuthorize uint32

func (authority StakeAuthorize) String() string {
	switch authority {
	case StakeAuthorizeStaker:
		return "Staker"
	case StakeAuthorizeWithdrawer:
		return "Withdrawer"
	default:
		return fmt.Sprintf("StakeAuthorize(%d)", uint32(authority))
	}
}

func (authority StakeAuthorize) valid() bool {
	return authority == StakeAuthorizeStaker || authority == StakeAuthorizeWithdrawer
}

// Authorized contains the staker and withdrawal authorities stored in stake
// account state and supplied to Initialize.
type Authorized struct {
	Staker     solana.PublicKey
	Withdrawer solana.PublicKey
}

// Lockup contains the complete lockup configuration supplied to Initialize.
type Lockup struct {
	UnixTimestamp int64
	Epoch         uint64
	Custodian     solana.PublicKey
}

// LockupArgs contains optional SetLockup replacements. A nil field preserves
// the corresponding value already stored in the stake account.
type LockupArgs struct {
	UnixTimestamp *int64
	Epoch         *uint64
	Custodian     *solana.PublicKey
}

// LockupCheckedArgs contains optional SetLockupChecked replacements. A new
// custodian, when present, is represented by the instruction's final signer
// account rather than instruction data.
type LockupCheckedArgs struct {
	UnixTimestamp *int64
	Epoch         *uint64
}

// AuthorizeWithSeedArgs is the encoded argument set for AuthorizeWithSeed.
type AuthorizeWithSeedArgs struct {
	NewAuthorized  solana.PublicKey
	StakeAuthorize StakeAuthorize
	AuthoritySeed  string
	AuthorityOwner solana.PublicKey
}

// AuthorizeCheckedWithSeedArgs is the encoded argument set for
// AuthorizeCheckedWithSeed. The new authority is a signer account.
type AuthorizeCheckedWithSeedArgs struct {
	StakeAuthorize StakeAuthorize
	AuthoritySeed  string
	AuthorityOwner solana.PublicKey
}

func (value Authorized) write(enc *binary.Encoder) {
	enc.WritePublicKey(value.Staker)
	enc.WritePublicKey(value.Withdrawer)
}

func (value *Authorized) decode(dec *binary.Decoder) {
	value.Staker = dec.ReadPublicKey()
	value.Withdrawer = dec.ReadPublicKey()
}

func (value Lockup) write(enc *binary.Encoder) {
	enc.WriteInt64(value.UnixTimestamp)
	enc.WriteUint64(value.Epoch)
	enc.WritePublicKey(value.Custodian)
}

func (value *Lockup) decode(dec *binary.Decoder) {
	value.UnixTimestamp = dec.ReadInt64()
	value.Epoch = dec.ReadUint64()
	value.Custodian = dec.ReadPublicKey()
}

func (value LockupArgs) write(enc *binary.Encoder) {
	enc.WriteOption(value.UnixTimestamp != nil)
	if value.UnixTimestamp != nil {
		enc.WriteInt64(*value.UnixTimestamp)
	}
	enc.WriteOption(value.Epoch != nil)
	if value.Epoch != nil {
		enc.WriteUint64(*value.Epoch)
	}
	enc.WriteOption(value.Custodian != nil)
	if value.Custodian != nil {
		enc.WritePublicKey(*value.Custodian)
	}
}

func (value *LockupArgs) decode(dec *binary.Decoder) {
	if dec.ReadOption() {
		unixTimestamp := dec.ReadInt64()
		value.UnixTimestamp = &unixTimestamp
	}
	if dec.ReadOption() {
		epoch := dec.ReadUint64()
		value.Epoch = &epoch
	}
	if dec.ReadOption() {
		custodian := dec.ReadPublicKey()
		value.Custodian = &custodian
	}
}

func (value LockupCheckedArgs) write(enc *binary.Encoder) {
	enc.WriteOption(value.UnixTimestamp != nil)
	if value.UnixTimestamp != nil {
		enc.WriteInt64(*value.UnixTimestamp)
	}
	enc.WriteOption(value.Epoch != nil)
	if value.Epoch != nil {
		enc.WriteUint64(*value.Epoch)
	}
}

func (value *LockupCheckedArgs) decode(dec *binary.Decoder) {
	if dec.ReadOption() {
		unixTimestamp := dec.ReadInt64()
		value.UnixTimestamp = &unixTimestamp
	}
	if dec.ReadOption() {
		epoch := dec.ReadUint64()
		value.Epoch = &epoch
	}
}
