package main

import (
	"fmt"
	"main/common"
	_ "sync"
	"time"
)

// Inizializza lo stato del client
var clientState *ClientState

const (
	synchronous = "synchronous"
	async       = "async"

	random   = "random"
	specific = "specific"
)

type Operation struct {
	ServerIndex   int
	OperationType string
	Key           string
	Value         string
}

func main() {

	clientState = NewClientState()

	for {
		// Stampa il menu interattivo
		fmt.Println("Scegli un'operazione:")
		fmt.Println("1. Consistenza causale")
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
			rpcName = common.Causal
			causal(rpcName)
			break
		case 2:
			rpcName = common.Sequential
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

		choice, done := chooseFuncTest()
		if done {
			return
		}

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
			clientState = NewClientState()
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
		choice, done := chooseFuncTest()
		if done {
			return
		}

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
			clientState = NewClientState()
			fmt.Println()
			return
		}

		time.Sleep(5000 * time.Millisecond) // Aggiungo un ritardo per evitare che le stampe si sovrappongano
	}
}

func chooseFuncTest() (int, bool) {
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
		return 0, true
	}
	fmt.Println()
	return choice, false
}
