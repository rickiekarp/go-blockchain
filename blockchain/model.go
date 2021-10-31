package blockchain

import (
	"time"
)

type PoSNetwork struct {
	Blockchain     []*Block
	BlockchainHead *Block
	Validators     []*Node
}

type Block struct {
	Id            int64
	Timestamp     time.Time
	transactions  []string
	PrevHash      []byte
	Hash          []byte
	ValidatorAddr string
}

type Node struct {
	Stake   int64
	Address string
}
