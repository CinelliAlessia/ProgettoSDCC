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

// Replicas Numero di repliche del server -> almeno 3
const Replicas = 3

// ReplicaPorts Lista delle porte su cui le repliche possono essere contattate
var ReplicaPorts = [Replicas]string{"8085", "8086", "8087"}

// GenerateUniqueID Genera un ID univoco utilizzando UUID
func GenerateUniqueID() string {
	id := uuid.New()
	return id.String()
}

// Max TODO: Controllare perché la funzione max non funziona
func Max(clock int, clock2 int) int {
	if clock > clock2 {
		return clock
	} else {
		return clock2
	}
}

// GetServerName restituisce il nome del server in base alla scelta di configurazione, tra locale e remota (docker)
func GetServerName(replicaPort string, id int) string {
	if os.Getenv("CONFIG") == "1" {
		return ":" + replicaPort // Locale
	} else if os.Getenv("CONFIG") == "2" {
		return "server" + strconv.Itoa(id+1) + ":" + replicaPort // Docker
	} else {
		fmt.Println("VARIABILE DI AMBIENTE ERRATA")
		return ""
	}
}

// RandomDelay Genera un numero casuale compreso tra 0 e 999 (max un secondo)
func RandomDelay() {
	delay := rand.Intn(1000)

	// Introduce un ritardo casuale
	time.Sleep(time.Millisecond * time.Duration(delay))
}
