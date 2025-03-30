package blockchain

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
	"unknownberrytrip/internal/transaction"
)

const reward = 10.0          // Base block reward in UNBT
const baseFee = 0.001        // Minimum fee in UNBT for regular transactions
const powerPerUNBT = 1000    // 0.001 UNBT = 1 Power
const baseTokenPower = 10    // Base cost for token transfer in Power
const dailyBP = 100          // 100 BP per day per address
const baseUNBTpower = 1      // 1 BP for regular UNBT transaction
const extraPowerCost = 0.001 // 0.001 UNBT per 1 Extra Power

// Blockchain is a chain of blocks
type Blockchain struct {
	Chain           []*Block
	Balances        map[string]float64        // Balance in UNBT
	Nonces          map[string]int            // Nonce for transaction ordering
	BasePower       map[string]int            // BasePower for addresses
	LastBPUpdate    map[string]int64          // Time of last BP update
	TransactionPool []transaction.Transaction // Transaction pool
	Miners          map[string]int            // Miner transaction counter in current block
	mu              sync.Mutex
}

// NewBlockchain creates a new blockchain with a genesis block
func NewBlockchain() *Blockchain {
	genesisBlock := NewGenesisBlock()
	bc := &Blockchain{
		Chain:           []*Block{genesisBlock},
		Balances:        make(map[string]float64),
		Nonces:          make(map[string]int),
		BasePower:       make(map[string]int),
		LastBPUpdate:    make(map[string]int64),
		TransactionPool: []transaction.Transaction{},
		Miners:          make(map[string]int),
	}
	bc.Balances[genesisBlock.Miner] = reward
	bc.Nonces[genesisBlock.Miner] = 0
	bc.BasePower[genesisBlock.Miner] = dailyBP
	bc.LastBPUpdate[genesisBlock.Miner] = time.Now().Unix()
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

// updateBasePower updates BP for address
func (bc *Blockchain) updateBasePower(address string) {
	now := time.Now().Unix()
	lastUpdate, exists := bc.LastBPUpdate[address]
	if !exists {
		bc.BasePower[address] = dailyBP
		bc.LastBPUpdate[address] = now
		return
	}
	daysPassed := float64(now-lastUpdate) / (24 * 3600)
	if daysPassed >= 1 {
		bc.BasePower[address] = int(math.Min(float64(bc.BasePower[address])+dailyBP*daysPassed, dailyBP))
		bc.LastBPUpdate[address] = now
	}
}

// calculatePendingBalance calculates available balance considering transactions in the pool
func (bc *Blockchain) calculatePendingBalance(address string) float64 {
	pendingCost := 0.0
	for _, tx := range bc.TransactionPool {
		if tx.From == address {
			pendingCost += tx.Amount
			if tx.IsTokenTransfer && bc.BasePower[address] < baseTokenPower {
				pendingCost += float64(baseTokenPower) / powerPerUNBT
			} else if !tx.IsTokenTransfer && bc.BasePower[address] < baseUNBTpower {
				pendingCost += baseFee
			}
			if tx.ExtraPower > 0 {
				pendingCost += float64(tx.ExtraPower) * extraPowerCost
			}
		}
	}
	return bc.Balances[address] - pendingCost
}

// AddTransactionToPool adds a transaction to the pool
func (bc *Blockchain) AddTransactionToPool(tx transaction.Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	fmt.Printf("[Pool] Attempting to add transaction from %s to %s, amount: %f, nonce: %d, extraPower: %d, isToken: %t\n", tx.From, tx.To, tx.Amount, tx.Nonce, tx.ExtraPower, tx.IsTokenTransfer)
	if !transaction.VerifyTxSignature(&tx) {
		fmt.Println("[Pool] Invalid signature")
		return fmt.Errorf("invalid signature")
	}

	// Get current nonce, if address doesn't exist — consider it 0
	currentNonce, exists := bc.Nonces[tx.From]
	if !exists {
		currentNonce = 0
	}
	expectedNonce := currentNonce // Expect current nonce, not +1
	if tx.Nonce != expectedNonce {
		fmt.Printf("[Pool] Invalid nonce: expected %d, got %d\n", expectedNonce, tx.Nonce)
		return fmt.Errorf("invalid nonce")
	}

	bc.updateBasePower(tx.From)
	requiredPower := baseUNBTpower // 1 BP for UNBT
	if tx.IsTokenTransfer {
		requiredPower = baseTokenPower // 10 BP for tokens
	}

	totalCost := 0.0
	if bc.BasePower[tx.From] >= requiredPower {
		bc.BasePower[tx.From] -= requiredPower
		fmt.Printf("[Pool] Used %d BasePower, remaining: %d\n", requiredPower, bc.BasePower[tx.From])
	} else {
		if tx.IsTokenTransfer {
			totalCost = float64(requiredPower) / powerPerUNBT // 0.01 UNBT for tokens
		} else {
			totalCost = baseFee // 0.001 UNBT for UNBT
		}
		fmt.Printf("[Pool] Insufficient BasePower, using %f UNBT instead\n", totalCost)
	}

	// Consider Extra Power
	extraPowerCostTotal := float64(tx.ExtraPower) * extraPowerCost
	if extraPowerCostTotal > 0 {
		totalCost += extraPowerCostTotal
		fmt.Printf("[Pool] Added %f UNBT for %d ExtraPower\n", extraPowerCostTotal, tx.ExtraPower)
	}

	pendingBalance := bc.calculatePendingBalance(tx.From)
	if pendingBalance < tx.Amount+totalCost {
		fmt.Printf("[Pool] Insufficient balance: need %f UNBT (amount + fee), have %f after pending\n", tx.Amount+totalCost, pendingBalance)
		return fmt.Errorf("insufficient balance")
	}

	bc.TransactionPool = append(bc.TransactionPool, tx)
	fmt.Printf("[Pool] Transaction added to pool, pool size: %d\n", len(bc.TransactionPool))

	// Increment nonce only after successful addition
	bc.Nonces[tx.From] = expectedNonce + 1
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

	bc.Miners = make(map[string]int)
	for i, tx := range transactions {
		fmt.Printf("[AddBlock:5] Processing tx #%d: %s -> %s, amount: %f, isToken: %t\n", i, tx.From, tx.To, tx.Amount, tx.IsTokenTransfer)
		bc.updateBasePower(tx.From)
		totalCost := 0.0
		// BP already written in AddTransactionToPool, here only UNBT
		if tx.IsTokenTransfer && bc.BasePower[tx.From] < baseTokenPower {
			totalCost = float64(baseTokenPower) / powerPerUNBT
		} else if !tx.IsTokenTransfer && bc.BasePower[tx.From] < baseUNBTpower {
			totalCost = baseFee
		}
		if tx.ExtraPower > 0 {
			totalCost += float64(tx.ExtraPower) * extraPowerCost
		}
		if totalCost > 0 {
			if bc.Balances[tx.From] < totalCost {
				fmt.Printf("[AddBlock:6] Insufficient balance for fee %f UNBT from %s\n", totalCost, tx.From)
				continue
			}
			bc.Balances[tx.From] -= totalCost
			fmt.Printf("[AddBlock:6] Deducted fee %f UNBT from %s\n", totalCost, tx.From)
		}

		if bc.Balances[tx.From] < tx.Amount {
			fmt.Printf("[AddBlock:7] Insufficient balance for amount %f UNBT from %s\n", tx.Amount, tx.From)
			continue
		}
		bc.Balances[tx.From] -= tx.Amount
		fmt.Printf("[AddBlock:7] Deducted amount %f UNBT from %s, new balance: %f\n", tx.Amount, tx.From, bc.Balances[tx.From])
		if _, ok := bc.Balances[tx.To]; !ok {
			bc.Balances[tx.To] = 0
			fmt.Printf("[AddBlock:8] Initialized balance for %s to 0\n", tx.To)
		}
		bc.Balances[tx.To] += tx.Amount
		fmt.Printf("[AddBlock:9] Added %f UNBT to %s, new balance: %f\n", tx.Amount, tx.To, bc.Balances[tx.To])
		bc.Nonces[tx.From]++
		fmt.Printf("[AddBlock:10] Incremented nonce for sender %s to %d\n", tx.From, bc.Nonces[tx.From])
		bc.Miners[miner]++
	}

	// Reward distribution
	totalTx := len(transactions)
	if totalTx > 0 {
		type minerStat struct {
			Miner   string
			TxCount int
		}
		var miners []minerStat
		for m, count := range bc.Miners {
			miners = append(miners, minerStat{Miner: m, TxCount: count})
		}
		sort.Slice(miners, func(i, j int) bool {
			return miners[i].TxCount > miners[j].TxCount
		})

		remainingReward := reward
		if _, ok := bc.Balances[miners[0].Miner]; !ok {
			bc.Balances[miners[0].Miner] = 0
		}
		bc.Balances[miners[0].Miner] += 5.0
		remainingReward -= 5.0
		fmt.Printf("[AddBlock:11] Top miner %s processed %d tx, awarded 5 UNBT, new balance: %f\n", miners[0].Miner, miners[0].TxCount, bc.Balances[miners[0].Miner])

		if len(miners) > 1 {
			remainingTx := totalTx - miners[0].TxCount
			if remainingTx > 0 {
				for i := 1; i < len(miners); i++ {
					if _, ok := bc.Balances[miners[i].Miner]; !ok {
						bc.Balances[miners[i].Miner] = 0
					}
					share := (float64(miners[i].TxCount) / float64(remainingTx)) * remainingReward
					bc.Balances[miners[i].Miner] += share
					fmt.Printf("[AddBlock:12] Miner %s processed %d tx, awarded %f UNBT, new balance: %f\n", miners[i].Miner, miners[i].TxCount, share, bc.Balances[miners[i].Miner])
				}
			}
		}
	}

	fmt.Println("[AddBlock:13] Appending block to chain...")
	bc.Chain = append(bc.Chain, newBlock)
	fmt.Printf("[AddBlock:14] Block added, chain length: %d\n", len(bc.Chain))
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
