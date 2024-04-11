// client.go
package main

import (
	"fmt"
	"main/common"
	"math/rand"
	"net/rpc"
)

const (
	put    = ".Put"
	get    = ".Get"
	delete = ".Delete"
)

func main() {

	for {
		// Stampa il menu interattivo
		fmt.Println("Scegli un'operazione:")
		fmt.Println("1. Consistenza Causale")
		fmt.Println("2. Consistenza Sequenziale")

		// Leggi l'input dell'utente per l'operazione
		fmt.Print("\nInserisci il numero dell'operazione desiderata: ")
		var choice int
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("Client -> Errore durante la lettura dell'input:", err)
			break
		}

		var rpcName string
		switch choice {
		case 1:
			fmt.Println("Scelta di consistenza causale")
			rpcName = "KeyValueStoreCausale"
			break
		case 2:
			fmt.Println("Scelta di consistenza sequenziale")
			rpcName = "KeyValueStoreSequential"
			break
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}

		// Esecuzione delle rpc
		done := make(chan bool)
		go run1(rpcName, done)
		go run2(rpcName, done)
		go run3(rpcName, done)
		go run4(rpcName, done)

		// Attendi il completamento di tutte le goroutine
		for i := 0; i < 4; i++ {
			<-done
		}
	}
}

// randomConnect restituisce una connessione random con un server definito in common/config.go
func randomConnect() *rpc.Client {
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(common.Replicas)

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[randomIndex], randomIndex)

	//fmt.Println("CLIENT: Contatto il server", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

// specificConnect restituisce una connessione specifica con un server definito in common/config.go, tramite index passato in argomento
func specificConnect(index int) *rpc.Client {
	if index >= common.Replicas {
		_ = fmt.Errorf("index out of range")
		return nil
	}

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[index], index)

	//fmt.Println("CLIENT: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

// executeCall esegue un comando ad un server random. Il comando da eseguire viene specificato tramite i parametri inseriti
func executeCall(rpcName, key string, values ...string) {
	var value string
	if len(values) > 0 {
		value = values[0]
	}

	args := common.Args{Key: key, Value: value}
	reply := common.Response{}
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}

	// TODO: Qui posso usare un id auto-incrementativo per un DEBUG accurato
	fmt.Println("Run", rpcName, key+":"+value)
	err := delayedCall(conn, args, &reply, rpcName)
	if err != nil {
		return
	}

}

// delayedCall esegue una chiamata RPC ritardata utilizzando il client RPC fornito.
// Prima di effettuare la chiamata, applica un ritardo casuale per simulare condizioni reali di rete.
func delayedCall(conn *rpc.Client, args common.Args, reply *common.Response, rpcName string) error {
	// Applica un ritardo casuale
	common.RandomDelay()

	// Effettua la chiamata RPC
	err := conn.Call(rpcName, args, reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC in client.call")
	}
	return nil
}

func run1(rpcName string, done chan bool) {

	executeCall(rpcName+put, "y", "0")
	executeCall(rpcName+put, "x", "1")
	executeCall(rpcName+get, "x")
	executeCall(rpcName+get, "y")

	done <- true
}

func run2(rpcName string, done chan bool) {

	executeCall(rpcName+put, "y", "1")
	executeCall(rpcName+get, "y")
	executeCall(rpcName+get, "x")

	done <- true
}

func run3(rpcName string, done chan bool) {
	executeCall(rpcName+get, "x")
	executeCall(rpcName+get, "y")
	done <- true
}

func run4(rpcName string, done chan bool) {
	executeCall(rpcName+put, "x", "0")
	executeCall(rpcName+get, "y")
	executeCall(rpcName+get, "x")

	done <- true
}
