package main

import (
	"fmt"
	"main/common"
)

var (
	firstRequest     = true                           // firstRequest è una variabile booleana che indica se è la prima richiesta
	err              error                            // err è una variabile di tipo error
	responses        [common.Replicas]common.Response // responses è un array di risposte
	listArgs         [common.Replicas]common.Args     // listArgs è un array di argomenti
	serverTimestamps = make(map[int]int)              // serverTimestamps è una mappa che associa a ogni server un timestamp
)

func testSequential(rpcName string, operations []Operation) {

	if firstRequest { // Inizializzazione

		for i := 0; i < common.Replicas; i++ {
			serverTimestamps[i] = 0
			listArgs[i] = common.NewArgs(serverTimestamps[i], "", "")
		}

		firstRequest = false
	}

	// endOps è un array di operazioni di tipo put che vengono eseguite su tutti i server
	endOps := getEndOps()

	for _, operation := range endOps {
		operations = append(operations, operation)
	}

	// executeCall esegue una chiamata RPC al server specificato e restituisce la risposta
	// in questo for vengono eseguite tutte le operazioni passate come argomento
	// async e specific sono due variabili booleane che indicano rispettivamente se la chiamata è sincrona o asincrona
	// e se è indirizzata ad un server specifico o random
	for _, op := range operations {

		args := listArgs[op.ServerIndex]

		args.SetSendingFIFO(serverTimestamps[op.ServerIndex])
		args.SetKey(op.Key)
		args.SetValue(op.Value)

		responses[op.ServerIndex], err = executeCall(op.ServerIndex, rpcName+op.OperationType, args, async, specific)

		serverTimestamps[op.ServerIndex]++
		//fmt.Println("Richiesta effettuata con ts", args.GetSendingFIFO(), "al server", op.ServerIndex, "nuovo ts", serverTimestamps[op.ServerIndex])

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
		{ServerIndex: 0, OperationType: del, Key: "x"},
		{ServerIndex: 1, OperationType: del, Key: "x"},
		{ServerIndex: 2, OperationType: del, Key: "x"},
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
