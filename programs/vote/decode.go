package vote

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// DecodeInstruction decodes a Vote Program instruction and retains accounts.
// Trailing bytes are accepted to match limited bincode deserialization.
func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("vote instruction type: %w", err)
	}
	out := DecodedInstruction{Type: typ}
	var err error
	switch typ {
	case InitializeAccountInstruction:
		out.InitializeAccount = &InitializeAccount{instruction: instruction{accounts}}
		out.InitializeAccount.VoteInit.decode(dec)
	case AuthorizeInstruction:
		out.Authorize = &Authorize{
			NewAuthority: dec.ReadPublicKey(),
			instruction:  instruction{accounts},
		}
		err = out.Authorize.AuthorizationType.decode(dec)
	case VoteInstruction:
		var vote voteData
		err = vote.decode(dec)
		out.Vote = &Vote{
			Slots:       vote.Slots,
			Hash:        vote.Hash,
			Timestamp:   vote.Timestamp,
			instruction: instruction{accounts},
		}
	case WithdrawInstruction:
		out.Withdraw = &Withdraw{Lamports: dec.ReadUint64(), instruction: instruction{accounts}}
	case UpdateValidatorIdentityInstruction:
		out.UpdateValidatorIdentity = &UpdateValidatorIdentity{instruction{accounts}}
	case UpdateCommissionInstruction:
		out.UpdateCommission = &UpdateCommission{
			Commission:  dec.ReadUint8(),
			instruction: instruction{accounts},
		}
	case VoteSwitchInstruction:
		var vote voteData
		if err = vote.decode(dec); err == nil {
			out.VoteSwitch = &VoteSwitch{
				Slots:       vote.Slots,
				VoteHash:    vote.Hash,
				Timestamp:   vote.Timestamp,
				ProofHash:   dec.ReadHash(),
				instruction: instruction{accounts},
			}
		}
	case AuthorizeCheckedInstruction:
		out.AuthorizeChecked = &AuthorizeChecked{instruction: instruction{accounts}}
		err = out.AuthorizeChecked.AuthorizationType.decode(dec)
	case UpdateVoteStateInstruction:
		out.UpdateVoteState = &UpdateVoteState{instruction: instruction{accounts}}
		err = out.UpdateVoteState.VoteStateUpdate.decode(dec)
	case UpdateVoteStateSwitchInstruction:
		out.UpdateVoteStateSwitch = &UpdateVoteStateSwitch{instruction: instruction{accounts}}
		if err = out.UpdateVoteStateSwitch.VoteStateUpdate.decode(dec); err == nil {
			out.UpdateVoteStateSwitch.ProofHash = dec.ReadHash()
		}
	case AuthorizeWithSeedInstruction:
		out.AuthorizeWithSeed = &AuthorizeWithSeed{instruction: instruction{accounts}}
		err = out.AuthorizeWithSeed.VoteAuthorizeWithSeedArgs.decode(dec)
	case AuthorizeCheckedWithSeedInstruction:
		out.AuthorizeCheckedWithSeed = &AuthorizeCheckedWithSeed{instruction: instruction{accounts}}
		err = out.AuthorizeCheckedWithSeed.VoteAuthorizeCheckedWithSeedArgs.decode(dec)
	case CompactUpdateVoteStateInstruction:
		out.CompactUpdateVoteState = &CompactUpdateVoteState{instruction: instruction{accounts}}
		err = out.CompactUpdateVoteState.VoteStateUpdate.decodeCompact(dec)
	case CompactUpdateVoteStateSwitchInstruction:
		out.CompactUpdateVoteStateSwitch = &CompactUpdateVoteStateSwitch{instruction: instruction{accounts}}
		if err = out.CompactUpdateVoteStateSwitch.VoteStateUpdate.decodeCompact(dec); err == nil {
			out.CompactUpdateVoteStateSwitch.ProofHash = dec.ReadHash()
		}
	case TowerSyncInstruction:
		out.TowerSync = &TowerSync{instruction: instruction{accounts}}
		err = out.TowerSync.TowerSyncUpdate.decode(dec)
	case TowerSyncSwitchInstruction:
		out.TowerSyncSwitch = &TowerSyncSwitch{instruction: instruction{accounts}}
		if err = out.TowerSyncSwitch.TowerSyncUpdate.decode(dec); err == nil {
			out.TowerSyncSwitch.ProofHash = dec.ReadHash()
		}
	case InitializeAccountV2Instruction:
		out.InitializeAccountV2 = &InitializeAccountV2{instruction: instruction{accounts}}
		out.InitializeAccountV2.VoteInitV2.decode(dec)
	case UpdateCommissionCollectorInstruction:
		raw := dec.ReadUint32()
		if raw > uint32(CommissionKindBlockRevenue) {
			err = fmt.Errorf("%w: %d", ErrInvalidCommissionKind, raw)
		} else {
			out.UpdateCommissionCollector = &UpdateCommissionCollector{
				Kind:        CommissionKind(raw),
				instruction: instruction{accounts},
			}
		}
	case UpdateCommissionBpsInstruction:
		bps := dec.ReadUint16()
		raw := dec.ReadUint32()
		if raw > uint32(CommissionKindBlockRevenue) {
			err = fmt.Errorf("%w: %d", ErrInvalidCommissionKind, raw)
		} else {
			out.UpdateCommissionBps = &UpdateCommissionBps{
				CommissionBps: bps,
				Kind:          CommissionKind(raw),
				instruction:   instruction{accounts},
			}
		}
	case DepositDelegatorRewardsInstruction:
		out.DepositDelegatorRewards = &DepositDelegatorRewards{
			Deposit:     dec.ReadUint64(),
			instruction: instruction{accounts},
		}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint32(typ))
	}
	if err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode vote %s: %w", typ, err)
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode vote %s: %w", typ, err)
	}
	switch typ {
	case AuthorizeWithSeedInstruction:
		if len(out.AuthorizeWithSeed.CurrentAuthorityDerivedKeySeed) > solana.MaxSeedLength {
			return DecodedInstruction{}, fmt.Errorf("decode vote %s: %w", typ, solana.ErrMaxSeedLengthExceeded)
		}
	case AuthorizeCheckedWithSeedInstruction:
		if len(out.AuthorizeCheckedWithSeed.CurrentAuthorityDerivedKeySeed) > solana.MaxSeedLength {
			return DecodedInstruction{}, fmt.Errorf("decode vote %s: %w", typ, solana.ErrMaxSeedLengthExceeded)
		}
	}
	return out, nil
}
