// client.go
package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"math/rand"
	"net/rpc"
	"strings"
	"time"
)

const (
	put = ".Put"
	get = ".Get"
	del = ".Delete"
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

		var rpcName string
		switch choice {
		case 1:
			fmt.Println("Scelta di consistenza causale")
			rpcName = "KeyValueStoreCausale"
			break
		case 2:
			fmt.Println("Scelta di consistenza sequenziale")
			rpcName = "KeyValueStoreSequential"
			break
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}

		/* --- Scegliere il tipo di test che si vuole eseguire --- */
		//test1(rpcName)
		//test2(rpcName)
		//basicTestCE(rpcName)
		//basicTestSeq(rpcName)
		mediumTestSeq(rpcName)
	}
}

/* ----- FUNZIONI UTILIZZATE PER LA CONNESSIONE -----*/

// randomConnect restituisce una connessione random con un server definito in common/config.go
func randomConnect() *rpc.Client {
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(common.Replicas)

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[randomIndex], randomIndex)

	//fmt.Println("CLIENT: Contatto il server", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

// specificConnect restituisce una connessione specifica con un server definito in common/config.go, tramite index passato in argomento
// Attenzione: Indicizzazione parte da 0!
func specificConnect(index int) *rpc.Client {
	if index > common.Replicas {
		_ = fmt.Errorf("index out of range")
		return nil
	}

	// Ottengo l'indirizzo a cui connettermi
	serverName := common.GetServerName(common.ReplicaPorts[index], index)

	//fmt.Println("CLIENT: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

func executeRandomCall(rpcName, key string, values ...string) {
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}

	if len(values) > 0 {
		executeCall(conn, rpcName, key, values[0])
	} else {
		executeCall(conn, rpcName, key)
	}
}

// executeSpecificCall:
//   - index rappresenta l'indice relativo al server da contattare, da 0 a (common.Replicas-1)
func executeSpecificCall(index int, rpcName, key string, values ...string) {
	conn := specificConnect(index)
	if conn == nil {
		fmt.Println("executeSpecificCall: Errore durante la connessione")
		return
	}

	if len(values) > 0 {
		executeCall(conn, rpcName, key, values[0])
	} else {
		executeCall(conn, rpcName, key)
	}
}

// executeCall esegue un comando ad un server random. Il comando da eseguire viene specificato tramite i parametri inseriti
func executeCall(conn *rpc.Client, rpcName, key string, values ...string) {
	var value string
	if len(values) > 0 {
		value = values[0]
	}

	args := common.Args{Key: key, Value: value}
	response := common.Response{}

	// TODO: Qui posso usare un id auto-incrementativo per un DEBUG accurato
	err := delayedCall(conn, args, &response, rpcName)
	if err != nil {
		return
	}
}

// delayedCall esegue una chiamata RPC ritardata utilizzando il client RPC fornito.
// Prima di effettuare la chiamata, applica un ritardo casuale per simulare condizioni reali di rete.
func delayedCall(conn *rpc.Client, args common.Args, response *common.Response, rpcName string) error {

	// Applica un ritardo casuale
	common.RandomDelay() // Vogliamo applicare o meno il ritardo? -> Scelto all'utilizzo e si imposta una variabile globale

	debugPrintRun(rpcName, args)

	// Effettua la chiamata RPC
	err := conn.Call(rpcName, args, response)
	if err != nil {
		return fmt.Errorf("errore durante la chiamata RPC in client.call: %s", err)
	}

	debugPrintResponse(rpcName, args, *response)

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("errore durante la chiusura della connessione in client.call: %s", err)
	}
	return nil
}

/* ----- FUNZIONI PER PRINT DI DEBUG ----- */

func debugPrintRun(rpcName string, args common.Args) {
	if common.GetDebug() {
		debugName := strings.SplitAfter(rpcName, ".")
		name := "." + debugName[1]

		switch name {
		case put:
			fmt.Println(color.BlueString("RUN Put"), args.Key+":"+args.Value)
		case get:
			fmt.Println(color.BlueString("RUN Get"), args.Key)
		case del:
			fmt.Println(color.BlueString("RUN Delete"), args.Key)
		default:
			fmt.Println(color.BlueString("RUN Unknown"), rpcName, args)
		}
	} else {
		return
	}
}

