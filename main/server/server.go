// server.go
package main

import (
	"fmt"
	"net"
	"net/rpc"
)

// server.go
// Questo è il codice dei server replica, che vengono contattati dai client tramite una procedura rpc
// e rispondono in maniera adeguata.

func main() {
	// Inizializzazione delle strutture KeyValueStoreCausale e KeyValueStoreSequential

	// ----- CONSISTENZA CAUSALE -----
	kvCausale := &KeyValueStoreCausale{
		datastore:   make(map[string]string),
		vectorClock: make([]int, 0), // Inizializzazione dell'orologio vettoriale
		queue:       make([]MessageC, 0),
	}

	// ----- CONSISTENZA SEQUENZIALE -----
	kvSequential := &KeyValueStoreSequential{
		datastore:    make(map[string]string),
		logicalClock: 0, // Inizializzazione dell'orologio logico scalare
		queue:        make([]Message, 0),
	}

	// Registrazione dei servizi RPC
	err := rpc.RegisterName("KeyValueStoreCausale", kvCausale)
	if err != nil {
		fmt.Println("SERVER: Errore durante la registrazione di KeyValueStoreCausale", err)
		return
	}
	err = rpc.RegisterName("KeyValueStoreSequential", kvSequential)
	if err != nil {
		fmt.Println("SERVER: Errore durante la registrazione di KeyValueStoreSequential", err)
		return
	}

	// Ottieni la porta da una variabile d'ambiente o assegna un valore predefinito
	/* port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Porta predefinita se RPC_PORT non è impostata !!! lettura file di config?
	} */

	fmt.Println("Inserisci la porta:")
	var port string
	_, err = fmt.Scanln(&port)
	if err != nil {
		fmt.Println("SERVER: Errore durante la lettura dell'input:", err)
		return
	}

	// Avvio del listener RPC sulla porta specificata
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("SERVER: Errore nell'avvio del listener RPC:", err)
		return
	}

	// Accettazione delle connessioni in arrivo
	rpc.Accept(listener)
}
