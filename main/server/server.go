// server.go
package main

import (
	"fmt"
	"main/common"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

// server.go
// Questo è il codice dei server replica, che vengono contattati dai client tramite una procedura rpc
// e rispondono in maniera adeguata.

func main() {

	/*---LOCALE---*/
	// Legge l'argomento passato
	/*if len(os.Args) < 2 {
		fmt.Println("Usare: ", os.Args[0], "<Porta di ascolto>", "<ID_Server>")
		os.Exit(1)
	}

	// Legge l'ID del server passato come argomento dalla riga di comando
	idStr := os.Args[1]*/

	/*---DOCKER---*/
	// Ottieni la porta da una variabile d'ambiente o assegna un valore predefinito
	idStr := os.Getenv("SERVER_ID")

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
		vectorClock: make([]int, 0), // Inizializzazione dell'orologio vettoriale
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

	// Accettazione delle connessioni in arrivo
	rpc.Accept(listener)
}
