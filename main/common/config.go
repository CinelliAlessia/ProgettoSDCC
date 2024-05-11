// Package common fornisce funzioni e variabili globali utilizzate in tutto il sistema.
package common

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	Replicas = 3    // Numero di repliche del server -> almeno 3
	TIMER    = 1000 // Ritardo di rete max espresso in Millisecondi
)

const (
	EndKey   = "endKey"
	EndValue = "endValue"

	Sequential = "KeyValueStoreSequential"
	Causal     = "KeyValueStoreCausale"

	PutRPC = ".Put"
	GetRPC = ".Get"
	DelRPC = ".Delete"

	Put = "Put"
	Get = "Get"
	Del = "Delete"
)

// ReplicaPorts Lista delle porte su cui le repliche possono essere contattate
var ReplicaPorts = [Replicas]string{"8085", "8086", "8087"}

// GenerateUniqueID Genera un ID univoco utilizzando UUID
func GenerateUniqueID() string {
	id := uuid.New()
	return id.String()
}

// Max Restituisce il massimo tra due numeri interi
// TODO: Controllare perché la funzione max non funziona
func Max(clock int, clock2 int) int {
	if clock > clock2 {
		return clock
	} else {
		return clock2
	}
}

// GetServerName Restituisce il nome del server in base alla configurazione scelta, tra locale e remota (docker)
// L'indice id parte da 0 fino a (Replicas-1)
func GetServerName(replicaPort string, id int) string {
	if os.Getenv("CONFIG") == "0" {
		return ":" + replicaPort // Locale
	} else if os.Getenv("CONFIG") == "1" {
		return "server" + strconv.Itoa(id+1) + ":" + replicaPort // Docker
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA, CONFIG:", os.Getenv("CONFIG"))
		return ""
	}
}

// GetDebug Restituisce un booleano in base alla variabile di ambiente DEBUG impostata:
//   - true: Esegue più print di debug.
//   - false: Esegue solamente i print necessari.
func GetDebug() bool {
	if os.Getenv("DEBUG") == "1" {
		return true
	} else if os.Getenv("DEBUG") == "0" {
		return false
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA DEBUG:", os.Getenv("DEBUG"))
		return true
	}
}

// RandomDelay Genera un ritardo casuale tra 0 e TIMER-1 millisecondi
func RandomDelay() {
	delay := rand.Intn(TIMER)

	// Introduce un ritardo casuale
	time.Sleep(time.Millisecond * time.Duration(delay))
}

// UseEndKey Utilizzo della chiave speciale "endKey" per terminare le operazioni
func UseEndKey() bool {
	if os.Getenv("USEENDKEY") == "1" {
		return true
	} else if os.Getenv("USEENDKEY") == "0" {
		return false
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA DEBUG:", os.Getenv("DEBUG"))
		return true
	}
}
