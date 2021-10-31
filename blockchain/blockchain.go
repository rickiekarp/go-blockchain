package blockchain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.gfsrv.net/research-and-development/go-blockchain/node"
	"gitlab.gfsrv.net/research-and-development/go-blockchain/util"
)

var Network PoSNetwork
var MasterValidator *Node

func InitNetwork() {
	logrus.Info("Initializing network")
	Network = PoSNetwork{
		nil,
		nil,
		nil,
	}

	if util.Exists(ChainConfig.Storage.Path + "/node") {
		content, err := ioutil.ReadFile(ChainConfig.Storage.Path + "/node")
		if err != nil {
			log.Fatal(err)
		}
		err = Network.AddNodeByAddress(string(content))
		if err != nil {
			logrus.Error("could not add node by address")
			os.Exit(1)
		}
		MasterValidator = Network.Validators[0]
	} else {

		// init master validator
		Network.AddNewNode(100)
		MasterValidator = Network.Validators[0]
		err := util.WriteToFile([]byte(MasterValidator.Address), ChainConfig.Storage.Path+"/node")
		if err != nil {
			logrus.Error(err)
		}
		util.WriteStructToFile(MasterValidator, ChainConfig.Storage.Path+"/nodes/"+MasterValidator.Address)
	}

	logrus.Info("Initializing chain")

	// generate an initial PoS network including a blockchain with a genesis block.
	genesisTime := time.Now()
	genesisData := []string{"genesis block"}
	genesisHash := []byte("")

	blockChain := []*Block{
		{
			Id:            1,
			Timestamp:     genesisTime,
			PrevHash:      genesisHash,
			transactions:  genesisData,
			Hash:          NewHash(genesisTime, genesisData, genesisHash),
			ValidatorAddr: Network.Validators[0].Address,
		},
	}

	switch ChainConfig.Storage.Type {
	case "offchain":
		err := WriteBlockToLocalFile(blockChain[0])
		if err != nil {
			logrus.Error(err)
		}
	}

	Network.Blockchain = blockChain
	Network.BlockchainHead = blockChain[0]

	PrintChain()

	// newBlock := NewBlock([]string{"first block"}, chain.blocks[len(chain.blocks)-1])
	// chain.AddBlock(*newBlock)

	// secondBlock := NewBlock([]string{"second block"}, chain.blocks[len(chain.blocks)-1])
	// chain.AddBlock(*secondBlock)
}

func GetChain() []*Block {
	return Network.Blockchain
}

func NewBlock(transactions []string, prevBlock *Block, validator *Node) *Block {
	currentTime := time.Now().UTC()
	return &Block{
		Id:            prevBlock.Id + 1,
		Timestamp:     currentTime,
		transactions:  transactions,
		PrevHash:      prevBlock.Hash,
		Hash:          NewHash(currentTime, transactions, prevBlock.Hash),
		ValidatorAddr: validator.Address,
	}
}

func NewHash(time time.Time, transactions []string, prevHash []byte) []byte {
	input := time.String() + string(prevHash)
	for transaction := range transactions {
		input += string(rune(transaction))
	}
	hash := sha256.Sum256([]byte(input))
	return hash[:]
}

func (n *PoSNetwork) AddNewNode(stake int64) {
	newNode := &Node{
		Address: node.RandStringRunes(32),
		Stake:   stake,
	}

	switch ChainConfig.Storage.Type {
	case "offchain":
		util.WriteStructToFile(newNode, ChainConfig.Storage.Path+"/nodes/"+newNode.Address)
	}
	n.Validators = append(n.Validators, newNode)
}

func (n *PoSNetwork) AddNodeByAddress(address string) error {
	content, err := ioutil.ReadFile(ChainConfig.Storage.Path + "/node")
	if err != nil {
		return err
	}

	node, err := ReadNodeByAddress(string(content))
	if err != nil {
		return err
	}
	n.Validators = append(n.Validators, node)
	return nil
}

