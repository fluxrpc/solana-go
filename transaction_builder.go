package solana_go

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
)

// TransactionOption configures how NewTransaction compiles instructions into
// a transaction message.
type TransactionOption interface {
	apply(opts *transactionOptions)
}

type transactionOptions struct {
	payer         PublicKey
	addressTables map[PublicKey][]PublicKey
}

type transactionOptionFunc func(opts *transactionOptions)

func (f transactionOptionFunc) apply(opts *transactionOptions) { f(opts) }

// TransactionPayer sets the fee payer account of the transaction. When not
// provided, NewTransaction falls back to the first signer account of the
// first instruction.
func TransactionPayer(payer PublicKey) TransactionOption {
	return transactionOptionFunc(func(opts *transactionOptions) { opts.payer = payer })
}

// TransactionAddressTables provides the on-chain contents of address lookup
// tables (table account -> ordered list of addresses stored in the table).
// Accounts that appear in a table and are neither signers nor invoked
// programs are moved out of the static account list and referenced through
// the table instead, producing a versioned (v0) message with
// AddressTableLookups set.
func TransactionAddressTables(tables map[PublicKey][]PublicKey) TransactionOption {
	return transactionOptionFunc(func(opts *transactionOptions) { opts.addressTables = tables })
}

// TransactionBuilder incrementally collects the pieces of a transaction and
// compiles them with Build. The zero value is not usable; create one with
// NewTransactionBuilder.
type TransactionBuilder struct {
	instructions    []Instruction
	recentBlockHash Hash
	opts            []TransactionOption
}

// NewTransactionBuilder creates a new empty transaction builder.
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{}
}

// AddInstruction appends the given instruction to the transaction.
func (builder *TransactionBuilder) AddInstruction(instruction Instruction) *TransactionBuilder {
	builder.instructions = append(builder.instructions, instruction)
	return builder
}

// SetRecentBlockHash sets the recent blockhash of the transaction.
func (builder *TransactionBuilder) SetRecentBlockHash(recentBlockHash Hash) *TransactionBuilder {
	builder.recentBlockHash = recentBlockHash
	return builder
}

// WithOpt adds a TransactionOption applied on Build.
func (builder *TransactionBuilder) WithOpt(opt TransactionOption) *TransactionBuilder {
	builder.opts = append(builder.opts, opt)
	return builder
}

// SetFeePayer sets the fee payer account of the transaction.
func (builder *TransactionBuilder) SetFeePayer(feePayer PublicKey) *TransactionBuilder {
	builder.opts = append(builder.opts, TransactionPayer(feePayer))
	return builder
}

// Build compiles the accumulated instructions into an unsigned Transaction.
func (builder *TransactionBuilder) Build() (*Transaction, error) {
	return NewTransaction(builder.instructions, builder.recentBlockHash, builder.opts...)
}

// lookupTableEntry locates one address inside a lookup table: which table
// account holds it and at which position.
type lookupTableEntry struct {
	table PublicKey
	index uint8
}

// loadedAccount is one account the compiled message loads through an address
// lookup table instead of listing statically.
type loadedAccount struct {
	key      PublicKey
	table    PublicKey
	index    uint8
	writable bool
}

