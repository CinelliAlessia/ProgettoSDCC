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

const (
	put    = ".Put"
	get    = ".Get"
	delete = ".Delete"
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

		done := make(chan bool)
		go run1(rpcName, done)
		go run2(rpcName, done)
		go run3(rpcName, done)
		go run4(rpcName, done)
		time.Sleep(time.Millisecond * 100)
		go run1(rpcName, done)
		go run2(rpcName, done)
		go run3(rpcName, done)
		go run4(rpcName, done)

		// Attendi il completamento di tutte le goroutine
		for i := 0; i < 8; i++ {
			println("OK", i)
			<-done
		}
	}
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

	fmt.Println("CLIENT: Contatto il server:", serverName)
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

func randomDelay() {
	// Genera un numero casuale compreso tra 0 e 999 (max un secondo)
	delay := rand.Intn(1000)

	// Introduce un ritardo casuale
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func executeCall(rpcName, key string, values ...string) {
	var value string
	if len(values) > 0 {
		value = values[0]
	}

	args := common.Args{Key: key, Value: value}
	reply := common.Response{}
	conn := randomConnect()
	if conn == nil {
		fmt.Println("CLIENT: Errore durante la connessione")
		return
	}
	err := call(conn, args, &reply, rpcName)
	if err != nil {
		return
	}
	fmt.Println("Run", rpcName, key+":"+value, reply.Reply)
}

func call(conn *rpc.Client, args common.Args, reply *common.Response, rpcName string) error {
	err := conn.Call(rpcName, args, reply)
	if err != nil {
		return fmt.Errorf("CLIENT: Errore durante la chiamata RPC")
	}
	return nil
}

//time.Sleep(time.Millisecond * 1000)
//fmt.Print("\nContinuare con la get: ")
//var choice int
//_, err = fmt.Scan(&choice)

func run1(rpcName string, done chan bool) {

	executeCall(rpcName+put, "y", "0")
	executeCall(rpcName+put, "x", "1")
	executeCall(rpcName+get, "x")
	executeCall(rpcName+get, "y")

	done <- true
}

func run2(rpcName string, done chan bool) {

	executeCall(rpcName+put, "y", "1")
	executeCall(rpcName+get, "y")
	executeCall(rpcName+get, "x")

	done <- true
}

func run3(rpcName string, done chan bool) {
	executeCall(rpcName+get, "x")
	executeCall(rpcName+get, "y")
	done <- true
}

func run4(rpcName string, done chan bool) {
	executeCall(rpcName+put, "x", "0")
	executeCall(rpcName+get, "y")
	executeCall(rpcName+get, "x")

	done <- true
}
