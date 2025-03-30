package main

import (
	"encoding/json"
	"fmt"
	"os"
	"unknownberrytrip/internal/api"
	"unknownberrytrip/internal/blockchain"
	"unknownberrytrip/internal/wallet"
)

type WalletData struct {
	Address    string
	PrivateKey string
	PublicKey  string
}

func main() {
	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Create miner wallet
	minerWallet := wallet.NewWallet()
	bc.Balances[minerWallet.Address] = 100.0
	bc.Nonces[minerWallet.Address] = 0

	// Save miner data to temporary file
	walletData := WalletData{
		Address:    minerWallet.Address,
		PrivateKey: fmt.Sprintf("%x", minerWallet.PrivateKey.D), // Private key in hex
		PublicKey:  fmt.Sprintf("%x", minerWallet.PublicKey.X.Bytes()) + fmt.Sprintf("%x", minerWallet.PublicKey.Y.Bytes()),
	}
	data, _ := json.Marshal(walletData)
	os.WriteFile("miner_wallet.json", data, 0644)

	// Start API
	api.StartAPI(bc)

	// Start mining
	bc.StartMining(minerWallet.Address)

	// Output initial state
	fmt.Printf("Blockchain started. Miner address: %s\n", minerWallet.Address)
	fmt.Println("API available at :8080/sendTransaction")

	// Block main thread
	select {}
}
