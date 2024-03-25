package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
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

		// Esegui l'operazione scelta
		switch choice {
		case 1:
			causal()
			break
		case 2:
			sequential()
			break
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}
	}
}

func sequential() {
	// Test consistenza sequenziale
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(Replicas)

	// Connessione al server RPC casuale
	conn, err := rpc.Dial("tcp", ":"+ReplicaPorts[randomIndex])
	if err != nil {
		fmt.Println("Errore durante la connessione al server:", err)
		return
	}

	args := Args{generateUniqueID(), "1234567890"}
	reply := Response{}

	// Effettua la chiamata RPC
	err = conn.Call("KeyValueStoreSequential.Put", args, &reply)
	if err != nil {
		fmt.Println("Errore durante la chiamata RPC:", err)
		return
	}

	// comandi random a server random, li conosce tutti e fine
}

func causal() {
	// Test consistenza causale
}
