package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

// StartNode : send CONNECT command to Introduction service
func StartNode() {
	msg := "CONNECT " + nodeName + " " + localIPAddr + " " + portNumber + "\n"
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Dial2Service(msg)
	}()
	//handle request/response from other nodes
	go Listen2Node()
	// send "requestNeighbor" to random nodes
	go fetchNeighbor()
	go blockSection()
	go pullTransaction()
	wg.Wait()
}

// Dial2Service : Dial to service only once (when first connected)
func Dial2Service(msg string) {
	var wg sync.WaitGroup
	wg.Add(1) // wait for one goroutine, count is 1
	godConn, _ = net.Dial("tcp", IntroAddr)
	defer godConn.Close()

	go func() {
		defer wg.Done()
		Listen2Service()
	}()
	_, err := godConn.Write([]byte(msg))
	if err != nil {
		log.Println("fail to send send msg to IntroductionService")
		return
	}
	//waiting for Listen2Service...
	wg.Wait()
}

func handleListen2Service(recvStr string) {
	head := strings.Split(recvStr, " ")[0]
	switch head {
	case "TRANSACTION":
		splited := strings.Split(recvStr, " ")
		transID := splited[2]
		amount, err := strconv.Atoi(splited[5])
		if err != nil {
			// log.Println("failed to convert transaction amount to integer")
			return
		}

		newTrans := TransRecord{splited[1], splited[2], splited[3], splited[4], amount}
		// log.Println("transaction ", transID)

		TransMap.Lock()
		_, ok := TransMap.data[transID]
		if !ok {
			TransMap.list = append(TransMap.list, newTrans)
		}
		TransMap.data[transID] = newTrans
		TransMap.Unlock()

		go Dial2RandNeighborNewTran(newTrans)

	case "INTRODUCE":
		splited := strings.Split(recvStr, "\n")
		neighborMap.Lock()
		for i := 0; i < len(splited); i++ {
			nb := strings.Split(splited[i], " ")
			key := NeighborInfo{nb[2], nb[1], nb[3]}
			neighborMap.data[key] = false
		}
		neighborMap.Unlock()

	case "SOLVED":
		splited := strings.Split(recvStr, " ")
		Puzzle := splited[1]
		Solution := splited[2]

		readyBlock.Puzzle = Puzzle
		readyBlock.Solution = Solution
		log.Println("CreateInitialBlock ", readyBlock.Puzzle[0:4])
		buildTree(readyBlock)

		TransactionInBlock.Lock()
		trans := readyBlock.BlockTransMap
		for TransactionID, record := range trans {
			_, ok := TransactionInBlock.data[TransactionID]
			if !ok {
				TransactionInBlock.data[TransactionID] = record
				log.Println("TransactionInBlock " + TransactionID)
			}
		}
		TransactionInBlock.Unlock()

		go broadcastBlock(readyBlock)

	case "VERIFY":
		splited := strings.Split(recvStr, " ")
		verifyResult := splited[1]
		verifyPuzzle := splited[2]
		if verifyResult == "OK" {
			var receivedBlock Block
			validMap.Lock()
			for k, v := range validMap.data {
				if k == verifyPuzzle {
					receivedBlock = v
					delete(validMap.data, k)
					break
				}
			}
			validMap.Unlock()
			log.Println("ReceiveVerifiedBlock ", receivedBlock.Puzzle[0:4])
			buildTree(receivedBlock)

			TransactionInBlock.Lock()
			trans := readyBlock.BlockTransMap
			for TransactionID, record := range trans {
				_, ok := TransactionInBlock.data[TransactionID]
				if !ok {
					TransactionInBlock.data[TransactionID] = record
					log.Println("TransactionInBlock " + TransactionID)
				}
			}
			TransactionInBlock.Unlock()

			currentBlock = whichNode().block
			currentPuzzle := calculateCurrentPuzzle()
			requestSolution(currentPuzzle)
		}

	}
}

// Listen2Service : Build the listen connection from service
func Listen2Service() {
	for {
		buf := make([]byte, 1024)
		length, err := godConn.Read(buf)

		if err != nil || godConn == nil {
			log.Println("Gossip lengthOfTransactions ", len(TransMap.data))
			log.Println("BlockChain lengthOfTransactions ", len(TransactionInBlock.data))
			RealEnd := whichNode()
			height := RealEnd.block.Index
			log.Print("height is: ", height)
			if RealEnd.block.BalanceMap != nil {
				log.Println("RealBalance ", RealEnd.block.BalanceMap)
			}
			log.Println("bandwidth ", bandwidth)
			splits := printTree()
			log.Println("splits ", splits)
			return
		}

		recvStr := strings.TrimSuffix(string(buf[0:length]), "\n")
		if recvStr == "DIE" || recvStr == "QUIT" {
			return
		}
		bandwidth = bandwidth + length

		go handleListen2Service(recvStr)
	}
}
