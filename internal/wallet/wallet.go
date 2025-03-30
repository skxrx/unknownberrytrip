package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unknownberrytrip/internal/transaction"
	"unknownberrytrip/internal/utils"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

// NewWallet creates a new wallet
func NewWallet() *Wallet {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := privateKey.PublicKey
	address := PubKeyToAddress(&publicKey)
	return &Wallet{privateKey, &publicKey, address}
}

// PubKeyToAddress generates an address from a public key
func PubKeyToAddress(pubKey *ecdsa.PublicKey) string {
	pubBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	hash := sha256.Sum256(pubBytes)
	return hex.EncodeToString(hash[:])
}

// SignTx signs a transaction
func (w *Wallet) SignTx(tx *transaction.Transaction) string {
	txForSign := struct {
		From   string
		To     string
		Amount float64
		Nonce  int
	}{tx.From, tx.To, tx.Amount, tx.Nonce}
	data, _ := json.Marshal(txForSign)
	hash := sha256.Sum256(data)
	r, s, _ := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature)
}

// CreateTransaction creates and signs a transaction
func (w *Wallet) CreateTransaction(to string, amount float64, nonce int) *transaction.Transaction {
	tx := &transaction.Transaction{
		From:   w.Address,
		To:     to,
		Amount: amount,
		Nonce:  nonce + 1,
		PubKey: utils.PubKeyToString(w.PublicKey),
	}
	tx.Signature = w.SignTx(tx)
	return tx
}