func GenerateNewBlock(Validator *Node) ([]*Block, *Block, error) {
	if err := Network.ValidateBlockchain(); err != nil {
		logrus.Error(err)
		Validator.Stake -= ChainConfig.Defaults.StakeReward
		return Network.Blockchain, Network.BlockchainHead, err
	}

	newBlock := NewBlock([]string{"data"}, Network.BlockchainHead, Validator)

	if err := ValidateBlockCandidate(newBlock); err != nil {
		logrus.Error(err)
		Validator.Stake -= ChainConfig.Defaults.StakeReward
		return Network.Blockchain, Network.BlockchainHead, err
	} else {
		logrus.Info("Add new block to chain with id: ", newBlock.Id)
		Network.Blockchain = append(Network.Blockchain, newBlock)
		Network.BlockchainHead = newBlock
	}
	return Network.Blockchain, newBlock, nil
}

func (n PoSNetwork) ValidateBlockchain() error {
	if len(n.Blockchain) <= 1 {
		return nil
	}

	currBlockIdx := len(n.Blockchain) - 1
	prevBlockIdx := len(n.Blockchain) - 2

	for prevBlockIdx >= 0 {
		currBlock := n.Blockchain[currBlockIdx]
		prevBlock := n.Blockchain[prevBlockIdx]
		if !bytes.Equal(currBlock.PrevHash, prevBlock.Hash) {
			return errors.New("blockchain has inconsistent hashes")
		}

		if n.BlockchainHead.Timestamp.Before(prevBlock.Timestamp) {
			return errors.New("blockchain has inconsistent timestamps")
		}

		// if NewBlockHash(prevBlock) != currBlock.Hash {
		// 	return errors.New("blockchain has inconsistent hash generation")
		// }
		currBlockIdx--
		prevBlockIdx--
	}
	return nil
}

func ValidateBlockCandidate(newBlock *Block) error {
	if !bytes.Equal(Network.BlockchainHead.Hash, newBlock.PrevHash) {
		return errors.New("blockchain HEAD hash is not equal to new block previous hash")
	}

	if Network.BlockchainHead.Timestamp.After(newBlock.Timestamp) {
		return errors.New("blockchain HEAD timestamp is greater than or equal to new block timestamp")
	}

	// if NewBlockHash(n.BlockchainHead) != newBlock.Hash {
	// 	return errors.New("new block hash of blockchain HEAD does not equal new block hash")
	// }
	return nil
}

func (n PoSNetwork) SelectWinner() (*Node, error) {
	var winnerPool []*Node
	var totalStake int64 = 0
	for _, node := range n.Validators {
		if node.Stake > 0 {
			winnerPool = append(winnerPool, node)
			totalStake += node.Stake
		}
	}
	if winnerPool == nil {
		return nil, errors.New("there are no nodes with stake in the network")
	}
	winnerNumber := rand.Int63n(totalStake)
	var tmp int64 = 0
	for _, node := range n.Validators {
		tmp += node.Stake
		if winnerNumber < tmp {
			return node, nil
		}
	}
	return nil, errors.New("a winner should have been picked but wasn't")
}

func ValidateStorage() {
	if !util.Exists(ChainConfig.Storage.Path) {
		logrus.Info("Creating storage location: " + ChainConfig.Storage.Path)
		util.CreateFolder(ChainConfig.Storage.Path)
	}
}

func WriteBlockToLocalFile(block *Block) error {
	blockPath := ChainConfig.Storage.Path + "/chain/blocks/"

	if !util.Exists(blockPath) {
		if err := os.MkdirAll(filepath.Dir(blockPath), 0770); err != nil {
			logrus.Println("Could not create path!")
		}
	}

	util.WriteStructToFile(block, blockPath+fmt.Sprint(block.Id))
	return nil
}

func PrintChain() {
	printValidators()
	logrus.Info("Blocks:")
	for _, block := range Network.Blockchain {
		printBlockInformation(block)
		logrus.Info("")
	}
}

func printBlockInformation(block *Block) {
	logrus.Printf("BlockID: %d\n", block.Id)
	logrus.Printf("Previous hash: %x\n", block.PrevHash)
	logrus.Printf("Timestamp: %s\n", block.Timestamp)
	logrus.Printf("Hash: %x\n", block.Hash)
	logrus.Printf("ValidatorAddress: %s\n", block.ValidatorAddr)
	printTransactions(block)
}

func printTransactions(block *Block) {
	for _, transaction := range block.transactions {
		logrus.Printf("Data: %s\n", transaction)
	}
}

func printValidators() {
	logrus.Info("Validators:")
	for _, validator := range Network.Validators {
		logrus.Printf("Address: %s, Stake: %d\n", validator.Address, validator.Stake)
	}
}
