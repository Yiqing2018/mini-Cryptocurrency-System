package main

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func fetchNeighborAll() {

	for {
		time.Sleep(1 * time.Second)
		neighborMap.RLock()
		length := len(neighborMap.data)
		if length == 0 {
			neighborMap.RUnlock()
			continue
		}
		// log.Println("1")
		if length == (numberOfNodes - 1) {
			neighborMap.RUnlock()
			// log.Println("Found all the Neighbors.")
			// Start Building the Block, set the flag to True
			IsNeighborFull = true
			return
		}
		// log.Println("2")
		for k := range neighborMap.data {
			// log.Println("3")

			_, _, _, dialAddr := decodeNeighborInfo(k)
			// log.Println(dialAddr)
			dialConn, err := net.Dial("tcp", dialAddr)
			if err != nil {
				// log.Println("I cannot dial to ", dialAddr)
				continue
			}

			msgRequest := "requestNeighbor" + " " + localIPAddr + " " + portNumber + " " + nodeName
			bandwidth = bandwidth + len(msgRequest)
			_, err = dialConn.Write([]byte(msgRequest))
			if err != nil {
				// log.Println("err in fetchNeighborAll()", err)
			}
			dialConn.Close()
		}
		// log.Println("4")
		neighborMap.RUnlock()
	}

}

func fetchNeighbor() {

	for {
		time.Sleep(1 * time.Second)
		neighborMap.RLock()
		length := len(neighborMap.data)
		bandwidth = bandwidth + length
		if length == 0 {
			neighborMap.RUnlock()
			continue
		}
		if length == (numberOfNodes - 1) {
			neighborMap.RUnlock()
			// log.Println("Found all the Neighbors.")
			// Start Building the Block, set the flag to True
			IsNeighborFull = true

			return
		}

		oneNeighbor := getRandomNeighbor()
		_, _, _, dialAddr := decodeNeighborInfo(oneNeighbor)

		dialConn, err := net.Dial("tcp", dialAddr)
		if err != nil {
			neighborMap.RUnlock()
			continue
		}

		msgRequest := "requestNeighbor" + " " + localIPAddr + " " + portNumber + " " + nodeName
		_, err = dialConn.Write([]byte(msgRequest))
		if err != nil {
			// log.Println(err)
		}
		dialConn.Close()

		neighborMap.RUnlock()
	}

}

// Function: send neighbor List to the specific node
func sendNeighbor(dialAddr string) {
	neighborMap.RLock()
	for key := range neighborMap.data {
		//only send real neighbours!!!
		if (key.ipAddr + ":" + key.listenPort) == dialAddr {
			continue
		}
		//try to dial to dialAddr
		dialConn, err := net.Dial("tcp", dialAddr)
		if err != nil {
			// log.Println("cannot dial to " + dialAddr)
			return
		}
		msg := "neighbor" + " " + key.ipAddr + " " + key.nodeName + " " + key.listenPort
		_, err = dialConn.Write([]byte(msg))
		bandwidth = bandwidth + len(msg)
		if err != nil {
			// log.Println("Failed to write to conn with " + dialAddr)
		}
		dialConn.Close()
	}
	neighborMap.RUnlock()

}

func handleListen2Node(conn net.Conn) {
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	recvStr := buf.String()
	bandwidth = bandwidth + len(recvStr)
	//---------------------------
	// recvStr := string(buf[0:length])
	splited := strings.Split(recvStr, " ")
	tag := splited[0]
	// log.Println("______REVEIVED_______ ", tag)
	switch tag {
	case "requestNeighbor":
		dialAddr := splited[1] + ":" + splited[2]
		newNei := NeighborInfo{splited[1], splited[3], splited[2]}
		//must be one of my NBs, put it into map
		neighborMap.Lock()
		neighborMap.data[newNei] = false
		neighborMap.Unlock()
		go sendNeighbor(dialAddr)

	case "neighbor":
		// log.Print("case neighbour ", recvStr)
		//"neighbor" + " " + key.ipAddr + " " + key.nodeName + " " + key.listenPort
		newNei := NeighborInfo{splited[1], splited[2], splited[3]}
		//must be one of my NBs, because already checked when sending the msg
		neighborMap.Lock()
		neighborMap.data[newNei] = false
		neighborMap.Unlock()
	case "requestInitTran":
		dialAddr := splited[1] + ":" + splited[2]
		if !IsNeighborFull {
			newNei := NeighborInfo{splited[1], splited[3], splited[2]}
			neighborMap.Lock()
			neighborMap.data[newNei] = false
			neighborMap.Unlock()
		}
		go sendInitTran(dialAddr)
	case "initTran":
		amount, err := strconv.Atoi(splited[5])
		if err != nil {
			// log.Println("failed to convert transaction amount to integer")
			return
		}
		TransMap.Lock()
		newTran := TransRecord{splited[1], splited[2], splited[3], splited[4], amount}
		_, ok := TransMap.data[newTran.TransactionID]
		if !ok {
			TransMap.list = append(TransMap.list, newTran)
			// log.Println("transaction " + newTran.TransactionID)
		}

		TransMap.data[newTran.TransactionID] = newTran

		TransMap.Unlock()
	case "block":
		received := receiveBlock(splited[1])
		//log.Println("received block index (before verify): ", received.Index)
		validMap.Lock()
		validMap.data[received.Puzzle] = received
		validMap.Unlock()
		isValid(received)
	}

	conn.Close()
}

