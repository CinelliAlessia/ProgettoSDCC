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
//   - una richiesta di put x:1 al server1,
//   - una richiesta di get x put y:2 al server2 (così da essere in relazione di causa-effetto)
func basicTestCE(rpcName string) {

	fmt.Println("In questo basicTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put al server1\n" +
		"- una richiesta di get x put y:2 al server2 (causa-effetto)")

	done := make(chan bool)

	go func() {
		executeSpecificCall(0, rpcName+put, "x", "1")
		done <- true
	}()

	go func() {
		executeSpecificCall(1, rpcName+get, "x")
		executeSpecificCall(1, rpcName+put, "y", "2")
		done <- true
	}()

	// Attendi il completamento di tutte le goroutine (2)
	for i := 0; i < 2; i++ {
		<-done
	}
}

// In questo mediumTestCE vengono inviate in goroutine:
//   - una richiesta di put x:a e put y:b al server1,
//   - una richiesta di get x e put x:b al server2,
//   - una richiesta di get y e put y:a al server3,
func mediumTestCE(rpcName string) {

	fmt.Println("In mediumTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di put x:a e put y:b al server1\n" +
		"- una richiesta di get x e put x:b al server2\n" +
		"- una richiesta di get y e put y:a al server3")

	done := make(chan bool)

	// Invia richieste al primo server
	go func() {
		executeSpecificCall(0, rpcName+put, "x", "a")
		executeSpecificCall(0, rpcName+put, "y", "b")
		done <- true
	}()

	// Invia richieste al secondo server
	go func() {
		executeSpecificCall(1, rpcName+get, "x")
		executeSpecificCall(1, rpcName+put, "x", "b")
		done <- true
	}()

	// Invia richieste al terzo server
	go func() {
		executeSpecificCall(2, rpcName+get, "y")
		executeSpecificCall(2, rpcName+put, "y", "a")
		done <- true
	}()

	// Attendi il completamento di tutte le goroutine (3)
	for i := 0; i < 3; i++ {
		<-done
	}
}

// In questo complexTestCE vengono inviate in goroutine:
//   - una richiesta di get y, get y, se leggo y:c -> put x:b e get y al server1,
//   - una richiesta di put y:b, get x, get y, get x al server2,
//   - una richiesta di get x, se leggo x:b -> put x:c, put y:c e get x al server3,
func complexTestCE(rpcName string) {
	fmt.Println("In questo complexTestCE vengono inviate in goroutine:\n" +
		"- una richiesta di get y, get y se leggo y:c -> put x:b e get y al server1\n" +
		"- una richiesta di put y:b, get x, get y, get x al server2\n" +
		"- una richiesta di get x se leggo x:b -> put x:c, put y:c e get x al server3")

	done := make(chan bool)

	// Invio richieste al primo server
	go func() {
		executeSpecificCall(0, rpcName+get, "y")

		response := executeSpecificCall(0, rpcName+get, "y")
		if response.Value == "c" {
			executeSpecificCall(0, rpcName+put, "x", "b")
		}

		executeSpecificCall(0, rpcName+put, "x", "b")

		executeSpecificCall(0, rpcName+get, "x")
		done <- true // Segnala il completamento del test
	}()

	go func() {
		executeSpecificCall(1, rpcName+put, "y", "b")
		executeSpecificCall(1, rpcName+get, "x")
		executeSpecificCall(1, rpcName+get, "x")
		executeSpecificCall(1, rpcName+get, "x")

		done <- true // Segnala il completamento del test
	}()

	// Invio di richieste al terzo server
	go func() {
		executeSpecificCall(2, rpcName+get, "x")
		executeSpecificCall(2, rpcName+put, "x", "c")
		executeSpecificCall(2, rpcName+put, "y", "c")
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
// - put prova:1 al server1,
// - put prova:2 al server2,
// - put prova:3 al server3.
func basicTestSeq(rpcName string) {
	fmt.Println("In questo basicTestSeq vengono inviate in goroutine:\n" +
		"- put prova:1 al server1\n" +
		"- put prova:2 al server2\n" +
		"- put prova:3 al server3")

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

// mediumTestSeq contatta tutti i server in goroutine con operazioni di put
// - put x:1, put y:1, put z:1 al server1,
// - put x:2, put y:2, put z:2 al server2,
// - put x:3, put y:3, put z:3 al server3.
func mediumTestSeq(rpcName string) {
	fmt.Println("In questo mediumTestSeq vengono inviate in goroutine:\n" +
		"- put x:1, put y:1, put z:1 al server1\n" +
		"- put x:2, put y:2, put z:2 al server2\n" +
		"- put x:3, put y:3, put z:3 al server3")

	done := make(chan bool)

	// Definisci le azioni da eseguire per ciascuna goroutine
	actions := []func(){
		func() {
			executeSpecificCall(0, rpcName+put, "x", "1")
			executeSpecificCall(0, rpcName+put, "y", "1")
			executeSpecificCall(0, rpcName+put, "z", "1")
		},

		func() {
			executeSpecificCall(1, rpcName+put, "x", "2")
			executeSpecificCall(1, rpcName+put, "y", "2")
			executeSpecificCall(1, rpcName+put, "z", "2")
		},

		func() {
			executeSpecificCall(2, rpcName+put, "x", "3")
			executeSpecificCall(2, rpcName+put, "y", "3")
			executeSpecificCall(2, rpcName+put, "z", "3")
		},
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

// In questo complexTestSeq vengono inviate in goroutine:
//   - put x:1, put x:2, get x al server1,
//   - put x:3, get x, put x:4 al server2,
//   - put x:5, get x, put x:6 al server3.
func complexTestSeq(rpcName string) {
	fmt.Println("In questo complexTestSeq vengono inviate in goroutine:\n" +
		"- put x:1, put x:2, get x al server1\n" +
		"- put x:3, get x, put x:4 al server2\n" +
		"- put x:5, get x, put x:6 al server3")

	done := make(chan bool)

	// Definisci le azioni da eseguire per ciascuna goroutine
	actions := []func(){
		func() {
			executeSpecificCall(0, rpcName+put, "x", "1")
			executeSpecificCall(0, rpcName+put, "x", "2")
			executeSpecificCall(0, rpcName+get, "x")
		},
		func() {
			executeSpecificCall(1, rpcName+put, "x", "3")
			executeSpecificCall(1, rpcName+get, "x")
			executeSpecificCall(1, rpcName+put, "x", "4")
		},
		func() {
			executeSpecificCall(2, rpcName+get, "x")
			executeSpecificCall(2, rpcName+put, "x", "5")
			executeSpecificCall(2, rpcName+get, "x")
		},
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
