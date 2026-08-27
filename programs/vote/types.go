package vote

import (
	"errors"
	"fmt"
	"math"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type instruction struct{ solana.AccountMetaSlice }

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

type InstructionType uint32

func (typ InstructionType) String() string {
	names := [...]string{
		"InitializeAccount", "Authorize", "Vote", "Withdraw", "UpdateValidatorIdentity",
		"UpdateCommission", "VoteSwitch", "AuthorizeChecked", "UpdateVoteState",
		"UpdateVoteStateSwitch", "AuthorizeWithSeed", "AuthorizeCheckedWithSeed",
		"CompactUpdateVoteState", "CompactUpdateVoteStateSwitch", "TowerSync",
		"TowerSyncSwitch", "InitializeAccountV2", "UpdateCommissionCollector",
		"UpdateCommissionBps", "DepositDelegatorRewards",
	}
	if uint32(typ) < uint32(len(names)) {
		return names[typ]
	}
	return fmt.Sprintf("InstructionType(%d)", uint32(typ))
}

// DecodedInstruction is the fully typed result of DecodeInstruction. Type
// identifies the one non-nil instruction field.
type DecodedInstruction struct {
	Type                         InstructionType
	InitializeAccount            *InitializeAccount
	Authorize                    *Authorize
	Vote                         *Vote
	Withdraw                     *Withdraw
	UpdateValidatorIdentity      *UpdateValidatorIdentity
	UpdateCommission             *UpdateCommission
	VoteSwitch                   *VoteSwitch
	AuthorizeChecked             *AuthorizeChecked
	UpdateVoteState              *UpdateVoteState
	UpdateVoteStateSwitch        *UpdateVoteStateSwitch
	AuthorizeWithSeed            *AuthorizeWithSeed
	AuthorizeCheckedWithSeed     *AuthorizeCheckedWithSeed
	CompactUpdateVoteState       *CompactUpdateVoteState
	CompactUpdateVoteStateSwitch *CompactUpdateVoteStateSwitch
	TowerSync                    *TowerSync
	TowerSyncSwitch              *TowerSyncSwitch
	InitializeAccountV2          *InitializeAccountV2
	UpdateCommissionCollector    *UpdateCommissionCollector
	UpdateCommissionBps          *UpdateCommissionBps
	DepositDelegatorRewards      *DepositDelegatorRewards
}

// VoteAuthorizeKind is the bincode enum discriminator for vote authority.
type VoteAuthorizeKind uint32

// VoteAuthorize is the current VoteAuthorize wire enum. The BLS arrays are
// required only for VoteAuthorizeVoterWithBLS.
type VoteAuthorize struct {
	Kind                 VoteAuthorizeKind
	BLSPubkey            *[BLSPublicKeyCompressedSize]byte
	BLSProofOfPossession *[BLSProofOfPossessionCompressedSize]byte
}

func (value VoteAuthorize) write(enc *bin.Encoder) error {
	if value.Kind > VoteAuthorizeVoterWithBLS {
		return fmt.Errorf("%w: %d", ErrInvalidVoteAuthorize, value.Kind)
	}
	enc.WriteUint32(uint32(value.Kind))
	if value.Kind == VoteAuthorizeVoterWithBLS {
		if value.BLSPubkey == nil || value.BLSProofOfPossession == nil {
			return errors.New("vote: VoterWithBLS requires a public key and proof of possession")
		}
		enc.WriteBytes(value.BLSPubkey[:])
		enc.WriteBytes(value.BLSProofOfPossession[:])
	}
	return enc.Err()
}

func (value *VoteAuthorize) decode(dec *bin.Decoder) error {
	value.Kind = VoteAuthorizeKind(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return err
	}
	switch value.Kind {
	case VoteAuthorizeVoter, VoteAuthorizeWithdrawer:
		return nil
	case VoteAuthorizeVoterWithBLS:
		key := new([BLSPublicKeyCompressedSize]byte)
		proof := new([BLSProofOfPossessionCompressedSize]byte)
		copy(key[:], dec.ReadBytes(BLSPublicKeyCompressedSize))
		copy(proof[:], dec.ReadBytes(BLSProofOfPossessionCompressedSize))
		value.BLSPubkey, value.BLSProofOfPossession = key, proof
		return dec.Err()
	default:
		return fmt.Errorf("%w: %d", ErrInvalidVoteAuthorize, value.Kind)
	}
}

// CommissionKind is kept compact in Go, but its bincode enum discriminator is
// encoded as a little-endian uint32 by the instruction codecs.
type CommissionKind uint8

func (kind CommissionKind) validate() error {
	if kind > CommissionKindBlockRevenue {
		return fmt.Errorf("%w: %d", ErrInvalidCommissionKind, kind)
	}
	return nil
}

type VoteInit struct {
	NodePubkey           solana.PublicKey
	AuthorizedVoter      solana.PublicKey
	AuthorizedWithdrawer solana.PublicKey
	Commission           uint8
}

func (value VoteInit) write(enc *bin.Encoder) {
	enc.WritePublicKey(value.NodePubkey)
	enc.WritePublicKey(value.AuthorizedVoter)
	enc.WritePublicKey(value.AuthorizedWithdrawer)
	enc.WriteUint8(value.Commission)
}

func (value *VoteInit) decode(dec *bin.Decoder) {
	value.NodePubkey = dec.ReadPublicKey()
	value.AuthorizedVoter = dec.ReadPublicKey()
	value.AuthorizedWithdrawer = dec.ReadPublicKey()
	value.Commission = dec.ReadUint8()
}

type VoteInitV2 struct {
	NodePubkey                          solana.PublicKey
	AuthorizedVoter                     solana.PublicKey
	AuthorizedVoterBLSPubkey            [BLSPublicKeyCompressedSize]byte
	AuthorizedVoterBLSProofOfPossession [BLSProofOfPossessionCompressedSize]byte
	AuthorizedWithdrawer                solana.PublicKey
	InflationRewardsCommissionBps       uint16
	BlockRevenueCommissionBps           uint16
}

func (value VoteInitV2) write(enc *bin.Encoder) {
	enc.WritePublicKey(value.NodePubkey)
	enc.WritePublicKey(value.AuthorizedVoter)
	enc.WriteBytes(value.AuthorizedVoterBLSPubkey[:])
	enc.WriteBytes(value.AuthorizedVoterBLSProofOfPossession[:])
	enc.WritePublicKey(value.AuthorizedWithdrawer)
	enc.WriteUint16(value.InflationRewardsCommissionBps)
	enc.WriteUint16(value.BlockRevenueCommissionBps)
}

func (value *VoteInitV2) decode(dec *bin.Decoder) {
	value.NodePubkey = dec.ReadPublicKey()
	value.AuthorizedVoter = dec.ReadPublicKey()
	copy(value.AuthorizedVoterBLSPubkey[:], dec.ReadBytes(BLSPublicKeyCompressedSize))
	copy(value.AuthorizedVoterBLSProofOfPossession[:], dec.ReadBytes(BLSProofOfPossessionCompressedSize))
	value.AuthorizedWithdrawer = dec.ReadPublicKey()
	value.InflationRewardsCommissionBps = dec.ReadUint16()
	value.BlockRevenueCommissionBps = dec.ReadUint16()
}

type Lockout struct {
	Slot              uint64
	ConfirmationCount uint32
}

type voteData struct {
	Slots     []uint64
	Hash      solana.Hash
	Timestamp *int64
}

func (vote voteData) write(enc *bin.Encoder) error {
	if len(vote.Slots) > MaxLockoutHistory {
		return fmt.Errorf("%w: %d > %d", ErrTooManyLockouts, len(vote.Slots), MaxLockoutHistory)
	}
	enc.WriteUint64(uint64(len(vote.Slots)))
	for _, slot := range vote.Slots {
		enc.WriteUint64(slot)
	}
	enc.WriteHash(vote.Hash)
	enc.WriteOption(vote.Timestamp != nil)
	if vote.Timestamp != nil {
		enc.WriteInt64(*vote.Timestamp)
	}
	return enc.Err()
}

func (vote *voteData) decode(dec *bin.Decoder) error {
	count := dec.ReadUint64()
	if err := dec.Err(); err != nil {
		return err
	}
	if count > MaxLockoutHistory {
		return fmt.Errorf("%w: %d > %d", ErrTooManyLockouts, count, MaxLockoutHistory)
	}

	if uint64(dec.Remaining()) < count*8+33 {
		return fmt.Errorf("vote: truncated slots: %w", bin.ErrUnexpectedEOF)
	}
	vote.Slots = make([]uint64, int(count))
	for i := range vote.Slots {
		vote.Slots[i] = dec.ReadUint64()
	}
	vote.Hash = dec.ReadHash()
	if dec.ReadOption() {
		timestamp := dec.ReadInt64()
		vote.Timestamp = &timestamp
	}
	return dec.Err()
}

type VoteStateUpdate struct {
	Lockouts  []Lockout
	Root      *uint64
	Hash      solana.Hash
	Timestamp *int64
}

type TowerSyncUpdate struct {
	Lockouts  []Lockout
	Root      *uint64
	Hash      solana.Hash
	Timestamp *int64
	BlockID   solana.Hash
}

type VoteAuthorizeWithSeedArgs struct {
	AuthorizationType               VoteAuthorize
	CurrentAuthorityDerivedKeyOwner solana.PublicKey
	CurrentAuthorityDerivedKeySeed  string
	NewAuthority                    solana.PublicKey
}

type VoteAuthorizeCheckedWithSeedArgs struct {
	AuthorizationType               VoteAuthorize
	CurrentAuthorityDerivedKeyOwner solana.PublicKey
	CurrentAuthorityDerivedKeySeed  string
}

func (update VoteStateUpdate) validate() error {
	if len(update.Lockouts) > MaxLockoutHistory {
		return fmt.Errorf("%w: %d > %d", ErrTooManyLockouts, len(update.Lockouts), MaxLockoutHistory)
	}
	return nil
}

func (update VoteStateUpdate) write(enc *bin.Encoder) error {
	if err := update.validate(); err != nil {
		return err
	}
	enc.WriteUint64(uint64(len(update.Lockouts)))
	for _, lockout := range update.Lockouts {
		enc.WriteUint64(lockout.Slot)
		enc.WriteUint32(lockout.ConfirmationCount)
	}
	enc.WriteOption(update.Root != nil)
	if update.Root != nil {
		enc.WriteUint64(*update.Root)
	}
	enc.WriteHash(update.Hash)
	enc.WriteOption(update.Timestamp != nil)
	if update.Timestamp != nil {
		enc.WriteInt64(*update.Timestamp)
	}
	return enc.Err()
}

func (value *VoteStateUpdate) decode(dec *bin.Decoder) error {
	count := dec.ReadUint64()
	if err := dec.Err(); err != nil {
		return err
	}
	if count > MaxLockoutHistory {
		return fmt.Errorf("%w: %d > %d", ErrTooManyLockouts, count, MaxLockoutHistory)
	}
	if uint64(dec.Remaining()) < count*12+34 {
		return fmt.Errorf("vote: truncated lockouts: %w", bin.ErrUnexpectedEOF)
	}
	value.Lockouts = make([]Lockout, int(count))
	for index := range value.Lockouts {
		value.Lockouts[index] = Lockout{Slot: dec.ReadUint64(), ConfirmationCount: dec.ReadUint32()}
	}
	if dec.ReadOption() {
		root := dec.ReadUint64()
		value.Root = &root
	}
	value.Hash = dec.ReadHash()
	if dec.ReadOption() {
		timestamp := dec.ReadInt64()
		value.Timestamp = &timestamp
	}
	return dec.Err()
}

func (update VoteStateUpdate) writeCompact(enc *bin.Encoder) error {
	if err := update.validate(); err != nil {
		return err
	}
	root := uint64(math.MaxUint64)
	if update.Root != nil {
		root = *update.Root
	}
	enc.WriteUint64(root)
	enc.WriteCompactU16(len(update.Lockouts))
	previous := uint64(0)
	if update.Root != nil {
		previous = *update.Root
	}
	for _, lockout := range update.Lockouts {
		if lockout.Slot < previous {
			return fmt.Errorf("%w: %d after %d", ErrInvalidLockoutOrder, lockout.Slot, previous)
		}
		delta := lockout.Slot - previous
		if lockout.ConfirmationCount > math.MaxUint8 {
			return fmt.Errorf("%w: %d", ErrInvalidConfirmationCount, lockout.ConfirmationCount)
		}
		enc.WriteVarUint64(delta)
		enc.WriteUint8(uint8(lockout.ConfirmationCount))
		previous = lockout.Slot
	}
	enc.WriteHash(update.Hash)
	enc.WriteOption(update.Timestamp != nil)
	if update.Timestamp != nil {
		enc.WriteInt64(*update.Timestamp)
	}
	return enc.Err()
}

func (value *VoteStateUpdate) decodeCompact(dec *bin.Decoder) error {
	root := dec.ReadUint64()
	count := dec.ReadCompactU16()
	if err := dec.Err(); err != nil {
		return err
	}
	if count > MaxLockoutHistory {
		return fmt.Errorf("%w: %d > %d", ErrTooManyLockouts, count, MaxLockoutHistory)
	}
	if dec.Remaining() < count*2+33 {
		return fmt.Errorf("vote: truncated compact lockouts: %w", bin.ErrUnexpectedEOF)
	}
	value.Lockouts = make([]Lockout, count)
	if root != math.MaxUint64 {
		value.Root = &root
	}
	previous := uint64(0)
	if value.Root != nil {
		previous = root
	}
	for index := range value.Lockouts {
		delta := dec.ReadVarUint64()
		confirmation := dec.ReadUint8()
		if err := dec.Err(); err != nil {
			return err
		}
		if math.MaxUint64-previous < delta {
			return ErrSlotOverflow
		}
		previous += delta
		value.Lockouts[index] = Lockout{Slot: previous, ConfirmationCount: uint32(confirmation)}
	}
	value.Hash = dec.ReadHash()
	if dec.ReadOption() {
		timestamp := dec.ReadInt64()
		value.Timestamp = &timestamp
	}
	return dec.Err()
}

func (sync TowerSyncUpdate) write(enc *bin.Encoder) error {
	if err := (VoteStateUpdate{
		Lockouts: sync.Lockouts, Root: sync.Root, Hash: sync.Hash, Timestamp: sync.Timestamp,
	}).writeCompact(enc); err != nil {
		return err
	}
	enc.WriteHash(sync.BlockID)
	return enc.Err()
}

func (sync *TowerSyncUpdate) decode(dec *bin.Decoder) error {
	var update VoteStateUpdate
	if err := update.decodeCompact(dec); err != nil {
		return err
	}
	sync.Lockouts = update.Lockouts
	sync.Root = update.Root
	sync.Hash = update.Hash
	sync.Timestamp = update.Timestamp
	sync.BlockID = dec.ReadHash()
	return dec.Err()
}

func (value VoteAuthorizeWithSeedArgs) write(enc *bin.Encoder) error {
	if err := value.AuthorizationType.write(enc); err != nil {
		return err
	}
	enc.WritePublicKey(value.CurrentAuthorityDerivedKeyOwner)
	enc.WriteBincodeString(value.CurrentAuthorityDerivedKeySeed)
	enc.WritePublicKey(value.NewAuthority)
	return enc.Err()
}

func (value *VoteAuthorizeWithSeedArgs) decode(dec *bin.Decoder) error {
	if err := value.AuthorizationType.decode(dec); err != nil {
		return err
	}
	value.CurrentAuthorityDerivedKeyOwner = dec.ReadPublicKey()
	value.CurrentAuthorityDerivedKeySeed = dec.ReadBincodeString()
	value.NewAuthority = dec.ReadPublicKey()
	return dec.Err()
}

func (value VoteAuthorizeCheckedWithSeedArgs) write(enc *bin.Encoder) error {
	if err := value.AuthorizationType.write(enc); err != nil {
		return err
	}
	enc.WritePublicKey(value.CurrentAuthorityDerivedKeyOwner)
	enc.WriteBincodeString(value.CurrentAuthorityDerivedKeySeed)
	return enc.Err()
}

func (value *VoteAuthorizeCheckedWithSeedArgs) decode(dec *bin.Decoder) error {
	if err := value.AuthorizationType.decode(dec); err != nil {
		return err
	}
	value.CurrentAuthorityDerivedKeyOwner = dec.ReadPublicKey()
	value.CurrentAuthorityDerivedKeySeed = dec.ReadBincodeString()
	return dec.Err()
}
