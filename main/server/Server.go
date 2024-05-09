// Server.go
package main

import (
	"fmt"
	"main/common"
	"main/server/algorithms/causal"
	"main/server/algorithms/sequential"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

// Questo è il codice dei server replica, che vengono contattati dai client tramite una procedura rpc.
func main() {

	var idStr string

	// CONFIG è una variabile d'ambiente utilizzata per identificare se il programma verrà eseguito in locale oppure su docker.
	if os.Getenv("CONFIG") == "1" { /*---LOCALE---*/

		if len(os.Args) < 2 { // Controllo se è stato passato per argomento l'id del server
			fmt.Println("Usare: ", os.Args[0], "<ID_Server>")
			os.Exit(1)
		}

		// Legge l'ID del server passato come argomento dalla riga di comando
		idStr = os.Args[1]
	} else { /*---DOCKER---*/

		// Ottieni la porta da una variabile d'ambiente o assegna un valore predefinito
		idStr = os.Getenv("SERVER_ID")
	}

	// Converti l'ID del server in un intero per calcolare il numero di porta su cui mettersi in ascolto
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Errore:", err)
		os.Exit(1)
	}
	port := common.ReplicaPorts[id]

	// Inizializzazione delle strutture

	// ----- CONSISTENZA CAUSALE -----
	kvCausale := causal.NewKeyValueStoreCausal(id)

	// ----- CONSISTENZA SEQUENZIALE -----
	kvSequential := sequential.NewKeyValueStoreSequential(id)

	// ----- REGISTRAZIONE DEI SERVIZI RPC -----
	// Registrazione dei servizi RPC
	err = rpc.RegisterName(common.Causal, kvCausale)
	if err != nil {
		fmt.Println("SERVER: Errore durante la registrazione di KeyValueStoreCausale", err)
		return
	}
	err = rpc.RegisterName(common.Sequential, kvSequential)
	if err != nil {
		fmt.Println("SERVER: Errore durante la registrazione di KeyValueStoreSequential", err)
		return
	}

	// Avvio del listener RPC sulla porta specificata
	fmt.Println("Server:", id+1)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("SERVER: Errore nell'avvio del listener RPC:", err)
		return
	}

	// Ciclo per accettare e gestire le connessioni in arrivo in maniera asincrona
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("SERVER: Errore nell'accettare la connessione dal client:", err)
			continue
		}

		// Avvia la gestione della connessione in un goroutine
		go func(conn net.Conn) {
			// Servi la connessione RPC
			rpc.ServeConn(conn)

			defer func() {
				err := conn.Close()
				if err != nil {
				}
			}()
		}(conn)
	}
}
