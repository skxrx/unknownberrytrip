package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"unknownberrytrip/internal/transaction"
	"unknownberrytrip/internal/utils"
	"unknownberrytrip/internal/wallet"
)

func main() {
	// Read miner data from file
	data, err := os.ReadFile("miner_wallet.json")
	if err != nil {
		panic("Failed to read miner_wallet.json: " + err.Error())
	}

	var walletData struct {
		Address    string
		PrivateKey string
		PublicKey  string
	}
	if err := json.Unmarshal(data, &walletData); err != nil {
		panic(err)
	}

	// Restore miner's wallet
	senderWallet := &wallet.Wallet{Address: walletData.Address}
	privKeyD, _ := new(big.Int).SetString(walletData.PrivateKey, 16)
	senderWallet.PrivateKey = &ecdsa.PrivateKey{
		D:         privKeyD,
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256()},
	}
	pubKeyBytes, err := hex.DecodeString(walletData.PublicKey)
	if err != nil || len(pubKeyBytes) != 64 { // Expecting 64 bytes (X+Y)
		panic("Invalid public key in miner_wallet.json")
	}
	senderWallet.PublicKey = &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(pubKeyBytes[:32]),
		Y:     new(big.Int).SetBytes(pubKeyBytes[32:]),
	}
	senderWallet.PrivateKey.PublicKey = *senderWallet.PublicKey

	// Create receiver wallet
	receiverWallet := wallet.NewWallet()

	// UNBT transaction
	tx1 := transaction.Transaction{
		From:       senderWallet.Address,
		To:         receiverWallet.Address,
		Amount:     50.0,
		Nonce:      0,
		ExtraPower: 5,
		PubKey:     utils.PubKeyToString(senderWallet.PublicKey), // Using correct format
	}

	// Sign transaction
	txForSign := struct {
		From            string
		To              string
		Amount          float64
		Nonce           int
		ExtraPower      int
		IsTokenTransfer bool
		TokenID         string
	}{tx1.From, tx1.To, tx1.Amount, tx1.Nonce, tx1.ExtraPower, tx1.IsTokenTransfer, tx1.TokenID}
	dataForSign, _ := json.Marshal(txForSign)
	hash := sha256.Sum256(dataForSign)
	r, s, err := ecdsa.Sign(rand.Reader, senderWallet.PrivateKey, hash[:])
	if err != nil {
		panic("Failed to sign transaction: " + err.Error())
	}
	signature := append(r.Bytes(), s.Bytes()...)
	tx1.Signature = hex.EncodeToString(signature)

	// Token transaction
	tx2 := transaction.Transaction{
		From:            senderWallet.Address,
		To:              receiverWallet.Address,
		Amount:          10.0,
		Nonce:           1,
		ExtraPower:      0,
		IsTokenTransfer: true,
		TokenID:         "BERRY_TOKEN",
		PubKey:          utils.PubKeyToString(senderWallet.PublicKey), // Using correct format
	}

	// Sign second transaction
	txForSign2 := struct {
		From            string
		To              string
		Amount          float64
		Nonce           int
		ExtraPower      int
		IsTokenTransfer bool
		TokenID         string
	}{tx2.From, tx2.To, tx2.Amount, tx2.Nonce, tx2.ExtraPower, tx2.IsTokenTransfer, tx2.TokenID}
	dataForSign2, _ := json.Marshal(txForSign2)
	hash2 := sha256.Sum256(dataForSign2)
	r2, s2, err := ecdsa.Sign(rand.Reader, senderWallet.PrivateKey, hash2[:])
	if err != nil {
		panic("Failed to sign transaction: " + err.Error())
	}
	signature2 := append(r2.Bytes(), s2.Bytes()...)
	tx2.Signature = hex.EncodeToString(signature2)

	// Send transactions
	txs := []transaction.Transaction{tx1, tx2}
	for _, tx := range txs {
		txJSON, err := json.Marshal(tx)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post("http://localhost:8080/sendTransaction", "application/json", bytes.NewBuffer(txJSON))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Failed to send transaction: %s\n", resp.Status)
			return
		}
		fmt.Printf("Transaction sent: %+v\n", tx)
	}
}
