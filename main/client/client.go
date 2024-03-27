// client.go
package main

import (
	"fmt"
	"main/common"
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
			fmt.Println("Scelta di consistenza causale")
			causal()
			break
		case 2:
			fmt.Println("Scelta di consistenza sequenziale")
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
	randomIndex := rand.Intn(common.Replicas)
	// Connessione al server RPC casuale
	fmt.Println("CLIENT: Contatto il server " + common.ReplicaPorts[randomIndex])
	conn, err := rpc.Dial("tcp", ":"+common.ReplicaPorts[randomIndex])
	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return
	}

	args := common.Args{Key: common.GenerateUniqueID(), Value: "1234567890"}
	reply := common.Response{}

	// Effettua la chiamata RPC
	err = conn.Call("KeyValueStoreSequential.Put", args, &reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC:", err)
		return
	}

	fmt.Println("CLIENT: Richiesta effettuata")
	// comandi random a server random, li conosce tutti e fine
}

func causal() {
	// Test consistenza causale
}
