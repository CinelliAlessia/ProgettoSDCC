package main

import (
	"fmt"
	"main/common"
)

// testCausal esegue una serie di operazioni passate in argomento
// Prende come argomento una lista di liste, dove ogni lista rappresenta una sequenza di operazioni
// Le operazioni vengono eseguite in ordine, in base all'ordine in cui sono state inserite ma in maniera concorrente
// rispetto agli altri server
func testCausal(rpcName string, operations [][]Operation) {

	if clientState.GetFirstRequest() { // Inizializzazione

		for i := 0; i < common.Replicas; i++ {
			clientState.SetSentMsgCounter(i, 0)
			clientState.SetListArgs(i, common.NewArgs(clientState.GetSendingTS(i), "", ""))
		}

		clientState.SetFirstRequest(false)
	}

	var err error

	for _, operation := range operations {

		go func(operation []Operation) {
			for _, op := range operation {

				args := clientState.GetListArgs(op.ServerIndex)

				args.SetSendingFIFO(clientState.GetSendingTS(op.ServerIndex))
				args.SetKey(op.Key)
				args.SetValue(op.Value)

				_, err = executeCall(op.ServerIndex, rpcName+op.OperationType, args, synchronous, specific)

				clientState.IncreaseSendingTS(op.ServerIndex) // Incremento il contatore dei messaggi inviati

				if err != nil {
					fmt.Println("testCausal: Errore durante l'esecuzione di executeCall:", err)
					return
				}
			}
		}(operation)
	}
}

// In questo basicTestCE vengono inviate in goroutine:
//   - una richiesta di put x:a e put y:b al server1,
//   - una richiesta di get x e put x:b al server2,
//   - una richiesta di get y e put y:a al server3,
func basicTestCE(rpcName string) {

	fmt.Println("Il test basico sulla consistenza causale invia in goroutine:\n" +
		"- una richiesta di put x:a e put y:b al server1\n" +
		"- una richiesta di get x e put x:b al server2\n" +
		"- una richiesta di get y e put y:a al server3")

	operations := [][]Operation{
		{
			{ServerIndex: 0, OperationType: common.PutRPC, Key: "x", Value: "a"},
			{ServerIndex: 0, OperationType: common.PutRPC, Key: "y", Value: "b"},
		},
		{
			{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
			{ServerIndex: 1, OperationType: common.PutRPC, Key: "x", Value: "b"},
		},
		{
			{ServerIndex: 2, OperationType: common.GetRPC, Key: "y"},
			{ServerIndex: 2, OperationType: common.PutRPC, Key: "y", Value: "a"},
		},
	}

	testCausal(rpcName, operations)
}

// In questo mediumTestCE vengono inviate in goroutine:
//   - una richiesta di put x:a, get y, put x:c al server1,
//   - una richiesta di get x, put x:b, get x al server2,
//   - una richiesta di put x:a, del x, put y:a, get x al server3.
func mediumTestCE(rpcName string) {
	fmt.Println("In questo mediumTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put x:a, get y, put x:c al server1\n" +
		"- una richiesta di get x, put x:b, get x al server2\n" +
		"- una richiesta di put x:a, del x, put y:a, get x al server3")

	operations := [][]Operation{
		{
			{ServerIndex: 0, OperationType: common.PutRPC, Key: "x", Value: "a"},
			{ServerIndex: 0, OperationType: common.GetRPC, Key: "y"},
			{ServerIndex: 0, OperationType: common.PutRPC, Key: "x", Value: "c"},
		},
		{
			{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
			{ServerIndex: 1, OperationType: common.PutRPC, Key: "x", Value: "b"},
			{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
		},
		{
			{ServerIndex: 2, OperationType: common.PutRPC, Key: "x", Value: "a"},
			{ServerIndex: 2, OperationType: common.DelRPC, Key: "x"},
			{ServerIndex: 2, OperationType: common.PutRPC, Key: "y", Value: "a"},
			{ServerIndex: 2, OperationType: common.GetRPC, Key: "x"},
		},
	}

	testCausal(rpcName, operations)
}

// In questo complexTestCE vengono inviate in goroutine:
//   - una richiesta di	get x, put x:b, get x, del z al server1,
//   - una richiesta di put x:a, put z:b, get x, put x:c al server2,
//   - una richiesta di get x, put y:c, get y, put y:c al server3.
func complexTestCE(rpcName string) {
	fmt.Println("In questo complexTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di get x, put x:b, get x, del z al server1\n" +
		"- una richiesta di put x:a, put z:b, get x, put x:c al server2\n" +
		"- una richiesta di get x, put y:c, get y, put y:c al server3")

	operations := [][]Operation{
		{
			{ServerIndex: 0, OperationType: common.GetRPC, Key: "x"},
			{ServerIndex: 0, OperationType: common.PutRPC, Key: "x", Value: "b"},
			{ServerIndex: 0, OperationType: common.GetRPC, Key: "y"},
			{ServerIndex: 0, OperationType: common.DelRPC, Key: "z"},
		},
		{
			{ServerIndex: 1, OperationType: common.PutRPC, Key: "x", Value: "a"},
			{ServerIndex: 1, OperationType: common.PutRPC, Key: "z", Value: "b"},
			{ServerIndex: 1, OperationType: common.GetRPC, Key: "x"},
			{ServerIndex: 1, OperationType: common.PutRPC, Key: "x", Value: "c"},
		},
		{
			{ServerIndex: 2, OperationType: common.GetRPC, Key: "x"},
			{ServerIndex: 2, OperationType: common.PutRPC, Key: "y", Value: "c"},
			{ServerIndex: 2, OperationType: common.GetRPC, Key: "y"},
			{ServerIndex: 2, OperationType: common.PutRPC, Key: "y", Value: "a"},
		},
	}

	testCausal(rpcName, operations)
}
