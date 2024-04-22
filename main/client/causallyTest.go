package main

import (
	"fmt"
	"main/common"
)

type CausalOperation struct {
	ServerIndex   int
	OperationType string
	Key           string
	Value         string
	//DependsOn     []int // IDs delle operazioni da cui dipende questa operazione
}

func testCausal(rpcName string, operations []CausalOperation) {
	fmt.Println("In questo test vengono inviate in sequenza rispettando le dipendenze causali:")
	//var operationChannels map[int]chan bool

	serverTimestamps := map[int]int{
		0: 0,
		1: 0,
		2: 0,
	}

	responses := make([]common.Response, 3)
	var err error

	// Assume operations are sorted in the order they should be executed
	for _ /*i*/, operation := range operations {
		// Wait for all operations this operation depends on to complete
		/*for _, dependency := range operation.DependsOn {
			// Wait for operation with ID 'dependency' to complete
			// This could be implemented with channels or other synchronization primitives
			<-operationChannels[dependency]
		}*/

		responses[operation.ServerIndex], err = executeCall(operation.ServerIndex, rpcName, operation.OperationType, operation.Key, operation.Value, serverTimestamps, Sync)
		if err != nil {
			fmt.Println("Errore durante l'esecuzione di executeCall")
			return
		}
		// Signal that this operation has completed
		// This could be implemented with channels or other synchronization primitives
		//operationChannels[i] <- true
	}

	endCall(rpcName, serverTimestamps)
}

// In questo basicTestCE vengono inviate in goroutine:
//   - una richiesta di put x:1 al server1,
//   - una richiesta di get x put y:2 al server2 (così da essere in relazione di causa-effetto)
func basicTestCE(rpcName string) {

	fmt.Println("In questo basicTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put x:1 al server1\n" +
		"- una richiesta di get x put y:2 al server2 (causa-effetto)")

	/*operations := []CausalOperation{
		{0, "put", "x", "1", nil},
		{ServerIndex: 1, OperationType: "get", Key: "x", DependsOn: []int{0}},
		{2, "put", "x", "2", []int{1}},
	}*/

	operations := []CausalOperation{
		{0, put, "x", "1"},
		{ServerIndex: 1, OperationType: get, Key: "x"},
		{2, get, "x", "2"},
	}
	testCausal(rpcName, operations)
}

// In questo mediumTestCE vengono inviate in goroutine:
//   - una richiesta di put x:a e put y:b al server1,
//   - una richiesta di get x e put x:b al server2,
//   - una richiesta di get y e put y:a al server3,
func mediumTestCE(rpcName string) {

	fmt.Println("In mediumTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put x:a e put y:b al server1\n" +
		"- una richiesta di get x e put x:b al server2\n" +
		"- una richiesta di get y e put y:a al server3")

	/*{0, "put", "x", "1", nil},
	{1, "get", "x", "", []int{0}},
	{2, "put", "x", "2", []int{1}},
	{0, "get", "x", "", []int{2}},
	{1, "put", "x", "3", []int{3}},*/

	operations := []CausalOperation{
		{0, put, "x", "1"},
		{ServerIndex: 1, OperationType: get, Key: "x"},
		{2, put, "x", "2"},
		{OperationType: get, Key: "x"},
		{1, put, "x", "3"},
	}
	testCausal(rpcName, operations)
}

// In questo complexTestCE vengono inviate in goroutine:
//   - una richiesta di get y, get y, se leggo y:c -> put x:b e get y al server1,
//   - una richiesta di put y:b, get x, get y, get x al server2,
//   - una richiesta di get x, se leggo x:b -> put x:c, put y:c e get x al server3,
func complexTestCE(rpcName string) {
	fmt.Println("In questo complexTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di get y, se leggo y:c -> put x:b e get y al server1\n" +
		"- una richiesta di put y:b, get x, get y, get x al server2\n" +
		"- una richiesta di get x se leggo x:b -> put x:c, put y:c e get x al server3")

	/*operations := []CausalOperation{
		{0, put, "x", "1", nil},
		{1, get, "x", "", []int{0}},
		{2, put, "x", "2", []int{1}},
		{0, get, "x", "", []int{2}},
		{1, put, "x", "3", []int{3}},
		{2, get, "x", "", []int{4}},
		{0, put, "x", "4", []int{5}},
		{1, get, "x", "", []int{6}},
	}*/

	operations := []CausalOperation{
		{0, put, "x", "1"},
		{1, get, "x", ""},
		{2, put, "x", "2"},
		{0, get, "x", ""},
		{1, put, "x", "3"},
		{2, get, "x", ""},
		{0, put, "x", "4"},
		{1, get, "x", ""},
	}

	testCausal(rpcName, operations)
}