//Listen2NodeforNB : listen to other nodes
func Listen2Node() {
	listenPort := ":" + portNumber
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listenPort)
	if err != nil {
		log.Println(err)
		return
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			// conn.Close()
			continue
		}

		go handleListen2Node(conn)
	}
}

// Helper function: decode NeighborInfo
func decodeNeighborInfo(key NeighborInfo) (string, string, string, string) {
	ipAddr := key.ipAddr
	symbolicName := key.nodeName
	listenPort := key.listenPort
	dialAddr := ipAddr + ":" + listenPort

	return ipAddr, symbolicName, listenPort, dialAddr
}

// Dial2RandNeighborNewTran : dial to random neighbors, and then send newTran
func Dial2RandNeighborNewTran(tran TransRecord) {

	neighborMap.RLock()
	defer neighborMap.RUnlock()
	pushFanout := len(neighborMap.data) / 3
	if pushFanout == 0 {
		return
	}

	for i := 0; i < 100; i++ {
		time.Sleep(time.Duration(1) * time.Second)
		recordNeighbor := make(map[NeighborInfo]bool)
		var randomAmount int
		if len(neighborMap.data) < pushFanout {
			randomAmount = len(neighborMap.data)
		} else {
			randomAmount = pushFanout
		}

		for {
			if len(recordNeighbor) == randomAmount {
				break
			}

			oneNeighbor := getRandomNeighbor()
			_, _, _, dialAddr := decodeNeighborInfo(oneNeighbor)

			dialConn, err := net.Dial("tcp", dialAddr)
			if err != nil {
				// delete(neighborMap, oneNeighbor)
				continue
			}

			msg := "initTran" + " " + tran.Timestamp + " " + tran.TransactionID + " " + tran.SourceAccNum + " " + tran.DestinationAccNum + " " + strconv.Itoa(tran.TransactionAmount)
			_, err = dialConn.Write([]byte(msg))
			bandwidth = bandwidth + len(msg)
			if err != nil {
				//log.Println("(Dial2RandNeighborNewTran) Failed to write to" + dialAddr + "--SAD")
				continue
			}
			recordNeighbor[oneNeighbor] = false
			dialConn.Close()

		}
	}
}

// Function: send Tran List to the specific node
func sendInitTran(dialAddr string) {
	TransMap.RLock()
	for key := range TransMap.data {

		tran := TransMap.data[key]
		msg := "initTran" + " " + tran.Timestamp + " " + tran.TransactionID + " " + tran.SourceAccNum + " " + tran.DestinationAccNum + " " + strconv.Itoa(tran.TransactionAmount)

		dialConn, err := net.Dial("tcp", dialAddr)
		bandwidth = bandwidth + len(msg)
		if err != nil {
			// log.Println("(sendInitTran) Failed to dial to" + dialAddr)
			break
		}
		_, err = dialConn.Write([]byte(msg))
		if err != nil {
			// log.Println("(sendInitTran) Failed to write to" + dialAddr)
		}
		dialConn.Close()

	}
	TransMap.RUnlock()

}

//unlimited loop
func pullTransaction() {
	for {

		time.Sleep(1 * time.Second)

		neighborMap.RLock()

		recordNeighbor := make(map[NeighborInfo]bool)
		var randomAmount int
		if len(neighborMap.data) == 0 {
			neighborMap.RUnlock()
			continue
		} else if len(neighborMap.data) < 4 {
			randomAmount = len(neighborMap.data)
		} else {
			randomAmount = 4
		}

		for {
			if len(recordNeighbor) == randomAmount {
				break
			}
			oneNeighbor := getRandomNeighbor()
			_, ok := recordNeighbor[oneNeighbor]
			if ok {
				continue
			}
			_, _, _, dialAddr := decodeNeighborInfo(oneNeighbor)

			dialConn, err := net.Dial("tcp", dialAddr)
			if err != nil {
				continue
			}

			msgRequestTranHist := "requestInitTran" + " " + localIPAddr + " " + portNumber + " " + nodeName
			_, err = dialConn.Write([]byte(msgRequestTranHist))
			bandwidth = bandwidth + len(msgRequestTranHist)
			if err != nil {
				continue
			}
			dialConn.Close()
			recordNeighbor[oneNeighbor] = false
		}

		neighborMap.RUnlock()
	}

}

// getIPaddr : get my own IP address
func getIPaddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		// log.Println("Failed to get local IP address", err)
	}
	var myIPaddress string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myIPaddress = ipnet.IP.String()
			}
		}
	}
	return myIPaddress
}

// Helper function: generate a random number
func generateRandomNumber(end int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(end)
	return num
}

// Helper function: generate a random valid neighbor, if nonexists return null
func getRandomNeighbor() NeighborInfo {
	nums := generateRandomNumber(len(neighborMap.data))
	indexMap := 0
	var result NeighborInfo
	for key := range neighborMap.data {
		if indexMap == nums {
			return key
		}
		indexMap = indexMap + 1
	}
	return result
}
