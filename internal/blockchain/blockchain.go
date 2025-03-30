package blockchain

import (
	"fmt"
	"sync"
	"time"
	"unknownberrytrip/internal/transaction"
)

const reward = 10.0

// Blockchain is a chain of blocks
type Blockchain struct {
	Chain           []*Block
	Balances        map[string]float64
	Nonces          map[string]int
	TransactionPool []transaction.Transaction
	mu              sync.Mutex
}

// NewBlockchain creates a new blockchain with a genesis block
func NewBlockchain() *Blockchain {
	genesisBlock := NewGenesisBlock()
	bc := &Blockchain{
		Chain:           []*Block{genesisBlock},
		Balances:        make(map[string]float64),
		Nonces:          make(map[string]int),
		TransactionPool: []transaction.Transaction{},
	}
	bc.Balances[genesisBlock.Miner] = reward
	return bc
}

// NewGenesisBlock creates a genesis block
func NewGenesisBlock() *Block {
	block := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []transaction.Transaction{},
		PrevHash:     "",
		Nonce:        0,
		Miner:        "genesis_miner",
	}
	block.MineBlock()
	return block
}

// AddTransactionToPool adds a transaction to the pool
func (bc *Blockchain) AddTransactionToPool(tx transaction.Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	fmt.Printf("[Pool] Attempting to add transaction from %s to %s, amount: %f, nonce: %d\n", tx.From, tx.To, tx.Amount, tx.Nonce)
	if !transaction.VerifyTxSignature(&tx) {
		fmt.Println("[Pool] Invalid signature")
		return fmt.Errorf("invalid signature")
	}
	if tx.Nonce != bc.Nonces[tx.From]+1 {
		fmt.Printf("[Pool] Invalid nonce: expected %d, got %d\n", bc.Nonces[tx.From]+1, tx.Nonce)
		return fmt.Errorf("invalid nonce")
	}
	if bc.Balances[tx.From] < tx.Amount {
		fmt.Println("[Pool] Insufficient balance")
		return fmt.Errorf("insufficient balance")
	}

	bc.TransactionPool = append(bc.TransactionPool, tx)
	fmt.Printf("[Pool] Transaction added to pool, pool size: %d\n", len(bc.TransactionPool))
	return nil
}

// AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(transactions []transaction.Transaction, miner string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	fmt.Println("[AddBlock:1] Starting block addition...")
	fmt.Printf("[AddBlock:2] Number of transactions to process: %d\n", len(transactions))
	prevBlock := bc.Chain[len(bc.Chain)-1]
	fmt.Printf("[AddBlock:3] Previous block index: %d, hash: %s\n", prevBlock.Index, prevBlock.Hash)
	newBlock := NewBlock(transactions, prevBlock, miner)
	fmt.Printf("[AddBlock:4] New block created with index: %d, hash: %s\n", newBlock.Index, newBlock.Hash)

	fmt.Println("[AddBlock:5] Starting transaction processing...")
	for i, tx := range transactions {
		fmt.Printf("[AddBlock:6] Processing tx #%d: %s -> %s, amount: %f\n", i, tx.From, tx.To, tx.Amount)
		bc.Balances[tx.From] -= tx.Amount
		fmt.Printf("[AddBlock:7] Deducted %f from %s, new balance: %f\n", tx.Amount, tx.From, bc.Balances[tx.From])
		if _, ok := bc.Balances[tx.To]; !ok {
			bc.Balances[tx.To] = 0
			fmt.Printf("[AddBlock:8] Initialized balance for %s to 0\n", tx.To)
		}
		bc.Balances[tx.To] += tx.Amount
		fmt.Printf("[AddBlock:9] Added %f to %s, new balance: %f\n", tx.Amount, tx.To, bc.Balances[tx.To])
		bc.Nonces[tx.From]++
		fmt.Printf("[AddBlock:10] Updated nonce for %s to %d\n", tx.From, bc.Nonces[tx.From])
	}

	fmt.Println("[AddBlock:11] Processing miner reward...")
	if _, ok := bc.Balances[miner]; !ok {
		bc.Balances[miner] = 0
		fmt.Printf("[AddBlock:12] Initialized balance for miner %s to 0\n", miner)
	}
	bc.Balances[miner] += reward
	fmt.Printf("[AddBlock:13] Added reward %f to miner %s, new balance: %f\n", reward, miner, bc.Balances[miner])

	fmt.Println("[AddBlock:14] Appending block to chain...")
	bc.Chain = append(bc.Chain, newBlock)
	fmt.Printf("[AddBlock:15] Block added, chain length: %d\n", len(bc.Chain))
}

// NewBlock creates a new block
func NewBlock(transactions []transaction.Transaction, prevBlock *Block, miner string) *Block {
	block := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevBlock.Hash,
		Miner:        miner,
		Nonce:        0,
	}
	block.MineBlock()
	return block
}

// IsBlockchainValid checks the integrity of the blockchain
func (bc *Blockchain) IsBlockchainValid() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		prevBlock := bc.Chain[i-1]
		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}
		if currentBlock.Hash != currentBlock.CalculateHash() {
			return false
		}
	}
	return true
}
