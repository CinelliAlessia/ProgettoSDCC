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
			//break
		}

		var rpcName string
		//choice := 2
		switch choice {
		case 1:
			fmt.Println("Scelta di consistenza causale")
			rpcName = "KeyValueStoreCausale"
			//causal(args, &reply)
			break
		case 2:
			fmt.Println("Scelta di consistenza sequenziale")
			rpcName = "KeyValueStoreSequential"
			//sequential(args, &reply)
			break
		default:
			fmt.Println("Scelta non valida. Riprova.")
		}

		go run1(rpcName)
		go run2(rpcName)
		go run3(rpcName)
		go run4(rpcName)
	}
}

func sequential(args common.Args, response *common.Response) {
	// Test consistenza sequenziale

	// PUT
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}

	err := conn.Call("KeyValueStoreSequential.Put", args, &response.Reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Put:", err)
		return
	}
	fmt.Println("CLIENT: Richiesta put effettuata " + response.Reply)

	//TODO Problema Client: Se viene eliminato il timer non avviene la corretta esecuzione
	time.Sleep(time.Millisecond * 1000)

	// GET
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione 2")
		return
	}

	// Effettua la chiamata RPC
	err = conn.Call("KeyValueStoreSequential.Get", args, &response.Reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Get:", err)
		return
	}

	fmt.Println("CLIENT: Richiesta get effettuata " + response.Reply)
	// comandi random a server random, li conosce tutti e fine
}
func causal(args common.Args, response *common.Response) {
	// Test consistenza causale

	// PUT
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}

	err := conn.Call("KeyValueStoreCausale.Put", args, &response.Reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Put:", err)
		return
	}
	fmt.Println("CLIENT: Richiesta put effettuata " + response.Reply)

	time.Sleep(time.Millisecond * 1000)

	// GET
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione 2")
		return
	}
	// Effettua la chiamata RPC
	err = conn.Call("KeyValueStoreCausale.Get", args, &response.Reply)
	if err != nil {
		fmt.Println("CLIENT: Errore durante la chiamata RPC Get:", err)
		return
	}

	fmt.Println("CLIENT: Richiesta get effettuata " + response.Reply)
	// comandi random a server random, li conosce tutti e fine
}

func randomConnect() *rpc.Client {
	// Genera un numero casuale tra 0 e il numero di repliche - 1
	randomIndex := rand.Intn(common.Replicas)

	var serverName string

	if os.Getenv("CONFIG") == "1" {
		/*---LOCALE---*/
		serverName = ":" + common.ReplicaPorts[randomIndex]
	} else if os.Getenv("CONFIG") == "2" {
		/*---DOCKER---*/
		serverName = "server" + strconv.Itoa(randomIndex+1) + ":" + common.ReplicaPorts[randomIndex]
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA")
		return nil
	}

	//fmt.Println("CLIENT: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}
func specificConnect(index int) *rpc.Client {
	if index >= common.Replicas {
		return nil
	}
	var serverName string

	if os.Getenv("CONFIG") == "1" {
		/*---LOCALE---*/
		serverName = ":" + common.ReplicaPorts[index]
	} else if os.Getenv("CONFIG") == "2" {
		/*---DOCKER---*/
		serverName = "server" + strconv.Itoa(index+1) + ":" + common.ReplicaPorts[index]
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA")
		return nil
	}

	fmt.Println("CLIENT: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)

	if err != nil {
		fmt.Println("CLIENT: Errore durante la connessione al server:", err)
		return nil
	}
	return conn
}

func callPut(conn *rpc.Client, args common.Args, reply *common.Response, rpcName string) error {
	err := conn.Call(rpcName+".Put", args, reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	//fmt.Println("CLIENT: Richiesta put effettuata: " + reply.Reply)
	return nil
}
func callGet(conn *rpc.Client, args common.Args, reply *common.Response, rpcName string) error {
	err := conn.Call(rpcName+".Get", args, reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	//fmt.Println("CLIENT: Richiesta get effettuata: " + reply.Reply)
	return nil
}
func callDelete(conn *rpc.Client, args common.Args, reply *common.Response, rpcName string) error {
	err := conn.Call(rpcName+".Delete", args, reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	//fmt.Println("CLIENT: Richiesta delete effettuata: " + reply.Reply)
	return nil
}

//time.Sleep(time.Millisecond * 1000)
//fmt.Print("\nContinuare con la get: ")
//var choice int
//_, err = fmt.Scan(&choice)

func run1(rpcName string) {

	// Esegui l'operazione scelta

	args := common.Args{Key: "y", Value: "0"}
	reply := common.Response{}
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err := callPut(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run1 put y:0", reply.Reply)

	args = common.Args{Key: "x", Value: "1"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callPut(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run1 put x:1", reply.Reply)

	args = common.Args{Key: "x"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run1 get x:", reply.Reply)

	args = common.Args{Key: "y"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run1 get y:", reply.Reply)
}

func run2(rpcName string) {

	// Esegui l'operazione scelta
	//args := common.Args{Key: common.GenerateUniqueID(), Value: "ciao"}
	args := common.Args{Key: "y", Value: "1"}
	reply := common.Response{}
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err := callPut(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run2 put y:1", reply.Reply)

	args = common.Args{Key: "y"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run2 get y:", reply.Reply)

	args = common.Args{Key: "x"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run2 get x:", reply.Reply)
}

func run3(rpcName string) {

	args := common.Args{Key: "x"}
	reply := common.Response{}
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err := callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run3 get x:", reply.Reply)

	args = common.Args{Key: "y"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run3 get y:", reply.Reply)
}

func run4(rpcName string) {

	// Esegui l'operazione scelta
	//args := common.Args{Key: common.GenerateUniqueID(), Value: "ciao"}
	args := common.Args{Key: "x", Value: "0"}
	reply := common.Response{}
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err := callPut(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run4 put x:0", reply.Reply)

	args = common.Args{Key: "y"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run4 get y:", reply.Reply)

	args = common.Args{Key: "x"}
	reply = common.Response{}
	conn = randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err = callGet(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run4 get x:", reply.Reply)
}
