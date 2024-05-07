package main

import (
	"fmt"
	"main/common"
)

var (
	firstRequestS = true                           // firstRequestS è una variabile booleana che indica se è la prima richiesta
	firstRequestC = true                           // firstRequestC è una variabile booleana che indica se è la prima richiesta
	responses     [common.Replicas]common.Response // responses è un array di risposte
	listArgs      [common.Replicas]common.Args     // listArgs è un array di argomenti ciascuna per un client
	sendingTS     = make(map[int]int)              // sendingTS è una mappa che associa a ogni server un timestamp di invio,
	// allegato successivamente a args. Per assunzione FIFO Ordering in andata
)

func testSequential(rpcName string, operations []Operation) {

	if firstRequestS { // Inizializzazione

		for i := 0; i < common.Replicas; i++ {
			sendingTS[i] = 0
			listArgs[i] = common.NewArgs(sendingTS[i], "", "")
		}

		firstRequestS = false
	}

	// endOps è un array di operazioni di tipo put che vengono eseguite su tutti i server
	endOps := getEndOps()

	for _, operation := range endOps {
		operations = append(operations, operation)
	}

	var err error

	// executeCall esegue una chiamata RPC al server specificato e restituisce la risposta
	// in questo for vengono eseguite tutte le operazioni passate come argomento
	// async e specific sono due variabili booleane che indicano rispettivamente se la chiamata è sincrona o asincrona
	// e se è indirizzata ad un server specifico o random
	for _, op := range operations {

		args := listArgs[op.ServerIndex]

		args.SetSendingFIFO(sendingTS[op.ServerIndex])
		args.SetKey(op.Key)
		args.SetValue(op.Value)

		responses[op.ServerIndex], err = executeCall(op.ServerIndex, rpcName+op.OperationType, args, async, specific)

		sendingTS[op.ServerIndex]++ // Incremento il contatore di timestamp

		//fmt.Println("Richiesta effettuata con ts", args.GetSendingFIFO(), "al server", op.ServerIndex, "nuovo ts", sendingTS[op.ServerIndex])

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
		endOps = append(endOps, Operation{i, common.PutRPC, common.EndKey, common.EndValue})
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
		{0, common.PutRPC, "x", "1"},
		{1, common.PutRPC, "x", "2"},
		{2, common.PutRPC, "x", "3"},

		{ServerIndex: 0, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 2, OperationType: common.GetRPC, Key: "x"},

		{ServerIndex: 0, OperationType: common.DelRPC, Key: "x"},
		{ServerIndex: 1, OperationType: common.DelRPC, Key: "x"},
		{ServerIndex: 2, OperationType: common.DelRPC, Key: "x"},

		{ServerIndex: 0, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 2, OperationType: common.GetRPC, Key: "x"},
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
		{0, common.PutRPC, "x", "1"},
		{1, common.PutRPC, "x", "2"},
		{2, common.PutRPC, "x", "3"},

		{0, common.PutRPC, "y", "1"},
		{1, common.PutRPC, "y", "2"},
		{2, common.PutRPC, "y", "3"},

		{0, common.PutRPC, "z", "1"},
		{1, common.PutRPC, "z", "2"},
		{2, common.PutRPC, "z", "3"},

		{ServerIndex: 0, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 1, OperationType: common.GetRPC, Key: "y"},
		{ServerIndex: 2, OperationType: common.GetRPC, Key: "z"},
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
		{0, common.PutRPC, "x", "1"},
		{1, common.PutRPC, "x", "2"},
		{2, common.PutRPC, "x", "3"},

		{0, common.PutRPC, "y", "2"},
		{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 2, OperationType: common.GetRPC, Key: "x"},

		{ServerIndex: 0, OperationType: common.GetRPC, Key: "x"},
		{ServerIndex: 1, OperationType: common.GetRPC, Key: "y"},
		{ServerIndex: 2, OperationType: common.GetRPC, Key: "y"},

		{0, common.PutRPC, "y", "1"},
		{1, common.PutRPC, "x", "3"},
		{2, common.PutRPC, "x", "4"},
	}
	testSequential(rpcName, operations)
}
