// client.go
package main

import (
	"fmt"
	"main/common"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"time"
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
		//choice := 2
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

	//args := common.Args{Key: common.GenerateUniqueID(), Value: "ciao"}
	args := common.Args{Key: common.GenerateUniqueID(), Value: "CIAO"}
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
	//TODO Problema Client: Se viene eliminato il timer non avviene la corretta esecuzione
	time.Sleep(time.Millisecond * 1000)

	//fmt.Print("\nContinuare con la get: ")
	//var choice int
	//_, err = fmt.Scan(&choice)

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

	//args := common.Args{Key: common.GenerateUniqueID(), Value: "ciao"}
	args := common.Args{Key: common.GenerateUniqueID(), Value: "CIAO"}
	reply := common.Response{}

	// PUT
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}

	err := conn.Call("KeyValueStoreCausale.Put", args, &reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Put:", err)
		return
	}
	fmt.Println("CLIENT: Richiesta put effettuata " + reply.Reply)

	time.Sleep(time.Millisecond * 1000)
	//fmt.Print("\nContinuare con la get: ")
	//var choice int
	//_, err = fmt.Scan(&choice)

	// GET
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione 2")
		return
	}
	// Effettua la chiamata RPC
	err = conn.Call("KeyValueStoreCausale.Get", args, &reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Get:", err)
		return
	}

	fmt.Println("CLIENT: Richiesta get effettuata " + reply.Reply)
	// comandi random a server random, li conosce tutti e fine
}

func randomConnect() *rpc.Client {
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(common.Replicas)

	var conn *rpc.Client
	var err error

	if os.Getenv("CONFIG") == "1" {
		/*---LOCALE---*/
		// Connessione al server RPC casuale
		fmt.Println("CLIENT: Contatto il server:", ":"+common.ReplicaPorts[randomIndex])
		conn, err = rpc.Dial("tcp", ":"+common.ReplicaPorts[randomIndex])

	} else if os.Getenv("CONFIG") == "2" {
		/*---DOCKER---*/
		// Connessione al server RPC casuale
		fmt.Println("CLIENT: Contatto il server:", "server"+strconv.Itoa(randomIndex+1)+":"+common.ReplicaPorts[randomIndex])
		conn, err = rpc.Dial("tcp", "server"+strconv.Itoa(randomIndex+1)+":"+common.ReplicaPorts[randomIndex])
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA")
	}
	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

func callPut(conn *rpc.Client, args common.Args, reply common.Response) error {
	err := conn.Call("KeyValueStoreSequential.Put", args, &reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	fmt.Println("CLIENT: Richiesta put effettuata: " + reply.Reply)
	return nil
}
func callGet(conn *rpc.Client, args common.Args, reply common.Response) error {
	err := conn.Call("KeyValueStoreSequential.Get", args, &reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	fmt.Println("CLIENT: Richiesta get effettuata: " + reply.Reply)
	return nil
}
func callDelete(conn *rpc.Client, args common.Args, reply common.Response) error {
	err := conn.Call("KeyValueStoreSequential.Delete", args, &reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	fmt.Println("CLIENT: Richiesta delete effettuata: " + reply.Reply)
	return nil
}

//time.Sleep(time.Millisecond * 1000)
//fmt.Print("\nContinuare con la get: ")
//var choice int
//_, err = fmt.Scan(&choice)
