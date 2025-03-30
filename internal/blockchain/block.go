package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unknownberrytrip/internal/transaction"
)

// Block represents a single block in the blockchain
type Block struct {
	Index        int
	Timestamp    int64
	Transactions []transaction.Transaction
	PrevHash     string
	Hash         string
	Nonce        int
	Miner        string
}

// CalculateHash computes the hash of the block
func (b *Block) CalculateHash() string {
	record, _ := json.Marshal(struct {
		Index        int
		Timestamp    int64
		Transactions []transaction.Transaction
		PrevHash     string
		Nonce        int
		Miner        string
	}{b.Index, b.Timestamp, b.Transactions, b.PrevHash, b.Nonce, b.Miner})
	h := sha256.New()
	h.Write(record)
	return hex.EncodeToString(h.Sum(nil))
}
