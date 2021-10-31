package api

import (
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gitlab.gfsrv.net/research-and-development/go-blockchain/blockchain"
)

func MakeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/addNode", addNewNodeToChain).Methods("GET")
	muxRouter.HandleFunc("/addBlock", addNewBlock).Methods("GET")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	blockchain.PrintChain()
}

func addNewNodeToChain(w http.ResponseWriter, r *http.Request) {
	blockchain.Network.AddNewNode(rand.Int63n(100))
	blockchain.PrintChain()
}

func addNewBlock(w http.ResponseWriter, r *http.Request) {
	_, _, err := blockchain.GenerateNewBlock(blockchain.MasterValidator)

	if err != nil {
		logrus.Error(err)
	}

	winnerNode, _ := blockchain.Network.SelectWinner()
	logrus.Info("Winner:", winnerNode)

	// reward validator
	winnerNode.Stake += 10

	blockchain.PrintChain()
}
