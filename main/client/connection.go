package main

import (
	"fmt"
	"main/common"
	"math/rand"
	"net/rpc"
)

/* ----- FUNZIONI UTILIZZATE PER LA CONNESSIONE -----*/

// executeCall, esegue una chiamata RPC:
//   - index rappresenta l'indice del server a cui inviare la richiesta (da 0 a common.Replicas-1)
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
//   - args rappresenta gli argomenti da passare alla funzione RPC
//   - opType rappresenta il tipo di operazione da eseguire (synchronous o async)
//   - specOrRandom rappresenta se la chiamata deve essere effettuata ad un server specifico o ad un server casuale
//
// Ritorna la risposta ricevuta dal server e un eventuale errore
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

// executeRandomCall, esegue una chiamata RPC ad un server randomico:
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
//   - args rappresenta gli argomenti da passare alla funzione RPC
//   - response rappresenta la risposta ricevuta dal server
//   - opType rappresenta il tipo di operazione da eseguire (synchronous o async)
//
// Ritorna un eventuale errore
func executeRandomCall(rpcName string, args common.Args, response *common.Response, opType string) error {

	randomIndex := rand.Intn(common.Replicas) // Genera un indice casuale tra 0 e common.Replicas-1
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
//   - args rappresenta gli argomenti da passare alla funzione RPC
//   - response rappresenta la risposta ricevuta dal server
//   - opType rappresenta il tipo di operazione da eseguire (synchronous o async)
//
// Ritorna un eventuale errore
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

// connect: si connette al server con l'indice specificato
// Ritorna un client RPC e un eventuale errore
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

// syncCall esegue una chiamata RPC utilizzando il client RPC fornito.
//   - conn rappresenta il client RPC connesso al server
//   - index rappresenta l'indice del server a cui inviare la richiesta (da 0 a common.Replicas-1)
//   - args rappresenta gli argomenti da passare alla funzione RPC
//   - response rappresenta la risposta ricevuta dal server
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
//
// Ritorna un eventuale errore
func syncCall(conn *rpc.Client, index int, args common.Args, response *common.Response, rpcName string) error {

	waitToSend(index, args)

	// Effettua la chiamata RPC
	debugPrintRun(rpcName, args)

	err := conn.Call(rpcName, args, response)
	if err != nil {
		return fmt.Errorf("syncCall: errore durante la chiamata RPC in client.call: %s\n", err)
	}

	clientState.UnlockMutexSentMsg(index)

	waitToAcceptResponse(index, args, rpcName, response)

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("syncCall: errore durante la chiusura della connessione in client.call: %s\n", err)
	}
	return nil
}

// asyncCall esegue una chiamata RPC in modo asincrono utilizzando il client RPC fornito.
//   - conn rappresenta il client RPC connesso al server
//   - index rappresenta l'indice del server a cui inviare la richiesta (da 0 a common.Replicas-1)
//   - args rappresenta gli argomenti da passare alla funzione RPC
//   - response rappresenta la risposta ricevuta dal server
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
//
// Ritorna un eventuale errore
func asyncCall(conn *rpc.Client, index int, args common.Args, response *common.Response, rpcName string) error {

	waitToSend(index, args)

	// Effettua la chiamata RPC
	call := conn.Go(rpcName, args, response, nil)
	debugPrintRun(rpcName, args)

	clientState.UnlockMutexSentMsg(index)

	go func() {
		<-call.Done // Aspetta che la chiamata RPC sia completata

		waitToAcceptResponse(index, args, rpcName, response)

		if call.Error != nil {
			fmt.Printf("asyncCall: errore durante la chiamata RPC in client: %s\n", call.Error)
		}
		err := conn.Close()
		if err != nil {
			fmt.Printf("asyncCall: errore durante la chiusura della connessione in client: %s\n", err)
		}
	}()

	return nil
}

// waitToAcceptResponse: aspetta che il client abbia ricevuto il messaggio che si aspettava, controllando ReceiveMsgCounter
//   - index rappresenta l'indice del server a cui inviare la richiesta (da 0 a common.Replicas-1)
//   - args rappresenta gli argomenti da passare alla funzione RPC
//   - rpcName rappresenta il nome della funzione RPC da chiamare + il tipo di operazione (put, get, delete)
//   - response rappresenta la risposta ricevuta dal server
func waitToAcceptResponse(index int, args common.Args, rpcName string, response *common.Response) {
	for {
		// Se ho ricevuto dal server un messaggio che mi aspettavo
		if clientState.GetReceiveMsgCounter(index)+1 == response.GetReceptionFIFO() {

			clientState.MutexReceiveLock(index)
			clientState.IncreaseReceiveMsg(index)
			debugPrintResponse(rpcName, index, args, *response)
			clientState.MutexReceiveUnlock(index)
			return
		}
	}
}

// waitToSend: attende fin quando il client può inviare il messaggio che si aspettava, controllando SentMsgCounter
//   - index rappresenta l'indice del server a cui inviare la richiesta (da 0 a common.Replicas-1)
//   - args rappresenta gli argomenti da passare alla funzione RPC
func waitToSend(index int, args common.Args) {
	for {
		if clientState.GetSentMsgCounter(index) == args.GetSendingFIFO() {
			clientState.LockMutexSentMsg(index)
			clientState.IncreaseSentMsg(index)
			break
		}
	}
}
