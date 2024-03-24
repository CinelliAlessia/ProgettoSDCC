package main

import "fmt"

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
		case 2:
			sequential()
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}
	}
}

func sequential() {
	// Test consistenza sequenziale

	// comandi random a server random, li conosce tutti e fine
}

func causal() {
	// Test consistenza causale
}
