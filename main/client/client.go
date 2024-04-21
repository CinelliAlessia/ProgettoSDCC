// client.go
package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"math/rand"
	"net/rpc"
	"strings"
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
			fmt.Println("Errore durante la lettura dell'input:", err)
			break
		}

		var rpcName string
		switch choice {
		case 1:
			rpcName = "KeyValueStoreCausale"
			causal(rpcName)
			break
		case 2:
			rpcName = "KeyValueStoreSequential"
			sequential(rpcName)
			break
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}
	}
}

// sequential() Scegliere il tipo di test che si vuole eseguire per verificare le garanzie di consistenza sequenziale
func sequential(rpcName string) {
	// Stampa il menu interattivo
	fmt.Println("\nConsistenza sequenziale, scegli il test da eseguire: ")
	fmt.Println("1. Basic Test")
	fmt.Println("2. Medium Test")
	fmt.Println("3. Complex Test")

	// Leggi l'input dell'utente per l'operazione
	fmt.Print("Inserisci il numero dell'operazione desiderata: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Errore durante la lettura dell'input:", err)
		return
	}
	fmt.Println()

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
	}
	fmt.Println()

}

// causal() Scegliere il tipo di test che si vuole eseguire per verificare le garanzie di consistenza causale
func causal(rpcName string) {
	// Stampa il menu interattivo
	fmt.Println("\nConsistenza causale, scegli il test da eseguire: ")
	fmt.Println("1. Basic Test")
	fmt.Println("2. Medium Test")
	fmt.Println("3. Complex Test")

	// Leggi l'input dell'utente per l'operazione
	fmt.Print("\nInserisci il numero dell'operazione desiderata: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Errore durante la lettura dell'input:", err)
		return
	}
	fmt.Println()

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
	}

	fmt.Println()
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

func executeRandomCall(rpcName, key string, values ...string) common.Response {
	var response common.Response

	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return common.Response{}
	}

	if len(values) > 0 {
		response = executeCall(conn, rpcName, key, values[0])
	} else {
		response = executeCall(conn, rpcName, key)
	}
	return response
}

// executeSpecificCall:
//   - index rappresenta l'indice relativo al server da contattare, da 0 a (common.Replicas-1)
func executeSpecificCall(index int, rpcName, key string, values ...string) common.Response {
	var response common.Response

	conn := specificConnect(index)
	if conn == nil {
		fmt.Println("executeSpecificCall: Errore durante la connessione")
		return common.Response{}
	}

	if len(values) > 0 {
		response = executeCall(conn, rpcName, key, values[0])
	} else {
		response = executeCall(conn, rpcName, key)
	}
	return response
}

// executeCall esegue un comando ad un server random. Il comando da eseguire viene specificato tramite i parametri inseriti
func executeCall(conn *rpc.Client, rpcName, key string, values ...string) common.Response {
	var value string
	if len(values) > 0 {
		value = values[0]
	}

	args := common.Args{Key: key, Value: value}
	response := common.Response{}

	err := delayedCall(conn, args, &response, rpcName)
	if err != nil {
		return common.Response{}
	}
	return response
}

// delayedCall esegue una chiamata RPC ritardata utilizzando il client RPC fornito.
// Prima di effettuare la chiamata, applica un ritardo casuale per simulare condizioni reali di rete.
func delayedCall(conn *rpc.Client, args common.Args, response *common.Response, rpcName string) error {

	// Applica un ritardo casuale
	//common.RandomDelay() // Vogliamo applicare o meno il ritardo? -> Scelto all'utilizzo e si imposta una variabile globale

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
