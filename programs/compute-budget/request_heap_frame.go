package computebudget

import (
	"github.com/fluxrpc/solana-go/binary"
)

// RequestHeapFrame requests a transaction-wide program heap size in bytes.
type RequestHeapFrame struct {
	instruction
	HeapSize uint32
}

func NewRequestHeapFrameInstruction(heapSize uint32) *RequestHeapFrame {
	return &RequestHeapFrame{HeapSize: heapSize}
}

func (inst *RequestHeapFrame) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 5))
	enc.WriteUint8(uint8(RequestHeapFrameInstruction))
	enc.WriteUint32(inst.HeapSize)
	return enc.Bytes(), enc.Err()
}
