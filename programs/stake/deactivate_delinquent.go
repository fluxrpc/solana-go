package stake

import solana "github.com/fluxrpc/solana-go"

type DeactivateDelinquent struct{ instruction }

func NewDeactivateDelinquentInstruction(stakeAccount, delinquentVoteAccount, referenceVoteAccount solana.PublicKey) *DeactivateDelinquent {
	return &DeactivateDelinquent{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(delinquentVoteAccount, false, false),
		solana.NewAccountMeta(referenceVoteAccount, false, false),
	}}}
}

func (*DeactivateDelinquent) Data() ([]byte, error) {
	return []byte{byte(DeactivateDelinquentInstruction), 0, 0, 0}, nil
}
