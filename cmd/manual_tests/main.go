package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
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

	// Restore miner wallet
	senderWallet := &wallet.Wallet{Address: walletData.Address}
	privKeyD, _ := new(big.Int).SetString(walletData.PrivateKey, 16)
	senderWallet.PrivateKey = &ecdsa.PrivateKey{
		D:         privKeyD,
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256()},
	}
	pubKeyBytes, _ := hex.DecodeString(walletData.PublicKey)
	senderWallet.PublicKey = &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(pubKeyBytes[:len(pubKeyBytes)/2]),
		Y:     new(big.Int).SetBytes(pubKeyBytes[len(pubKeyBytes)/2:]),
	}
	senderWallet.PrivateKey.PublicKey = *senderWallet.PublicKey

	// Create receiver wallet
	receiverWallet := wallet.NewWallet()

	// Transaction parameters
	amount := 50.0
	nonce := 0 // Start with 0 since this is the first transaction

	// Create and sign transaction
	tx := senderWallet.CreateTransaction(receiverWallet.Address, amount, nonce)

	// Output debug data
	fmt.Printf("Sender Address (Miner): %s\n", senderWallet.Address)
	fmt.Printf("Receiver Address: %s\n", receiverWallet.Address)
	fmt.Printf("Transaction: %+v\n", tx)

	// Prepare JSON for sending
	txJSON, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}

	// Send request to API
	resp, err := http.Post("http://localhost:8080/sendTransaction", "application/json", bytes.NewBuffer(txJSON))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to send transaction: %s\n", resp.Status)
		return
	}

	fmt.Println("Transaction sent successfully")
}