func debugPrintResponse(rpcName string, args common.Args, response common.Response) {

	debugName := strings.SplitAfter(rpcName, ".")
	name := "." + debugName[1]

	switch name {
	case put:
		fmt.Println(color.GreenString("RISPOSTA Put"), "key:"+args.Key, "value:"+args.Value, "result:", response.Result)
	case get:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Get"), "key:"+args.Key, "response:"+response.Value)
		} else {
			fmt.Println(color.RedString("RISPOSTA Get fallita"), "key:"+args.Key)
		}
	case del:
		fmt.Println(color.GreenString("RISPOSTA Delete"), "key:"+args.Key, "result:", response.Result)
	default:
		fmt.Println(color.GreenString("RISPOSTA Unknown"), rpcName, args, response)
	}
}

/* ----- FUNZIONI DI TEST ----- */

func test1(rpcName string) {

	// Esecuzione delle rpc
	done := make(chan bool)
	cycles := 25

	for i := 0; i < cycles; i++ {
		go run1(rpcName, done)
		go run2(rpcName, done)
		go run3(rpcName, done)
		go run4(rpcName, done)
	}

	// Attendi il completamento di tutte le goroutine
	for i := 0; i < cycles*4; i++ {
		<-done
	}
}

func run1(rpcName string, done chan bool) {

	executeRandomCall(rpcName+put, "y", "0")
	executeRandomCall(rpcName+put, "x", "1")
	executeRandomCall(rpcName+get, "x")
	executeRandomCall(rpcName+get, "y")

	done <- true
}

func run2(rpcName string, done chan bool) {

	executeRandomCall(rpcName+put, "y", "1")
	executeRandomCall(rpcName+get, "y")
	executeRandomCall(rpcName+get, "x")

	done <- true
}

func run3(rpcName string, done chan bool) {
	executeRandomCall(rpcName+put, "x", "9")
	executeRandomCall(rpcName+get, "y")
	done <- true
}

func run4(rpcName string, done chan bool) {
	executeRandomCall(rpcName+put, "x", "0")
	executeRandomCall(rpcName+put, "y", "3")
	executeRandomCall(rpcName+get, "x")

	done <- true
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
		go client1(rpcName, done)
		go client2(rpcName, done)
		go client3(rpcName, done)

		time.Sleep(time.Millisecond * 100)
	}

	// Attendi il completamento di tutte le goroutine
	for i := 0; i < cycles*3; i++ {
		<-done
	}
}

func client1(rpcName string, done chan bool) {
	executeRandomCall(rpcName+put, "giorno", "18")
	executeRandomCall(rpcName+put, "mese", "02")
	done <- true
}

func client2(rpcName string, done chan bool) {
	executeRandomCall(rpcName+put, "giorno", "16")
	executeRandomCall(rpcName+put, "mese", "09")
	done <- true
}

func client3(rpcName string, done chan bool) {
	executeRandomCall(rpcName+get, "giorno")
	executeRandomCall(rpcName+get, "mese")
	executeRandomCall(rpcName+put, "giorno", "20")
	executeRandomCall(rpcName+put, "mese", "07")
	done <- true
}

// In questo basicTestCE vengono inviate in goroutine:
//   - una richiesta di put al server1,
//   - al server due le richieste di get e successivamente una put (così da essere in relazione di causa-effetto) sul singolo server
func basicTestCE(rpcName string) {
	go executeSpecificCall(0, rpcName+put, "prova1", "1")

	time.Sleep(time.Millisecond * 100)

	go func() {
		executeSpecificCall(1, rpcName+get, "prova1")
		executeSpecificCall(1, rpcName+put, "prova2", "2")
	}()
}

// basicTestSeq contatta tutti i server in goroutine con operazioni di put
// un corretto funzionamento della consistenza sequenziale prevede che a prescindere dall'ordinamento
// tutti i server eseguiranno nello stesso ordine le richieste.
func basicTestSeq(rpcName string) {
	go executeSpecificCall(0, rpcName+put, "prova", "1")
	go executeSpecificCall(1, rpcName+put, "prova", "2")
	go executeSpecificCall(2, rpcName+put, "prova", "3")
}

func mediumTestSeq(rpcName string) {
	go executeSpecificCall(0, rpcName+put, "prova", "1")
	go executeSpecificCall(1, rpcName+put, "prova", "2")
	go executeSpecificCall(2, rpcName+put, "prova", "3")

	go executeSpecificCall(0, rpcName+put, "prova1", "1")
	go executeSpecificCall(1, rpcName+put, "prova1", "2")
	go executeSpecificCall(2, rpcName+put, "prova1", "3")

	go executeSpecificCall(0, rpcName+put, "prova2", "1")
	go executeSpecificCall(1, rpcName+put, "prova2", "2")
	go executeSpecificCall(2, rpcName+put, "prova2", "3")
}
