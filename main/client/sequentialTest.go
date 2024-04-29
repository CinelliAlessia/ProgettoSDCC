package main

import (
	"fmt"
	"main/common"
)

func testSequential(rpcName string, operations []Operation) {

	// serverTimestamps è una mappa che associa a ogni server un timestamp
	serverTimestamps := make(map[int]int)
	for i := 0; i < common.Replicas; i++ {
		serverTimestamps[i] = 0
	}
	responses := [common.Replicas]common.Response{}
	var err error

	// executeCall esegue una chiamata RPC al server specificato e restituisce la risposta
	// in questo for vengono eseguite tutte le operazioni passate come argomento
	// async e specific sono due variabili booleane che indicano rispettivamente se la chiamata è sincrona o asincrona
	// e se è indirizzata ad un server specifico o random
	for _, operation := range operations {

		args := common.Args{Key: operation.Key, Value: operation.Value, Timestamp: serverTimestamps[operation.ServerIndex]}
		responses[operation.ServerIndex], err = executeCall(operation.ServerIndex, rpcName+operation.OperationType, args, async, specific)
		serverTimestamps[operation.ServerIndex]++

		if err != nil {
			fmt.Println("testSequential: Errore durante l'esecuzione di executeCall", err)
			return
		}
	}

	// ----- Da qui inseriamo le operazioni di fine ----- //

	// timestamp è il massimo tra i timestamp dei server + 1
	timestamp := 0
	for i, _ := range serverTimestamps {
		timestamp = common.Max(serverTimestamps[i], timestamp)
	}
	timestamp += 1

	// endOps è un array di operazioni di tipo put che vengono eseguite su tutti i server
	endOps := getEndOps()
	for _, operation := range endOps {
		args := common.Args{Key: operation.Key, Value: operation.Value, Timestamp: timestamp}
		_, err = executeCall(operation.ServerIndex, rpcName+operation.OperationType, args, async, specific)
		serverTimestamps[operation.ServerIndex] = timestamp + 1

		if err != nil {
			fmt.Println("testSequential: Errore durante l'esecuzione di executeCall", err)
			return
		}
	}
}

// getEndOps restituisce un array di operazioni di tipo put che vengono eseguite su tutti i server
func getEndOps() []Operation {
	var endOps []Operation
	for i := 0; i < common.Replicas; i++ {
		endOps = append(endOps, Operation{i, put, common.EndKey, common.EndValue})
	}
	return endOps
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
//   - put x:3, get x, get y al server3.
func complexTestSeq(rpcName string) {
	fmt.Println("In questo complexTestSeq vengono inviate in sequenza:\n" +
		"- put x:1, put y:2, get x, put y:1 al server1\n" +
		"- put x:2, get x, get y, put x:3 al server2\n" +
		"- put x:3, get x, get y al server3")

	operations := []Operation{
		{0, put, "x", "1"},
		{1, put, "x", "2"},
		{2, put, "x", "3"},
		{0, put, "y", "2"},
		{ServerIndex: 1, OperationType: get, Key: "x"},
		{ServerIndex: 2, OperationType: get, Key: "x"},
		{ServerIndex: 0, OperationType: get, Key: "x"},
		{ServerIndex: 1, OperationType: get, Key: "y"},
		{ServerIndex: 2, OperationType: get, Key: "y"},
		{0, put, "y", "1"},
		{1, put, "x", "3"},
		{2, put, "x", "4"},
	}
	testSequential(rpcName, operations)
}