// NewTransaction compiles instructions into an unsigned Transaction.
//
// Account keys are deduplicated and ordered as the runtime requires: fee
// payer first, then remaining signers before non-signers, writable before
// readonly within each class, ties broken by pubkey. An account required as
// writable by any instruction is writable in the message.
//
// The fee payer comes from the TransactionPayer option, falling back to the
// first signer of the first instruction. With TransactionAddressTables,
// eligible accounts are compressed into AddressTableLookups and the message
// becomes v0; otherwise a legacy message is produced. The output is
// byte-identical to upstream solana-go's NewTransaction for the same inputs.
//
// Sign the result with (*Transaction).Sign before submitting it.
func NewTransaction(instructions []Instruction, recentBlockHash Hash, opts ...TransactionOption) (*Transaction, error) {
	if len(instructions) == 0 {
		return nil, errors.New("requires at least one instruction to create a transaction")
	}

	var options transactionOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	feePayer := options.payer
	if feePayer.IsZero() {
		found := false
		for _, meta := range instructions[0].Accounts() {
			if meta.IsSigner {
				feePayer = meta.PublicKey
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("cannot determine fee payer: pass the TransactionPayer option, or make a signer the first instruction's first account")
		}
	}

	// Index every address stored in the lookup tables: address -> (table,
	// position). Tables are visited in sorted key order so that an address
	// present in several tables deterministically resolves to the smallest
	// table key.
	var lookupEntries map[PublicKey]lookupTableEntry
	if len(options.addressTables) > 0 {
		totalTableEntries := 0
		for _, table := range options.addressTables {
			totalTableEntries += len(table)
		}
		lookupEntries = make(map[PublicKey]lookupTableEntry, totalTableEntries)
		sortedTableKeys := make([]PublicKey, 0, len(options.addressTables))
		for k := range options.addressTables {
			sortedTableKeys = append(sortedTableKeys, k)
		}
		slices.SortFunc(sortedTableKeys, func(a, b PublicKey) int {
			return bytes.Compare(a[:], b[:])
		})
		for _, tableKey := range sortedTableKeys {
			table := options.addressTables[tableKey]
			if len(table) > 256 {
				return nil, fmt.Errorf("max lookup table index exceeded for %s table", tableKey)
			}
			for i, address := range table {
				if _, ok := lookupEntries[address]; ok {
					continue
				}
				lookupEntries[address] = lookupTableEntry{table: tableKey, index: uint8(i)}
			}
		}
	}

	// Collect every account meta by value (never mutating the caller's
	// metas), then the invoked program IDs as readonly non-signers.
	totalAccounts := 0
	for _, instruction := range instructions {
		totalAccounts += len(instruction.Accounts())
	}
	metas := make([]AccountMeta, 0, totalAccounts+len(instructions))
	programIDs := make([]PublicKey, 0, len(instructions))
	for _, instruction := range instructions {
		for _, meta := range instruction.Accounts() {
			metas = append(metas, *meta)
		}
		if programID := instruction.ProgramID(); !slices.Contains(programIDs, programID) {
			programIDs = append(programIDs, programID)
		}
	}
	for _, programID := range programIDs {
		metas = append(metas, AccountMeta{PublicKey: programID})
	}

	// Sort: signers first, then writable, ties broken by pubkey. The sort is
	// stable, but the pubkey tiebreak makes the order fully deterministic.
	slices.SortStableFunc(metas, func(a, b AccountMeta) int {
		if a.IsSigner != b.IsSigner {
			if a.IsSigner {
				return -1
			}
			return 1
		}
		if a.IsWritable != b.IsWritable {
			if a.IsWritable {
				return -1
			}
			return 1
		}
		return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
	})

	// Deduplicate in place. The first occurrence of a key has the strongest
	// signer class (signers sorted first); writability is OR-ed across
	// duplicates.
	uniqIndex := make(map[PublicKey]int, len(metas))
	uniq := metas[:0]
	for _, meta := range metas {
		if at, ok := uniqIndex[meta.PublicKey]; ok {
			uniq[at].IsWritable = uniq[at].IsWritable || meta.IsWritable
			continue
		}
		uniqIndex[meta.PublicKey] = len(uniq)
		uniq = append(uniq, meta)
	}

	// Move the fee payer to the front (adding it if absent), forcing it
	// signer and writable. When the payer is already present this rotates
	// the prefix in place; only an absent payer costs a new slice.
	allKeys := uniq
	if feePayerAt, ok := uniqIndex[feePayer]; ok {
		copy(allKeys[1:feePayerAt+1], allKeys[:feePayerAt])
	} else {
		allKeys = make([]AccountMeta, len(uniq)+1)
		copy(allKeys[1:], uniq)
	}
	allKeys[0] = AccountMeta{PublicKey: feePayer, IsSigner: true, IsWritable: true}

	message := Message{
		RecentBlockhash: recentBlockHash,
	}

	// Split the ordered keys into static message keys and table-loaded keys,
	// assigning static indexes as we go. Only accounts that live in a lookup
	// table, are not signers, and are not invoked as programs can be loaded
	// through a table.
	keyIndex := make(map[PublicKey]uint16, len(allKeys))
	message.AccountKeys = make([]PublicKey, 0, len(allKeys))
	var loaded []loadedAccount
	next := uint16(0)
	for idx, meta := range allKeys {
		if entry, inTable := lookupEntries[meta.PublicKey]; inTable &&
			idx != 0 && !meta.IsSigner && !slices.Contains(programIDs, meta.PublicKey) {
			loaded = append(loaded, loadedAccount{
				key:      meta.PublicKey,
				table:    entry.table,
				index:    entry.index,
				writable: meta.IsWritable,
			})
			continue
		}

		message.AccountKeys = append(message.AccountKeys, meta.PublicKey)
		keyIndex[meta.PublicKey] = next
		next++
		if meta.IsSigner {
			message.Header.NumRequiredSignatures++
			if !meta.IsWritable {
				message.Header.NumReadonlySignedAccounts++
			}
			continue
		}
		if !meta.IsWritable {
			message.Header.NumReadonlyUnsignedAccounts++
		}
	}

	// Table-loaded keys are indexed after the static keys: all writable
	// lookups first, then all readonly lookups, tables visited in sorted key
	// order. One backing array serves every lookup's index slices.
	if len(loaded) > 0 {
		tableKeys := make([]PublicKey, 0, len(options.addressTables))
		for _, l := range loaded {
			if !slices.Contains(tableKeys, l.table) {
				tableKeys = append(tableKeys, l.table)
			}
		}
		slices.SortFunc(tableKeys, func(a, b PublicKey) int {
			return bytes.Compare(a[:], b[:])
		})
		tableLookups := make([]MessageAddressTableLookup, len(tableKeys))
		idxBacking := make([]uint8, len(loaded))
		pos := 0
		for ti, tableKey := range tableKeys {
			start := pos
			for _, l := range loaded {
				if l.table == tableKey && l.writable {
					idxBacking[pos] = l.index
					keyIndex[l.key] = next
					next++
					pos++
				}
			}
			tableLookups[ti].AccountKey = tableKey
			if pos > start {
				tableLookups[ti].WritableIndexes = idxBacking[start:pos:pos]
			}
		}
		for ti, tableKey := range tableKeys {
			start := pos
			for _, l := range loaded {
				if l.table == tableKey && !l.writable {
					idxBacking[pos] = l.index
					keyIndex[l.key] = next
					next++
					pos++
				}
			}
			if pos > start {
				tableLookups[ti].ReadonlyIndexes = idxBacking[start:pos:pos]
			}
		}
		message.SetAddressTableLookups(tableLookups)
	}

	message.Instructions = make([]CompiledInstruction, 0, len(instructions))
	for txIdx, instruction := range instructions {
		accounts := instruction.Accounts()
		accountIndexes := make([]uint16, len(accounts))
		for idx, meta := range accounts {
			accountIndexes[idx] = keyIndex[meta.PublicKey]
		}
		data, err := instruction.Data()
		if err != nil {
			return nil, fmt.Errorf("encoding instruction [%d] data: %w", txIdx, err)
		}
		message.Instructions = append(message.Instructions, CompiledInstruction{
			ProgramIDIndex: keyIndex[instruction.ProgramID()],
			Accounts:       accountIndexes,
			Data:           data,
		})
	}

	return &Transaction{Message: message}, nil
}
