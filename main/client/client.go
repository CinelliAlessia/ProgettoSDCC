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

	args := common.Args{Key: common.GenerateUniqueID(), Value: "ciao"}
	reply := common.Response{}

	// PUT
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}

	err := conn.Call("KeyValueStoreSequential.Put", args, &reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Put:", err)
		return
	}

	fmt.Println("CLIENT: Richiesta put effettuata " + reply.Reply)

	// GET
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione 2")
		return
	}
	// Effettua la chiamata RPC
	err = conn.Call("KeyValueStoreSequential.Get", args, &reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Get:", err)
		return
	}

	fmt.Println("CLIENT: Richiesta get effettuata " + reply.Reply)
	// comandi random a server random, li conosce tutti e fine
}

func causal() {
	// Test consistenza causale
}

func randomConnect() *rpc.Client {
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(common.Replicas)

	// Connessione al server RPC casuale
	fmt.Println("CLIENT: Contatto il server " + common.ReplicaPorts[randomIndex])
	conn, err := rpc.Dial("tcp", ":"+common.ReplicaPorts[randomIndex])
	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}
