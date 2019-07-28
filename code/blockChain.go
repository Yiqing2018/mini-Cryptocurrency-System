package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

func generatePuzzle(block Block) string {
	obj := block.PreviousHash + strconv.Itoa(block.Index) + block.Timestamp
	copyBlockTransMap := transMapDeepCopy(block.BlockTransMap)

	for key := range copyBlockTransMap {
		tran := copyBlockTransMap[key]
		tranString := tran.Timestamp + " " + tran.TransactionID + " " + tran.SourceAccNum + " " + tran.DestinationAccNum + " " + strconv.Itoa(tran.TransactionAmount)
		obj = obj + tranString
	}

	h := sha256.New()
	h.Write([]byte(obj))
	hashed := h.Sum(nil)
	returnHash := hex.EncodeToString(hashed)

	return returnHash
}

func requestSolution(puzzleString string) {
	msg := "SOLVE" + " " + puzzleString + "\n"
	_, err := godConn.Write([]byte(msg))
	if err != nil {
		// log.Println("fail to wirte requestSolution to IS ", err)
	}
	// log.Println("Have sent RequestSolve to IS.")
}

func encodeGob(block Block) string {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(block)
	if err != nil {
		fmt.Println("Something wrong to ENcodeGob", err)
	}
	s := base64.StdEncoding.EncodeToString(buf.Bytes())
	return s
}

func decodeGob(serialization string) Block {
	var block Block
	by, err := base64.StdEncoding.DecodeString(serialization)
	if err != nil {
		fmt.Println("Something wrong to read serialization", err)
	}

	buf := bytes.Buffer{}
	buf.Write(by)
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&block)
	if err != nil {
		fmt.Println("Something wrong to DEcodeGob", err)
	}
	return block
}

func dialOnce(msg string, targetAddr string) {
	for {
		dialConn, err := net.Dial("tcp", targetAddr)
		if err != nil {
			// log.Println("dialOnce1 --cannot establish dial connection ", err)
			continue
		}
		bandwidth = bandwidth + len(msg)

		_, err = dialConn.Write([]byte(msg))
		if err != nil {
			dialConn.Close()
			// log.Println("dialOnce2 --cannot write into connection", err)
			continue
		}

		dialConn.Close()
		break
	}
}

func sendBlock(block Block, targetAddr string) {
	msg := "block" + " " + encodeGob(block)
	dialOnce(msg, targetAddr)
}

func receiveBlock(str string) Block {
	return decodeGob(str)
}

func isValid(block Block) {
	checkCode := block.Puzzle
	msg := "VERIFY" + " " + checkCode + " " + block.Solution + "\n"
	// dial to service and request VERIFY
	// log.Println("msg: ", msg)

	for {
		_, err := godConn.Write([]byte(msg))
		//log.Println("successfully write into isValid conn")
		if err != nil {
			continue
		}
		break
	}
}

func balanceMapDeepCopy(map1 map[int]int) map[int]int {
	map2 := make(map[int]int)
	for k, v := range map1 {
		map2[k] = v
	}

	return map2
}

func transMapDeepCopy(map1 map[string]TransRecord) map[string]TransRecord {
	map2 := make(map[string]TransRecord)
	for k, v := range map1 {
		map2[k] = v
	}

	return map2
}

// broadcastBlock
func broadcastBlock(block Block) {
	log.Println("BroadBlockIndex: ", block.Index, " ", block.Puzzle[0:4])
	neighborMap.RLock()
	for key := range neighborMap.data {
		_, _, _, dialAddr := decodeNeighborInfo(key)
		sendBlock(block, dialAddr)
	}
	neighborMap.RUnlock()
}

// generate the Zero Block
func initBlock() Block {
	t := time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
	timeNow := strconv.FormatInt(t, 10)
	initBalanceMap := make(map[int]int)
	initTransMap := make(map[string]TransRecord)
	initBalanceMap[0] = 999999999999
	blockZero := Block{
		PreviousHash: "0",
		Index:        0,
		Timestamp:    timeNow,
		Puzzle:       "0",
		Solution:     "0",
	}

	blockZero.BlockTransMap = transMapDeepCopy(initTransMap)
	blockZero.BalanceMap = balanceMapDeepCopy(initBalanceMap)
	return blockZero
}

// generate the 1st block (before 0 block), and boardcast it.
func blockSection() {
	for true {
		time.Sleep(1 * time.Second)
		if IsNeighborFull {
			currentBlock = initBlock()
			currentPuzzle := calculateCurrentPuzzle()
			requestSolution(currentPuzzle)
			break
		}
	}

}

func calculateCurrentPuzzle() string {
	postTransMap, postBalanceMap := searchTree()
	tempTransactionMap := make(map[string]TransRecord)

	TransMap.RLock()
	for _, item := range TransMap.list {
		id := item.TransactionID
		_, idExisted := postTransMap[id]
		if idExisted {
			continue
		} else {
			from := item.SourceAccNum
			richMan, _ := strconv.Atoi(from)

			_, accountExisted := postBalanceMap[richMan]
			richManBalance := 0
			if accountExisted {
				richManBalance = postBalanceMap[richMan] - item.TransactionAmount
				if richManBalance < 0 {
					continue
				}
			} else {
				continue
				// tempBalanceMap[richMan] = 0
			}

			to := item.DestinationAccNum
			poorMan, _ := strconv.Atoi(to)
			poorManBalance := postBalanceMap[poorMan] + item.TransactionAmount

			tempTransactionMap[id] = item
			postBalanceMap[richMan] = richManBalance
			postBalanceMap[poorMan] = poorManBalance
		}
	}
	TransMap.RUnlock()

	timehey := time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
	timeNow := strconv.FormatInt(timehey, 10)
	var newBlock Block
	newBlock.PreviousHash = currentBlock.Solution
	newBlock.Index = currentBlock.Index + 1
	newBlock.Timestamp = timeNow
	newBlock.BlockTransMap = transMapDeepCopy(tempTransactionMap)
	newBlock.BalanceMap = balanceMapDeepCopy(postBalanceMap)

	result := generatePuzzle(newBlock)
	readyBlock = newBlock

	return result
}
