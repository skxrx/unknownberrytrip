package blockchain

import (
	"testing"
	"unknownberrytrip/internal/transaction"
	"unknownberrytrip/internal/wallet"
)

func TestAddBlock(t *testing.T) {
	t.Log("Creating a new blockchain instance")
	bc := NewBlockchain()

	t.Log("Creating miner and user wallets")
	minerWallet := wallet.NewWallet()
	userWallet := wallet.NewWallet()

	t.Logf("Miner address: %s", minerWallet.Address)
	t.Logf("User address: %s", userWallet.Address)

	// Set initial balance
	t.Logf("Setting initial miner balance to 100.0")
	bc.Balances[minerWallet.Address] = 100.0
	bc.Nonces[minerWallet.Address] = 0

	// Create transaction
	t.Logf("Creating transaction: miner -> user, amount: 50.0, nonce: %d", bc.Nonces[minerWallet.Address])
	tx := minerWallet.CreateTransaction(userWallet.Address, 50.0, bc.Nonces[minerWallet.Address])
	t.Logf("Transaction created with signature: %s", tx.Signature)

	// Add block
	t.Log("Adding block with the transaction to the blockchain")
	bc.AddBlock([]transaction.Transaction{*tx}, minerWallet.Address)
	t.Logf("Current blockchain length: %d blocks", len(bc.Chain))

	// Check balances
	expectedMinerBalance := 100.0 - 50.0 + reward // 60.0
	t.Logf("Expected miner balance: %f (100.0 - 50.0 + %f)", expectedMinerBalance, reward)
	t.Logf("Actual miner balance: %f", bc.Balances[minerWallet.Address])

	if bc.Balances[minerWallet.Address] != expectedMinerBalance {
		t.Errorf("Expected miner balance %f, got %f", expectedMinerBalance, bc.Balances[minerWallet.Address])
	}

	t.Logf("Expected user balance: 50.0")
	t.Logf("Actual user balance: %f", bc.Balances[userWallet.Address])

	if bc.Balances[userWallet.Address] != 50.0 {
		t.Errorf("Expected user balance 50.0, got %f", bc.Balances[userWallet.Address])
	}

	t.Log("TestAddBlock completed successfully")
}

func TestAddValidSignedTransactionToPool(t *testing.T) {
	// Initialize blockchain
	t.Log("Initializing a new blockchain")
	bc := NewBlockchain()

	// Create sender wallet
	t.Log("Creating sender wallet")
	senderWallet := wallet.NewWallet()
	t.Logf("Sender address: %s", senderWallet.Address)

	// Give sender initial balance and set nonce
	t.Log("Setting initial sender balance to 100.0")
	bc.Balances[senderWallet.Address] = 100.0
	bc.Nonces[senderWallet.Address] = 0

	// Create receiver
	t.Log("Creating receiver wallet")
	receiverWallet := wallet.NewWallet()
	t.Logf("Receiver address: %s", receiverWallet.Address)

	// Create and sign transaction
	t.Logf("Creating transaction: sender -> receiver, amount: 50.0, nonce: %d", bc.Nonces[senderWallet.Address])
	tx := senderWallet.CreateTransaction(receiverWallet.Address, 50.0, bc.Nonces[senderWallet.Address])
	t.Logf("Transaction created with signature: %s", tx.Signature)

	// Verify that signature is valid
	t.Log("Verifying transaction signature")
	isValid := transaction.VerifyTxSignature(tx)
	t.Logf("Signature verification result: %v", isValid)

	if !isValid {
		t.Fatal("Transaction should have a valid signature")
	}

	// Add transaction to pool
	t.Log("Adding transaction to the pool")
	err := bc.AddTransactionToPool(*tx)
	t.Logf("AddTransactionToPool result: %v", err == nil)

	if err != nil {
		t.Fatalf("Expected successful transaction addition to pool, got error: %v", err)
	}

	// Verify that transaction appears in the pool
	bc.mu.Lock()
	t.Logf("Transaction pool size: %d", len(bc.TransactionPool))
	if len(bc.TransactionPool) > 0 {
		poolTx := bc.TransactionPool[0]
		t.Logf("Pool transaction details - From: %s, To: %s, Amount: %f",
			poolTx.From, poolTx.To, poolTx.Amount)
	}
	bc.mu.Unlock()

	bc.mu.Lock()
	defer bc.mu.Unlock()
	if len(bc.TransactionPool) != 1 {
		t.Errorf("Expected 1 transaction in pool, got %d", len(bc.TransactionPool))
	}
	if bc.TransactionPool[0].From != senderWallet.Address || bc.TransactionPool[0].To != receiverWallet.Address || bc.TransactionPool[0].Amount != 50.0 {
		t.Errorf("Transaction in pool does not match expected: %+v", bc.TransactionPool[0])
	}

	t.Log("TestAddValidSignedTransactionToPool completed successfully")
}
