// server.go
package main

import (
	"fmt"
	"main/common"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

// server.go
// Questo è il codice dei server replica, che vengono contattati dai client tramite una procedura rpc
// e rispondono in maniera adeguata.

func main() {

	var idStr string

	if os.Getenv("CONFIG") == "1" {
		/*---LOCALE---*/
		if len(os.Args) < 2 { //Legge l'argomento passato
			fmt.Println("Usare: ", os.Args[0], "<ID_Server>")
			os.Exit(1)
		}

		// Legge l'ID del server passato come argomento dalla riga di comando
		idStr = os.Args[1]
	} else {
		/*---DOCKER---*/
		// Ottieni la porta da una variabile d'ambiente o assegna un valore predefinito
		idStr = os.Getenv("SERVER_ID")
	}

	// Converti l'ID del server in un intero
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Errore:", err)
		os.Exit(1)
	}
	port := common.ReplicaPorts[id]

	// Inizializzazione delle strutture KeyValueStoreCausale e KeyValueStoreSequential

	// ----- CONSISTENZA CAUSALE -----
	kvCausale := &KeyValueStoreCausale{
		datastore:   make(map[string]string),
		vectorClock: [common.Replicas]int{}, // Array di lunghezza fissa inizializzato a zero
		queue:       make([]MessageC, 0),
		id:          id,
	}

	// ----- CONSISTENZA SEQUENZIALE -----
	kvSequential := &KeyValueStoreSequential{
		datastore:    make(map[string]string),
		logicalClock: 0, // Inizializzazione dell'orologio logico scalare
		queue:        make([]Message, 0),
	}

	// Registrazione dei servizi RPC
	err = rpc.RegisterName("KeyValueStoreCausale", kvCausale)
	if err != nil {
		fmt.Println("SERVER: Errore durante la registrazione di KeyValueStoreCausale", err)
		return
	}
	err = rpc.RegisterName("KeyValueStoreSequential", kvSequential)
	if err != nil {
		fmt.Println("SERVER: Errore durante la registrazione di KeyValueStoreSequential", err)
		return
	}

	// Avvio del listener RPC sulla porta specificata
	fmt.Println("LA MIA PORTA", port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("SERVER: Errore nell'avvio del listener RPC:", err)
		return
	}

	// Avvio della goroutine per stampare il datastore
	go printQueue(kvSequential)
	go printDatastore(kvSequential)

	// Ciclo per accettare e gestire le connessioni in arrivo
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("SERVER: Errore nell'accettare la connessione:", err)
			continue
		}

		// Avvia la gestione della connessione in un goroutine
		go rpc.ServeConn(conn)
	}
}

// Funzione per stampare il datastore ogni 500 millisecondi
func printQueue(kv *KeyValueStoreSequential) {
	for {
		time.Sleep(1500 * time.Millisecond)

		// Controllo se il datastore è vuoto
		if len(kv.queue) == 0 {
			fmt.Println("Queue vuota")
			continue // Salta alla prossima iterazione del ciclo
		}

		fmt.Println("Queue:", kv.queue)
	}
}

// Funzione per stampare il datastore ogni 500 millisecondi
func printDatastore(kv *KeyValueStoreSequential) {
	for {
		time.Sleep(1500 * time.Millisecond)

		// Controllo se il datastore è vuoto
		if len(kv.datastore) == 0 {
			fmt.Println("Datastore vuota")
			continue // Salta alla prossima iterazione del ciclo
		}

		fmt.Println("Datastore:", kv.datastore)
	}
}
