package transaction

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"unknownberrytrip/internal/utils"
)

type Transaction struct {
	From            string  // Sender address
	To              string  // Recipient address
	Amount          float64 // Amount
	Nonce           int     // Transaction counter
	ExtraPower      int     // Processing priority
	IsTokenTransfer bool    // Token transfer flag
	TokenID         string  // Token ID (if IsTokenTransfer = true)
	PubKey          string  // Sender's public key
	Signature       string  // Signature
}

// VerifyTxSignature verifies the transaction signature
func VerifyTxSignature(tx *Transaction) bool {
	pubKey := utils.StringToPubKey(tx.PubKey)
	txForSign := struct {
		From            string
		To              string
		Amount          float64
		Nonce           int
		ExtraPower      int
		IsTokenTransfer bool
		TokenID         string
	}{tx.From, tx.To, tx.Amount, tx.Nonce, tx.ExtraPower, tx.IsTokenTransfer, tx.TokenID}
	data, _ := json.Marshal(txForSign)
	hash := sha256.Sum256(data)
	sigBytes, _ := hex.DecodeString(tx.Signature)
	r := big.NewInt(0).SetBytes(sigBytes[:len(sigBytes)/2])
	s := big.NewInt(0).SetBytes(sigBytes[len(sigBytes)/2:])
	return ecdsa.Verify(pubKey, hash[:], r, s)
}

// Hash computes the transaction hash
func (tx *Transaction) Hash() string {
	record := fmt.Sprintf("%s%s%f%d%d%t%s", tx.From, tx.To, tx.Amount, tx.Nonce, tx.ExtraPower, tx.IsTokenTransfer, tx.TokenID)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}
