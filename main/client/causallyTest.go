package main

import (
	"fmt"
	"main/common"
)

func testCausal(rpcName string, operations [][]Operation) {

	// serverTimestamps è una mappa che associa a ogni server un timestamp
	serverTimestamps := make(map[int]int)
	for i := 0; i < common.Replicas; i++ {
		serverTimestamps[i] = 0
	}

	responses := make([]common.Response, common.Replicas)
	var err error

	// Assume operations are sorted in the order they should be executed
	for _, operation := range operations {

		go func(operation []Operation) {
			for _, op := range operation {
				args := common.Args{Key: op.Key, Value: op.Value, Timestamp: serverTimestamps[op.ServerIndex]}
				responses[op.ServerIndex], err = executeCall(op.ServerIndex, rpcName+op.OperationType, args, sync, specific)
				serverTimestamps[op.ServerIndex]++ // In caso sincrono? qui?

				if err != nil {
					fmt.Println("Errore durante l'esecuzione di executeCall")
					return
				}
			}
		}(operation)
	}

	// ----- Da qui inseriamo le operazioni di fine ??? ----- //
}

// In questo basicTestCE vengono inviate in goroutine:
//   - una richiesta di put x:a e put y:b al server1,
//   - una richiesta di get x e put x:b al server2,
//   - una richiesta di get y e put y:a al server3,
func basicTestCE(rpcName string) {

	fmt.Println("In mediumTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put x:a e put y:b al server1\n" +
		"- una richiesta di get x e put x:b al server2\n" +
		"- una richiesta di get y e put y:a al server3")

	operations := [][]Operation{
		{
			{ServerIndex: 0, OperationType: put, Key: "x", Value: "a"},
			{ServerIndex: 0, OperationType: put, Key: "y", Value: "b"},
		},
		{
			{ServerIndex: 1, OperationType: get, Key: "x"},
			{ServerIndex: 1, OperationType: put, Key: "x", Value: "b"},
		},
		{
			{ServerIndex: 2, OperationType: get, Key: "y"},
			{ServerIndex: 2, OperationType: put, Key: "y", Value: "a"},
		},
	}

	testCausal(rpcName, operations)
}

// In questo mediumTestCE vengono inviate in goroutine:
//   - una richiesta di get y, get y, se leggo y:c -> put x:b e get y al server1,
//   - una richiesta di put y:b, get x, get y, get x al server2,
//   - una richiesta di get x, se leggo x:b -> put x:c, put y:c e get x al server3,
func mediumTestCE(rpcName string) {
	fmt.Println("In questo complexTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di get y, put x:b e get y al server1\n" +
		"- una richiesta di put y:b, get x, get y, get x al server2\n" +
		"- una richiesta di get x, put y:c e get x al server3")

	operations := [][]Operation{
		{
			{ServerIndex: 0, OperationType: get, Key: "y"},
			{ServerIndex: 0, OperationType: put, Key: "x", Value: "b"},
			{ServerIndex: 0, OperationType: get, Key: "y"},
		},
		{
			{ServerIndex: 1, OperationType: put, Key: "y", Value: "b"},
			{ServerIndex: 1, OperationType: get, Key: "x"},
			{ServerIndex: 1, OperationType: get, Key: "y"},
			{ServerIndex: 1, OperationType: get, Key: "x"},
		},
		{
			{ServerIndex: 2, OperationType: get, Key: "x"},
			{ServerIndex: 2, OperationType: put, Key: "y", Value: "c"},
			{ServerIndex: 2, OperationType: get, Key: "x"},
		},
	}

	testCausal(rpcName, operations)
}

// In questo mediumTestCE vengono inviate in goroutine:
//   - una richiesta di put x:a, put x:b, get x, put x:d al server1,
//   - una richiesta di get x, pu x:c, get x al server2,
//   - una richiesta di put x:a, get x, get x al server3,
func complexTestCE(rpcName string) {
	fmt.Println("In questo complexTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put x:a, put x:b, get x, put x:d al server1\n" +
		"- una richiesta di get x, put x:c, get x al server2\n" +
		"- una richiesta di put x:a, get x, get x al server3")

	operations := [][]Operation{
		{
			{ServerIndex: 0, OperationType: put, Key: "x", Value: "a"},
			{ServerIndex: 0, OperationType: put, Key: "x", Value: "b"},
			{ServerIndex: 0, OperationType: get, Key: "x"},
			{ServerIndex: 0, OperationType: put, Key: "x", Value: "d"},
		},
		{
			{ServerIndex: 1, OperationType: get, Key: "x"},
			{ServerIndex: 1, OperationType: put, Key: "x", Value: "c"},
			{ServerIndex: 1, OperationType: get, Key: "x"},
		},
		{
			{ServerIndex: 2, OperationType: put, Key: "x", Value: "a"},
			{ServerIndex: 2, OperationType: get, Key: "x"},
			{ServerIndex: 2, OperationType: get, Key: "x"},
		},
	}

	testCausal(rpcName, operations)
}
