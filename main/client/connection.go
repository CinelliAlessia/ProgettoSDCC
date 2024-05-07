package main

import (
	"fmt"
	"main/common"
	"math/rand"
	"net/rpc"
	"strconv"
)

/* ----- FUNZIONI UTILIZZATE PER LA CONNESSIONE -----*/

// executeCall esegue un comando ad un server. Il comando da eseguire viene specificato tramite i parametri inseriti
// si occupa di eseguire le operazioni di put, get e del, in maniera synchronous o async
func executeCall(index int, rpcName string, args common.Args, opType string, specOrRandom string) (common.Response, error) {
	response := common.Response{}
	var err error

	switch specOrRandom {
	case specific:
		err = executeSpecificCall(index, rpcName, args, &response, opType)
	case random:
		err = executeRandomCall(rpcName, args, &response, opType)
	}

	if err != nil {
		return response, err
	}

	return response, nil
}

// executeRandomCall:
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
func executeRandomCall(rpcName string, args common.Args, response *common.Response, opType string) error {
	randomIndex := rand.Intn(common.Replicas)

	conn, err := connect(randomIndex)
	if conn == nil || err != nil {
		return fmt.Errorf("executeRandomCall: Errore durante la connessione con il server specifico %d, %v\n",
			randomIndex, err)
	}

	switch opType {
	case synchronous:
		// Esegui l'operazione in modo sincrono
		err := syncCall(conn, randomIndex, args, response, rpcName)
		if err != nil {
			return err
		}
	case async:
		// Esegui l'operazione in modo asincrono
		err := asyncCall(conn, randomIndex, args, response, rpcName)
		if err != nil {
			return err
		}
	default:
		// Gestisci i casi non previsti
	}

	return nil
}

// executeSpecificCall:
//   - index rappresenta l'indice relativo al server da contattare, da 0 a (common.Replicas-1)
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
func executeSpecificCall(index int, rpcName string, args common.Args, response *common.Response, opType string) error {

	conn, err := connect(index)
	if conn == nil || err != nil {
		return fmt.Errorf("executeSpecificCall: Errore durante la connessione con il server specifico %d, %v\n",
			index, err)
	}

	switch opType {
	case synchronous:
		// Esegui l'operazione in modo sincrono
		err := syncCall(conn, index, args, response, rpcName)
		if err != nil {
			return err
		}
	case async:
		// Esegui l'operazione in modo asincrono
		err := asyncCall(conn, index, args, response, rpcName)
		if err != nil {
			return err
		}
	default:
		// Gestisci i casi non previsti
	}
	return nil
}

func connect(index int) (*rpc.Client, error) {
	if index > common.Replicas {
		return nil, fmt.Errorf("connect: index out of range")
	}

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[index], index)

	//fmt.Println("CLIENT: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)
	if err != nil {
		return nil, fmt.Errorf("specificConnect: Errore durante la connessione al server %v\n", err)
	}
	return conn, err
}

// syncCall esegue una chiamata RPC ritardata utilizzando il client RPC fornito.
// Prima di effettuare la chiamata, applica un ritardo casuale per simulare condizioni reali di rete.
func syncCall(conn *rpc.Client, index int, args common.Args, response *common.Response, rpcName string) error {

	for {
		if clientState.GetSendIndex(index) == args.GetSendingFIFO() {
			clientState.MutexSent[index].Lock()
			clientState.SendIndex[index]++
			break
		}
	}

	// Effettua la chiamata RPC
	debugPrintRun(rpcName, args)
	err := conn.Call(rpcName, args, response)
	if err != nil {
		return fmt.Errorf("syncCall: errore durante la chiamata RPC in client.call: %s\n", err)
	}

	clientState.MutexSent[index].Unlock()

	//waitToAccept(index, args, rpcName, response)
	debugPrintResponse(rpcName, strconv.Itoa(index), args, *response)

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("syncCall: errore durante la chiusura della connessione in client.call: %s\n", err)
	}
	return nil
}

// asyncCall: utilizzato per lsa consistenza sequenziale
func asyncCall(conn *rpc.Client, index int, args common.Args, response *common.Response, rpcName string) error {

	for {
		if clientState.GetSendIndex(index) == args.GetSendingFIFO() {
			clientState.MutexSent[index].Lock()
			clientState.SendIndex[index]++
			break
		}
	}

	// Effettua la chiamata RPC
	call := conn.Go(rpcName, args, response, nil)
	debugPrintRun(rpcName, args)

	clientState.MutexSent[index].Unlock()

	go func() {
		<-call.Done // Aspetta che la chiamata RPC sia completata

		waitToAccept(index, args, rpcName, response)

		if call.Error != nil {
			fmt.Printf("asyncCall: errore durante la chiamata RPC in client: %s\n", call.Error)
			response.SetDone(false)
		} else {
			response.SetDone(true)
		}

		err := conn.Close()
		if err != nil {
			fmt.Printf("asyncCall: errore durante la chiusura della connessione in client: %s\n", err)
		}
	}()

	return nil
}

func waitToAccept(index int, args common.Args, rpcName string, response *common.Response) {
	for {
		// Se ho ricevuto dal server un messaggio che mi aspettavo
		if clientState.GetReceiveIndex(index) == response.GetReceptionFIFO() {

			clientState.MutexReceive[index].Lock()
			clientState.ReceiveIndex[index]++
			debugPrintResponse(rpcName, strconv.Itoa(index), args, *response)
			clientState.MutexReceive[index].Unlock()
			return

		}
	}
}
