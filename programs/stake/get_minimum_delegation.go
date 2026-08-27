package stake

type GetMinimumDelegation struct{ instruction }

func NewGetMinimumDelegationInstruction() *GetMinimumDelegation { return &GetMinimumDelegation{} }

func (*GetMinimumDelegation) Data() ([]byte, error) {
	return []byte{byte(GetMinimumDelegationInstruction), 0, 0, 0}, nil
}
