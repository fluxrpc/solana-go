package rpc

// GetSignatureStatusesResult is the result of the getSignatureStatuses method.
type GetSignatureStatusesResult struct {
	RPCContext
	Value []*SignatureStatusesResult `json:"value"`
}

// SignatureStatusesResult is the status of a single queried signature.
type SignatureStatusesResult struct {
	// The slot the transaction was processed.
	Slot uint64 `json:"slot"`

	// Number of blocks since signature confirmation,
	// null if rooted or finalized by a supermajority of the cluster.
	Confirmations *uint64 `json:"confirmations"`

	// Error if transaction failed, null if transaction succeeded.
	Err any `json:"err"`

	// The transaction's cluster confirmation status; either processed, confirmed, or finalized.
	ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus"`

	// DEPRECATED: Transaction status.
	Status DeprecatedTransactionMetaStatus `json:"status"`
}
