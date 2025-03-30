package blockchain

import (
	"fmt"
	"strings"
)

const difficulty = 2

// MineBlock performs the Proof of Work to mine the block
func (b *Block) MineBlock() {
	target := strings.Repeat("0", difficulty)
	for !strings.HasPrefix(b.Hash, target) {
		b.Nonce++
		b.Hash = b.CalculateHash()
	}
	fmt.Printf("[POW] Block #%d mined successfully with hash: %s, nonce: %d\n", b.Index, b.Hash, b.Nonce)
}
