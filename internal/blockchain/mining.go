package blockchain

import (
	"fmt"
	"time"
	"unknownberrytrip/internal/transaction"
)

func (bc *Blockchain) StartMining(minerAddress string) {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			var transactions []transaction.Transaction
			bc.mu.Lock() // Lock for read pool and create block
			if len(bc.TransactionPool) > 0 {
				fmt.Printf("[Mining] Started: %d transactions in pool\n", len(bc.TransactionPool))
				transactions = bc.TransactionPool
				bc.TransactionPool = []transaction.Transaction{}
			} else {
				fmt.Println("[Mining] Tick: No transactions in pool")
				bc.mu.Unlock()
				continue
			}
			prevBlock := bc.Chain[len(bc.Chain)-1]
			fmt.Printf("[Mining] Creating new block #%d...\n", prevBlock.Index+1)
			newBlock := NewBlock(transactions, prevBlock, minerAddress)
			fmt.Printf("[Mining] Block #%d mined successfully with hash: %s, nonce: %d\n", newBlock.Index, newBlock.Hash, newBlock.Nonce)
			bc.mu.Unlock() // Unlock before AddBlock

			fmt.Println("[Mining] Adding block to blockchain...")
			bc.AddBlock(transactions, minerAddress)
			fmt.Printf("[Mining] New block #%d added with %d transactions\n", newBlock.Index, len(transactions))
		}
	}()
}
