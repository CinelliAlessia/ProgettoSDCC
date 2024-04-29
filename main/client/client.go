// client.go
package main

import (
	"fmt"
	"main/common"
	"math/rand"
	"net/rpc"
	"time"
)

func main() {

	for {
		// Stampa il menu interattivo
		fmt.Println("Scegli un'operazione:")
		fmt.Println("1. Consistenza Causale")
		fmt.Println("2. Consistenza Sequenziale")

		// Leggi l'input dell'utente per l'operazione
		fmt.Print("Inserisci il numero dell'operazione desiderata: ")
		var choice int
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("Errore durante la lettura dell'input:", err)
			break
		}

		var rpcName string
		switch choice {
		case 1:
			rpcName = "KeyValueStoreCausale"
			causal(rpcName)
			break
		case 2:
			rpcName = "KeyValueStoreSequential"
			sequential(rpcName)
			break
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}
	}
}

// sequential() Scegliere il tipo di test che si vuole eseguire per verificare le garanzie di consistenza sequenziale
func sequential(rpcName string) {
	for {
		// Stampa il menu interattivo
		fmt.Println("Consistenza sequenziale, scegli il test da eseguire: ")
		fmt.Println("1. Basic Test")
		fmt.Println("2. Medium Test")
		fmt.Println("3. Complex Test")
		fmt.Println("4. Esci")

		// Leggi l'input dell'utente per l'operazione
		fmt.Print("Inserisci il numero dell'operazione desiderata: ")
		var choice int
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("Errore durante la lettura dell'input:", err)
			return
		}
		fmt.Println()

		switch choice {
		case 1:
			basicTestSeq(rpcName)
			break
		case 2:
			mediumTestSeq(rpcName)
			break
		case 3:
			complexTestSeq(rpcName)
			break
		case 4:
			fmt.Println()
			return
		}
		time.Sleep(5000 * time.Millisecond) // Aggiungo un ritardo per evitare che le stampe si sovrappongano
	}
}

// causal() Scegliere il tipo di test che si vuole eseguire per verificare le garanzie di consistenza causale
func causal(rpcName string) {
	for { // Stampa il menu interattivo
		fmt.Println("Consistenza causale, scegli il test da eseguire: ")
		fmt.Println("1. Basic Test")
		fmt.Println("2. Medium Test")
		fmt.Println("3. Complex Test")
		fmt.Println("4. Esci")

		// Leggi l'input dell'utente per l'operazione
		fmt.Print("Inserisci il numero dell'operazione desiderata: ")
		var choice int
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("Errore durante la lettura dell'input:", err)
			return
		}
		fmt.Println()

		switch choice {
		case 1:
			basicTestCE(rpcName)
			break
		case 2:
			mediumTestCE(rpcName)
			break
		case 3:
			complexTestCE(rpcName)
			break
		case 4:
			fmt.Println()
			return
		}

		time.Sleep(5000 * time.Millisecond) // Aggiungo un ritardo per evitare che le stampe si sovrappongano
	}
}

/* ----- FUNZIONI UTILIZZATE PER LA CONNESSIONE -----*/

// randomConnect restituisce una connessione random con un server definito in common/config.go
func randomConnect() *rpc.Client {
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(common.Replicas)

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[randomIndex], randomIndex)

	//fmt.Println("CLIENT: Contatto il server", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("randomConnect: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

// specificConnect restituisce una connessione specifica con un server definito in common/config.go, tramite index passato in argomento
// Attenzione: Indicizzazione parte da 0!
func specificConnect(index int) *rpc.Client {
	if index > common.Replicas {
		_ = fmt.Errorf("index out of range")
		return nil
	}

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[index], index)

	//fmt.Println("CLIENT: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("specificConnect: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

// executeCall esegue un comando ad un server specifico. Il comando da eseguire viene specificato tramite i parametri inseriti
// si occupa di eseguire le operazioni di put, get e del, in maniera sync o async e incrementa il timestamp dello specifico client
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
	conn := randomConnect()
	if conn == nil {
		return fmt.Errorf("executeRandomCall: Errore durante la connessione con un server random\n")
	}

	switch opType {
	case sync:
		// Esegui l'operazione in modo sincrono
		err := syncCall(conn, args, response, rpcName)
		if err != nil {
			return err
		}
	case async:
		// Esegui l'operazione in modo asincrono
		err := asyncCall(conn, args, response, rpcName)
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

	conn := specificConnect(index)
	if conn == nil {
		return fmt.Errorf("executeSpecificCall: Errore durante la connessione con il server specifico %d\n", index)
	}

	switch opType {
	case sync:
		// Esegui l'operazione in modo sincrono
		err := syncCall(conn, args, response, rpcName)
		if err != nil {
			return err
		}
	case async:
		// Esegui l'operazione in modo asincrono
		err := asyncCall(conn, args, response, rpcName)
		if err != nil {
			return err
		}
	default:
		// Gestisci i casi non previsti
	}
	return nil
}

// syncCall esegue una chiamata RPC ritardata utilizzando il client RPC fornito.
// Prima di effettuare la chiamata, applica un ritardo casuale per simulare condizioni reali di rete.
func syncCall(conn *rpc.Client, args common.Args, response *common.Response, rpcName string) error {

	debugPrintRun(rpcName, args)

	// Effettua la chiamata RPC
	err := conn.Call(rpcName, args, response)
	if err != nil {
		return fmt.Errorf("syncCall: errore durante la chiamata RPC in client.call: %s\n", err)
	}

	debugPrintResponse(rpcName, args, *response)

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("syncCall: errore durante la chiusura della connessione in client.call: %s\n", err)
	}
	return nil
}

// asyncCall: utilizzato per lsa consistenza sequenziale
func asyncCall(conn *rpc.Client, args common.Args, response *common.Response, rpcName string) error {

	// Effettua la chiamata RPC
	call := conn.Go(rpcName, args, response, nil)
	debugPrintRun(rpcName, args)

	go func() {
		<-call.Done // Aspetta che la chiamata RPC sia completata
		debugPrintResponse(rpcName, args, *response)

		if call.Error != nil {
			fmt.Printf("asyncCall: errore durante la chiamata RPC in client: %s\n", call.Error)
			response.Done <- false
		} else {
			//debugPrintResponse(rpcName, args, *response)
			response.Done <- true
		}

		err := conn.Close()
		if err != nil {
			fmt.Printf("asyncCall: errore durante la chiusura della connessione in client: %s\n", err)
		}
	}()

	return nil
}
