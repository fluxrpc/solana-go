package computebudget

// Unused is the current reserved tag-zero Compute Budget variant.
type Unused struct{ instruction }

func NewUnusedInstruction() *Unused { return &Unused{} }

func (*Unused) Data() ([]byte, error) { return []byte{uint8(UnusedInstruction)}, nil }
