package main

import (
	"log"
	"strconv"
)

//找最长链的端点 返回回溯的结果-所有transaction和balanceMap
func searchTree() (map[string]TransRecord, map[int]int) {
	if currentBlock.Index == 0 {
		return currentBlock.BlockTransMap, currentBlock.BalanceMap
	}

	longestChainEndNode := whichNode()
	myBalance := longestChainEndNode.block.BalanceMap
	TransRecords := getAllTx(longestChainEndNode)
	return TransRecords, myBalance
}

func getAllTx(endNode *TreeNode) map[string]TransRecord {
	data := make(map[string]TransRecord)
	for endNode != nil {
		TransRecords := endNode.block.BlockTransMap
		//map[string]TransRecord
		for k, v := range TransRecords {
			data[k] = v
		}
		endNode = endNode.parent
	}
	return data
}

// decide which TreeNode we want to backtracking
func whichNode() *TreeNode {
	idx := -1
	//find out the highest height!
	for k := range treeMap {
		if k > idx {
			idx = k
		}
	}
	//if we have multiple paths with same longest height, return the first end
	return treeMap[idx][0]
}

func buildTree(thisBlock Block) {
	index := thisBlock.Index
	//find parent and create TreeNode
	var newNode TreeNode
	if index != 1 {
		MyParent := findParent(thisBlock)
		newNode = TreeNode{
			parent: MyParent,
			block:  thisBlock,
		}
	} else {
		newNode = TreeNode{
			parent: nil,
			block:  thisBlock,
		}
	}
	//insert into treeMap
	if MySlice, ok := treeMap[index]; ok {
		MySlice = append(MySlice, &newNode)
		treeMap[index] = MySlice
	} else {
		var MySlice []*TreeNode
		MySlice = append(MySlice, &newNode)
		treeMap[index] = MySlice
	}
}

func findParent(thisBlock Block) *TreeNode {
	index := thisBlock.Index
	//index>=2
	if MySlice, ok := treeMap[index-1]; ok {
		for i := 0; i < len(MySlice); i++ {
			potentialParent := MySlice[i].block
			if isParent(thisBlock, potentialParent) {
				return MySlice[i]
			}

		}

	}
	//log.Println(thisBlock.Puzzle[0:4], "can't find my dad")
	return nil

}

func isParent(newBlock Block, potentialParent Block) bool {
	if newBlock.PreviousHash == potentialParent.Solution {
		return true
	}
	return false

}

func printTree() int {
	idx := -1
	splits := -1
	for i := range treeMap {
		if i > idx {
			idx = i
		}
	}
	//idx is maximum height
	for i := 1; i <= idx; i++ {
		Myslice := treeMap[i]

		if len(Myslice) > splits {
			splits = len(Myslice)
		}

		str := "level " + strconv.Itoa(i) + " : "
		for i := 0; i < len(Myslice); i++ {
			if i == len(Myslice)-1 {
				str = str + Myslice[i].block.Puzzle[0:4]
				break
			}
			str = str + Myslice[i].block.Puzzle[0:4] + "->"
		}
		log.Print(str)

	}
	return splits
}
