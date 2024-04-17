package main

import (
	"fmt"
	"time"
)

/* ----- FUNZIONI DI TEST ----- */

func test1(rpcName string) {

	// Esecuzione delle rpc
	done := make(chan bool)
	cycles := 25

	for i := 0; i < cycles; i++ {
		go func() {
			executeRandomCall(rpcName+put, "y", "0")
			executeRandomCall(rpcName+put, "x", "1")
			executeRandomCall(rpcName+get, "x")
			executeRandomCall(rpcName+get, "y")
			done <- true
		}()
		go func() {
			executeRandomCall(rpcName+put, "y", "1")
			executeRandomCall(rpcName+get, "y")
			executeRandomCall(rpcName+get, "x")

			done <- true
		}()
		go func() {
			executeRandomCall(rpcName+put, "x", "9")
			executeRandomCall(rpcName+get, "y")
			done <- true
		}()
		go func() {
			executeRandomCall(rpcName+put, "x", "0")
			executeRandomCall(rpcName+put, "y", "3")
			executeRandomCall(rpcName+get, "x")

			done <- true
		}()
	}

	// Attendi il completamento di tutte le goroutine
	for i := 0; i < cycles*4; i++ {
		<-done
	}
}

func test2(rpcName string) {
	// Esecuzione delle rpc
	done := make(chan bool)
	cycles := 2

	fmt.Println("\nOrdinamento goroutine (senza ritardi di rete) ripetute", cycles, "volte"+
		"\nClient1: \nPut-giorno:18 Put-mese:02 "+
		"\nClient2: \nPut-giorno:16 Put-mese:02 "+
		"\nClient3: \nGet-giorno Get-mese Put-giorno:20 Put-mese:07")

	for i := 0; i < cycles; i++ {
		go func() {
			executeRandomCall(rpcName+put, "giorno", "18")
			executeRandomCall(rpcName+put, "mese", "02")
			done <- true
		}()
		go func() {
			executeRandomCall(rpcName+put, "giorno", "16")
			executeRandomCall(rpcName+put, "mese", "09")
			done <- true
		}()
		go func() {
			executeRandomCall(rpcName+get, "giorno")
			executeRandomCall(rpcName+get, "mese")
			executeRandomCall(rpcName+put, "giorno", "20")
			executeRandomCall(rpcName+put, "mese", "07")
			done <- true
		}()

		time.Sleep(time.Millisecond * 100)
	}

	// Attendi il completamento di tutte le goroutine
	for i := 0; i < cycles*3; i++ {
		<-done
	}
}

/* ----- CONSISTENZA CAUSALE ----- */

// In questo basicTestCE vengono inviate in goroutine:
//   - una richiesta di put al server1,
//   - al server due le richieste di get e successivamente una put (così da essere in relazione di causa-effetto) sul singolo server
func basicTestCE(rpcName string) {
	done := make(chan bool)

	go func() {
		executeSpecificCall(0, rpcName+put, "prova1", "1")
		done <- true
	}()

	time.Sleep(time.Millisecond * 100)

	go func() {
		executeSpecificCall(1, rpcName+get, "prova1")
		executeSpecificCall(1, rpcName+put, "prova2", "2")
		done <- true
	}()

	// Attendi il completamento di tutte le goroutine (2)
	for i := 0; i < 2; i++ {
		<-done
	}
}

func mediumTestCE(rpcName string) {
	done := make(chan bool)

	// Invia richieste al primo server
	go func() {
		executeSpecificCall(0, rpcName+put, "x", "a")
		executeSpecificCall(0, rpcName+put, "x", "b")
		done <- true
	}()

	// Invia richieste al secondo server
	go func() {
		executeSpecificCall(1, rpcName+get, "x")
		executeSpecificCall(1, rpcName+put, "x", "c")
		done <- true
	}()

	// Invia richieste al terzo server
	go func() {
		executeSpecificCall(2, rpcName+get, "x")
		executeSpecificCall(2, rpcName+put, "x", "d")
		done <- true
	}()

	// Attendi il completamento di tutte le goroutine (3)
	for i := 0; i < 3; i++ {
		<-done
	}
}

func complexTestCE(rpcName string) {
	done := make(chan bool)

	// Invio richieste al primo server
	go func() {
		executeSpecificCall(0, rpcName+put, "x", "a")
		executeSpecificCall(0, rpcName+put, "x", "b")
		response := executeSpecificCall(0, rpcName+get, "x")
		if response.Value == "c" {
			executeSpecificCall(0, rpcName+put, "x", "d")
		}
		done <- true // Segnala il completamento del test
	}()

	go func() {
		response := executeSpecificCall(1, rpcName+get, "x")
		if response.Value == "a" {
			executeSpecificCall(1, rpcName+put, "x", "c")
		}
		executeSpecificCall(1, rpcName+put, "x", "d")
		done <- true // Segnala il completamento del test
	}()

	// Invio di richieste al terzo server
	go func() {
		executeSpecificCall(2, rpcName+put, "x", "a")
		executeSpecificCall(2, rpcName+get, "x")
		executeSpecificCall(2, rpcName+get, "x")
		done <- true // Segnala il completamento del test
	}()

	// Attendi il completamento di tutte le goroutine (3)
	for i := 0; i < 3; i++ {
		<-done
	}
}

/* ----- CONSISTENZA SEQUENZIALE ----- */

// basicTestSeq contatta tutti i server in goroutine con operazioni di put
// un corretto funzionamento della consistenza sequenziale prevede che a prescindere dall'ordinamento
// tutti i server eseguiranno nello stesso ordine le richieste.
func basicTestSeq(rpcName string) {
	done := make(chan bool)

	go func() {
		executeSpecificCall(0, rpcName+put, "prova", "1")
		done <- true
	}()
	go func() {
		executeSpecificCall(1, rpcName+put, "prova", "2")
		done <- true
	}()
	go func() {
		executeSpecificCall(2, rpcName+put, "prova", "3")
		done <- true
	}()

	// Attendi il completamento di tutte le goroutine (3)
	for i := 0; i < 3; i++ {
		<-done
	}
}

func mediumTestSeq(rpcName string) {
	done := make(chan bool)

	// Definisci le azioni da eseguire per ciascuna goroutine
	actions := []func(){
		func() { executeSpecificCall(0, rpcName+put, "prova", "1") },
		func() { executeSpecificCall(1, rpcName+put, "prova", "2") },
		func() { executeSpecificCall(2, rpcName+put, "prova", "3") },
		func() { executeSpecificCall(0, rpcName+put, "prova1", "1") },
		func() { executeSpecificCall(1, rpcName+put, "prova1", "2") },
		func() { executeSpecificCall(2, rpcName+put, "prova1", "3") },
		func() { executeSpecificCall(0, rpcName+put, "prova2", "1") },
		func() { executeSpecificCall(1, rpcName+put, "prova2", "2") },
		func() { executeSpecificCall(2, rpcName+put, "prova2", "3") },
	}

	// Avvia le goroutine per ciascuna azione
	for _, action := range actions {
		go func(act func()) {
			act()
			done <- true
		}(action)
	}

	// Attendi il completamento di tutte le goroutine
	for i := 0; i < len(actions); i++ {
		<-done
	}
}

func complexTestSeq(rpcName string) {

}
