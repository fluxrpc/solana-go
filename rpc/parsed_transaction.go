package rpc

import (
	"fmt"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// ParsedTransaction is a transaction in the jsonParsed encoding, with
// account keys and program-specific instruction details resolved.
type ParsedTransaction struct {
	Signatures []solana.Signature `json:"signatures"`
	Message    ParsedMessage      `json:"message"`
}

// ParsedTransactionMeta is the transaction status metadata object as
// returned with the jsonParsed transaction encoding.
type ParsedTransactionMeta struct {
	// Error if transaction failed, null if transaction succeeded.
	// https://github.com/solana-labs/solana/blob/master/sdk/src/transaction.rs#L24
	Err any `json:"err"`

	// Fee this transaction was charged
	Fee uint64 `json:"fee"`

	// Array of u64 account balances from before the transaction was processed
	PreBalances []uint64 `json:"preBalances"`

	// Array of u64 account balances after the transaction was processed
	PostBalances []uint64 `json:"postBalances"`

	// List of inner instructions or omitted if inner instruction recording
	// was not yet enabled during this transaction
	InnerInstructions []ParsedInnerInstruction `json:"innerInstructions"`

	// List of token balances from before the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PreTokenBalances []TokenBalance `json:"preTokenBalances"`

	// List of token balances from after the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PostTokenBalances []TokenBalance `json:"postTokenBalances"`

	// Array of string log messages or omitted if log message
	// recording was not yet enabled during this transaction
	LogMessages []string `json:"logMessages"`

	// DEPRECATED: Transaction status.
	Status DeprecatedTransactionMetaStatus `json:"status"`

	Rewards []BlockReward `json:"rewards"`

	LoadedAddresses LoadedAddresses `json:"loadedAddresses"`

	ReturnData ReturnData `json:"returnData"`

	ComputeUnitsConsumed *uint64 `json:"computeUnitsConsumed"`

	// Transaction cost units, as reported by newer validators.
	CostUnits *uint64 `json:"costUnits,omitempty"`
}

// ParsedInnerInstruction groups the parsed cross-program instructions
// invoked during one transaction instruction.
type ParsedInnerInstruction struct {
	Index        uint64               `json:"index"`
	Instructions []*ParsedInstruction `json:"instructions"`
}

// ParsedMessageAccount is an account entry of a parsed message.
type ParsedMessageAccount struct {
	PublicKey solana.PublicKey `json:"pubkey"`
	Signer    bool             `json:"signer"`
	Writable  bool             `json:"writable"`

	// The source of the account key: "transaction" for keys from the message
	// itself, "lookupTable" for keys resolved from address lookup tables.
	// Nil for legacy transactions.
	Source *AccountKeySource `json:"source,omitempty"`
}

// ParsedMessage is the message of a jsonParsed-encoded transaction.
type ParsedMessage struct {
	AccountKeys     []ParsedMessageAccount `json:"accountKeys"`
	Instructions    []*ParsedInstruction   `json:"instructions"`
	RecentBlockHash string                 `json:"recentBlockhash"`
}

// ParsedInstruction is an instruction of a jsonParsed-encoded transaction.
// Instructions the node could parse carry Program/Parsed; the rest carry
// Accounts/Data.
type ParsedInstruction struct {
	Program     string                   `json:"program,omitempty"`
	ProgramId   solana.PublicKey         `json:"programId,omitempty"`
	Parsed      *InstructionInfoEnvelope `json:"parsed,omitempty"`
	Data        solana.Base58            `json:"data,omitempty"`
	Accounts    []solana.PublicKey       `json:"accounts,omitempty"`
	StackHeight int64                    `json:"stackHeight"`
}

// InstructionInfoEnvelope wraps the "parsed" field of a parsed instruction,
// which is either a plain string (the instruction type) or an
// InstructionInfo object.
type InstructionInfoEnvelope struct {
	asString          string
	asInstructionInfo *InstructionInfo
}

func (wrap InstructionInfoEnvelope) MarshalJSON() ([]byte, error) {
	if wrap.asString != "" {
		return sonic.Marshal(wrap.asString)
	}
	return sonic.Marshal(wrap.asInstructionInfo)
}

func (wrap *InstructionInfoEnvelope) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || (len(data) == 4 && string(data) == "null") {
		return nil
	}

	switch data[0] {
	case '"':
		// A plain string.
		return sonic.Unmarshal(data, &wrap.asString)
	case '{':
		// An InstructionInfo object.
		return sonic.Unmarshal(data, &wrap.asInstructionInfo)
	default:
		return fmt.Errorf("unknown kind: %v", data)
	}
}

// InstructionInfo is the program-specific decoded form of an instruction.
type InstructionInfo struct {
	Info            map[string]any `json:"info"`
	InstructionType string         `json:"type"`
}
