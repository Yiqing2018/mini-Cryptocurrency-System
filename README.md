# mini Cryptocurrency System

## Overview
This is a [course project](https://courses.engr.illinois.edu/ece428/sp2019//mps/mp2.html) of CS425 in UIUC.  

Some basice functions within cryptocurrency system are implemented, including **Gossip Protocol**, transaction **verification**, Nakamoto consensus with **longest chain rule**  

## Running Instruction
step 1. running [Introduction Service](https://gitlab.engr.illinois.edu/nikita/cryptomp-service/blob/master/mp2_service.py)

```
$python3.6 mp2_service.py <port> <tx_rate>
```

* **port** represents the port that Introduction Service listens on, in our program it is coded as "4444", you may want to still use it or modify that
* **tx_rate** represents the rate at which the service will generate transactions.

step 2. initialize new nodes and join the money talk!  

```
$go build ./
$./helper.sh
```  

step 3. Kill half nodes (optional)  

```
$ cd [IntroductionService]
$thanos
```  

step 4. shut down IntroductionService, terminated all nodes, check log files! 

```
$ cd [IntroductionService]
$ QUIT/DIE
```

***

## Design

### Marshalled Message Format
* Trasaction   

messages over the network follow this format:  
```
"TRANSACTION [timestamp] [transactionID] [source account] [destination account] [transaction amount]"
```

* Neighbour information   

message over the network follow this format:  

```
"neighbour [ipAddr] [nodeName] [listenPort]"
```
code structure: 

```
type NeighborInfo struct {
	ipAddr     string
	nodeName   string
	listenPort string
}
```  

*  Block
message over the network follow this format:  

```
"block [Serialized string with Base64]"
```

code structure:

```
type Block struct {
	PreviousHash  string
	Index         int
	Timestamp     string
	BlockTransMap map[string]TransRecord
	BalanceMap    map[int]int
	Puzzle        string
	Solution      string
}
```

### Gossip Protocol - Push+Pull
![](http://ww1.sinaimg.cn/large/006tNc79ly1g5fofs9wffj30k906y0tn.jpg)

* **Transaction Broadcast**  

**push** the newest transaction to its random neighbors  

When received new transactions from Introduction Service, the node will forward this transaction to its neighbors randomly, which can guarantee that the transaction from Introduction Service will not be lost when this node is not the neighbor of any other nodes.  

**pull** history transaction list from its random neighbors  

the node will send "transaction request" to its neighbors randomly in order to "merge" the transaction list with the transactions from other nodes.  

This operation can guarantee that each node will update its transaction list at a quick rate.  

* **Neighbor Update**  

**pull** the neighbor(s) from Introduction Service  

the newly joined node will get the introduced neighbors from IntroductionService after send CONNECT message.

periodically **pull**  neighbor for neighbour’s neighbours  

pull the neighbor list from its random neighbors and merge with original neighbor list, and also the node could take adavantage of  transaction broadcast, we can receive transaction information along with neighbour address! 

* **Node Termination**

Whenever receiving a QUIT/DIE message from IntroductionService, the receiver would quit all processes and die immediately.
If IntroductionService fails, all nodes will continue broadcasting message to other nodes for a limited time(hard coded in our project)

* **Parameter Selection**  

You may **NEED** modify some paramters based on the number of the nodes
For example, when a node tries to pull neighbor list, how many times for sending request to others? 15 rounds for 100 nodes might be ideal.

### Transaction Verification

<p align="center">

  <img width="600" src="http://ww4.sinaimg.cn/large/006tNc79ly1g5fp5x85mwj30hq07b74r.jpg">
  
</p>

* Solve the puzzle from received block
* Generate new block (previous blockID, new data, solution)
* Gossip the new block to other nodes

Note: this part of work is done by Introduction Service, we only need to send **the SHA256 hash of the block** and get new puzzle; request a puzzle solution from the service by issuing a **SOLVE** command
 
### Longest Chain Rule
Using **nTree** structure to preserve longest chain rule.  
Once the node is verified, add the node into tree structure, and find out the longest chain, generate a new block based on the end of longest chain.

<p align="center">

  <img width="600" src="http://ww1.sinaimg.cn/large/006tNc79ly1g5fpdf3shij30iv07mmxu.jpg">
  
</p>


Because we keep parent reference within in a block, it is easy to track all transactions within any chain. Once we know which transactions are maintained in the chain, it is convenient to generate a new block based on new transactions.

*** 

## Experiment
* **Q1.Node Reachability**

<p align="center">

  <img width="600" src="http://ww2.sinaimg.cn/large/006tNc79ly1g5fpk6fh7hj30kt0ek0tw.jpg">
  
</p>

This figure show how many transactions every node eventually received. 
The blue points represent the nodes which were killed on the half way; the red points represent the alive nodes after Thanos.
It is obvious that 100 nodes were reached before Thanos, and 50 nodes were reached after Thanos.

* **Q2.Block Reachability**

<p align="center">

  <img width="300" src="http://ww3.sinaimg.cn/large/006tNc79ly1g5fplamssoj30jn0epq3p.jpg">
    <img width="300" src="http://ww2.sinaimg.cn/large/006tNc79ly1g5fpln9x51j30ke0efq3u.jpg">
  
</p>
Figure1 shows the propagation time of every block. All blocks are propagated within 200ms. Figure2 shows the number of received blocks for every node. Most nodes receive about 90 blocks.

* **Q3.The performance of blockchain**  

According to our experiment, there is only 1 split
![](http://ww3.sinaimg.cn/large/006tNc79ly1g5fpj9al3dj30q708s105.jpg)

How long does each transaction take to appear in a block? 

You **MAY** want to explore more about this implementation... go check the log files OR generate your own logs!

*** 

 
