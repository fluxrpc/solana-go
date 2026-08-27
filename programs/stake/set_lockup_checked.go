package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type SetLockupChecked struct {
	Args LockupCheckedArgs
	instruction
}

func NewSetLockupCheckedInstruction(args LockupCheckedArgs, stakeAccount, authority solana.PublicKey, newCustodian *solana.PublicKey) *SetLockupChecked {
	accounts := make(solana.AccountMetaSlice, 0, 3)
	accounts = append(accounts,
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(authority, false, true),
	)
	if newCustodian != nil {
		accounts = append(
			accounts,
			solana.NewAccountMeta(*newCustodian, false, true),
		)
	}
	return &SetLockupChecked{Args: args, instruction: instruction{AccountMetaSlice: accounts}}
}

func (inst *SetLockupChecked) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 22))
	enc.WriteUint32(uint32(SetLockupCheckedInstruction))
	inst.Args.write(enc)
	return enc.Bytes(), enc.Err()
}
