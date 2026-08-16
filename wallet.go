package solana_go

import "fmt"

// Wallet couples a PrivateKey with convenience constructors for the common
// ways a keypair enters a program: random generation, base58, a
// solana-keygen file, or a BIP-39 mnemonic.
type Wallet struct {
	PrivateKey PrivateKey
}

// NewWallet creates a wallet with a new random keypair. Panics only if the
// system's entropy source fails.
func NewWallet() *Wallet {
	privateKey, err := NewRandomPrivateKey()
	if err != nil {
		panic(fmt.Sprintf("failed to generate private key: %s", err))
	}
	return &Wallet{PrivateKey: privateKey}
}

// WalletFromPrivateKeyBase58 creates a wallet from a base58-encoded private
// key.
func WalletFromPrivateKeyBase58(privateKey string) (*Wallet, error) {
	k, err := PrivateKeyFromBase58(privateKey)
	if err != nil {
		return nil, fmt.Errorf("wallet from base58: %w", err)
	}
	return &Wallet{PrivateKey: k}, nil
}

// WalletFromSolanaKeygenFile creates a wallet from a keypair file written
// by `solana-keygen new`.
func WalletFromSolanaKeygenFile(file string) (*Wallet, error) {
	k, err := PrivateKeyFromSolanaKeygenFile(file)
	if err != nil {
		return nil, err
	}
	return &Wallet{PrivateKey: k}, nil
}

// NewWalletFromMnemonic creates a wallet whose private key is derived from
// an English BIP-39 mnemonic using the default Solana derivation path
// m/44'/501'/0'/0'.
func NewWalletFromMnemonic(mnemonic, passphrase string) (*Wallet, error) {
	pk, err := PrivateKeyFromMnemonic(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}
	return &Wallet{PrivateKey: pk}, nil
}

// PublicKey returns the public key of the wallet's keypair.
func (a *Wallet) PublicKey() PublicKey {
	return a.PrivateKey.PublicKey()
}
