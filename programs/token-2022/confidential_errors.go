package token2022

import "errors"

// ConfidentialTransferError classifies confidential-transfer failures without changing their messages.
type ConfidentialTransferError struct {
	err  error
	kind uint8
}

func (err *ConfidentialTransferError) Error() string {
	return err.err.Error()
}

func (err *ConfidentialTransferError) Unwrap() error {
	return err.err
}

func (err *ConfidentialTransferError) InvalidInput() bool {
	return err.kind == 1
}

func (err *ConfidentialTransferError) InvalidProof() bool {
	return err.kind == 2
}

func (err *ConfidentialTransferError) InsufficientFunds() bool {
	return err.kind == 3
}

func (err *ConfidentialTransferError) Decryption() bool {
	return err.kind == 4
}

func (service ConfidentialTransferService) confidentialTransferError(err error, kind uint8) error {
	var classified *ConfidentialTransferError
	if errors.As(err, &classified) {
		return err
	}
	return &ConfidentialTransferError{err: err, kind: kind}
}

func (service ConfidentialTransferService) invalidInputError(err error) error {
	return service.confidentialTransferError(err, 1)
}

func (service ConfidentialTransferService) invalidProofError(err error) error {
	return service.confidentialTransferError(err, 2)
}

func (service ConfidentialTransferService) classifyInvalidProof(err *error) {
	if *err != nil {
		*err = service.invalidProofError(*err)
	}
}

func (service ConfidentialTransferService) insufficientFundsError(err error) error {
	return service.confidentialTransferError(err, 3)
}

func (service ConfidentialTransferService) decryptionError(err error) error {
	return service.confidentialTransferError(err, 4)
}
