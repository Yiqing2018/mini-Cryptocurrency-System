package main

import (
	"net"
	"sync"
)

var treeMap = make(map[int][]*TreeNode)

type TreeNode struct {
	parent *TreeNode
	block  Block
}

type Block struct {
	PreviousHash  string
	Index         int
	Timestamp     string
	BlockTransMap map[string]TransRecord
	BalanceMap    map[int]int
	Puzzle        string
	Solution      string
}

type NeighborInfo struct {
	ipAddr     string
	nodeName   string
	listenPort string
}

var neighborMap = struct {
	sync.RWMutex
	data map[NeighborInfo]bool
}{data: make(map[NeighborInfo]bool)}

// TransRecord : transaction
type TransRecord struct {
	Timestamp         string
	TransactionID     string
	SourceAccNum      string
	DestinationAccNum string
	TransactionAmount int
}

// IntroAddr : the IP address and Port number of Introduction Service
var IntroAddr = "172.22.158.45:4444"

// var TransMap = make(map[string]TransRecord)

var TransMap = struct {
	sync.RWMutex
	data map[string]TransRecord
	list []TransRecord
}{data: make(map[string]TransRecord)}

// IP address, name, port number of local node
var localIPAddr string
var nodeName string
var portNumber string
var numberOfNodes = 100
var IsNeighborFull = false
var godConn net.Conn

var validMap = struct {
	sync.RWMutex
	data map[string]Block
}{data: make(map[string]Block)}

var TransactionInBlock = struct {
	sync.RWMutex
	data map[string]TransRecord
}{data: make(map[string]TransRecord)}

var currentBlock Block
var readyBlock Block
var bandwidth = 0
