package main

import (
	"fmt"
	"main/common"
)

type Operation struct {
	ServerIndex   int
	OperationType string
	Key           string
	Value         string
}

// TODO: decidere quando fare endCall
func testSequential(rpcName string, operations []Operation) {

	endOps := []Operation{
		{0, put, "endKey", "end"},
		{1, put, "endKey", "end"},
		{2, put, "endKey", "end"},
	}

	operations = append(operations, endOps...)

	for _, operation := range operations {
		fmt.Println(operation)
	}

	serverTimestamps := map[int]int{
		0: 0,
		1: 0,
		2: 0,
	}

	responses := make([]common.Response, 3)
	var err error

	for _, operation := range operations {

		responses[operation.ServerIndex], err = executeCall(operation.ServerIndex, rpcName, operation.OperationType,
			operation.Key, operation.Value, serverTimestamps, Async)
		if err != nil {
			fmt.Println("Errore durante l'esecuzione di executeCall")
			return
		}
	}
}

// basicTestSeq contatta tutti i server in goroutine con operazioni di put
// - put x:1 al server1,
// - put x:2 al server2,
// - put x:3 al server3.
func basicTestSeq(rpcName string) {
	fmt.Println("In questo basicTestSeq vengono inviate in goroutine:\n" +
		"- put x:1 al server1\n" +
		"- put x:2 al server2\n" +
		"- put x:3 al server3")

	operations := []Operation{
		{0, put, "x", "1"},
		{1, put, "x", "2"},
		{2, put, "x", "3"},
	}

	testSequential(rpcName, operations)
}

// mediumTestSeq contatta tutti i server in goroutine con operazioni di put
// - put x:1, put y:1, put z:1 al server1,
// - put x:2, put y:2, put z:2 al server2,
// - put x:3, put y:3, put z:3 al server3.
func mediumTestSeq(rpcName string) {
	fmt.Println("In questo mediumTestSeq vengono inviate in sequenza:\n" +
		"- put x:1, put y:1, put z:1 al server1\n" +
		"- put x:2, put y:2, put z:2 al server2\n" +
		"- put x:3, put y:3, put z:3 al server3")

	operations := []Operation{
		{0, put, "x", "1"},
		{1, put, "x", "2"},
		{2, put, "x", "3"},
		{0, put, "y", "1"},
		{1, put, "y", "2"},
		{2, put, "y", "3"},
		{0, put, "z", "1"},
		{1, put, "z", "2"},
		{2, put, "z", "3"},
	}
	testSequential(rpcName, operations)
}

// In questo complexTestSeqNew vengono inviate in goroutine:
//   - put x:1, put y:2, get x, put y:1 al server1,
//   - put x:2, get x, get y, put x:3 al server2,
//   - put x:4, get x, get y al server3.
func complexTestSeq(rpcName string) {
	fmt.Println("In questo complexTestSeq vengono inviate in sequenza:\n" +
		"- put x:1, put y:2, get x, put y:1 al server1\n" +
		"- put x:2, get x, get y, put x:3 al server2\n" +
		"- put x:4, get x, get y al server3")

	operations := []Operation{
		{0, put, "x", "1"},
		{1, put, "x", "2"},
		{2, put, "x", "4"},
		{0, put, "y", "2"},
		{ServerIndex: 1, OperationType: get, Key: "x"},
		{ServerIndex: 2, OperationType: get, Key: "x"},
		{ServerIndex: 0, OperationType: get, Key: "x"},
		{ServerIndex: 1, OperationType: get, Key: "y"},
		{ServerIndex: 2, OperationType: get, Key: "y"},
		{0, put, "y", "1"},
		{1, put, "x", "3"},
	}
	testSequential(rpcName, operations)
}

func endCall(rpcName string, serverTimestamps map[int]int) {
	operations := []Operation{
		{0, put, "endKey", "end"},
		{1, put, "endKey", "end"},
		{2, put, "endKey", "end"},
	}
	testSequential(rpcName, operations)
}
