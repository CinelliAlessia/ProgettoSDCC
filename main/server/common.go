package main

import (
	"fmt"
	"main/common"
	"net/rpc"
)

func sendToAllServer(rpcName string, message interface{}, response *common.Response) error {
	// Canale per ricevere i risultati delle chiamate RPC
	resultChan := make(chan error, common.Replicas)

	// Itera su tutte le repliche e avvia le chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		go func(i int) {
			switch msg := message.(type) {
			case MessageC:
				callRPC(rpcName, msg, response, resultChan, i)
			case MessageS:
				callRPC(rpcName, msg, response, resultChan, i)
			default:
				resultChan <- fmt.Errorf("unsupported message type: %T", msg)
			}
		}(i)
	}

	// Raccoglie i risultati dalle chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		if err := <-resultChan; err != nil {
			return err
		}
	}
	return nil
}

func callRPC(rpcName string, message interface{}, response *common.Response, resultChan chan<- error, replicaIndex int) {
	serverName := common.GetServerName(common.ReplicaPorts[replicaIndex], replicaIndex)

	conn, err := rpc.Dial("tcp", serverName)
	if err != nil {
		resultChan <- fmt.Errorf("errore durante la connessione con %s: %v", serverName, err)
		return
	}

	common.RandomDelay()
	switch msg := message.(type) {
	case MessageC:
		err = conn.Call(rpcName, msg, response)
	case MessageS:
		err = conn.Call(rpcName, msg, response)
	default:
		resultChan <- fmt.Errorf("unsupported message type: %T", msg)
		return
	}

	if err != nil {
		resultChan <- fmt.Errorf("errore durante la chiamata RPC %s a %s: %v", rpcName, serverName, err)
		return
	}

	err = conn.Close()
	if err != nil {
		resultChan <- fmt.Errorf("errore durante la connessione in KeyValueStoreCausale.callRPC: %s", err)
		return
	}

	resultChan <- nil
}
