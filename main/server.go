package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
)

// server.go
// Questo è il codice dei server replica, che vengono contattati dai client tramite una procedura rpc
// e rispondono in maniera adeguata.

func main() {
	// Inizializzazione delle strutture KeyValueStoreCausale e KeyValueStoreSequential
	kvCausale := &KeyValueStoreCausale{
		data:        make(map[string]string),
		vectorClock: make([]int, 0), // Inizializzazione dell'orologio vettoriale
	}

	kvSequential := &KeyValueStoreSequential{
		dataStore:    make(map[string]string),
		logicalClock: 0, // Inizializzazione dell'orologio logico scalare
	}

	// Registrazione dei servizi RPC
	err := rpc.RegisterName("KeyValueStoreCausale", kvCausale)
	if err != nil {
		return
	}
	err = rpc.RegisterName("KeyValueStoreSequential", kvSequential)
	if err != nil {
		return
	}

	// Ottieni la porta da una variabile d'ambiente o assegna un valore predefinito
	port := os.Getenv("RPC_PORT")
	if port == "" {
		port = "8080" // Porta predefinita se RPC_PORT non è impostata !!! lettura file di config?
	}

	// Avvio del listener RPC sulla porta specificata
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Errore nell'avvio del listener RPC:", err)
		return
	}

	// Accettazione delle connessioni in arrivo
	rpc.Accept(listener)
}
