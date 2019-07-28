package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	nodeName = os.Args[1]
	portNumber = os.Args[2]
	localIPAddr = getIPaddr()
	file, err := os.Create(nodeName + ".log")
	defer file.Close()
	if err != nil {
		fmt.Println("fail to create log file")
		return
	}
	log.SetOutput(file)
	defer file.Close()
	StartNode()
}
