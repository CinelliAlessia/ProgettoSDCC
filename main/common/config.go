package common

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"strconv"
)

// Configurazione globale del programma

// Replicas Numero di repliche del server -> almeno 3
const Replicas = 3

// ReplicaPorts Lista delle porte su cui le repliche possono essere contattate
var ReplicaPorts = [Replicas]string{"8085", "8086", "8087"}

// GenerateUniqueID Genera un ID univoco utilizzando UUID
func GenerateUniqueID() string {
	id := uuid.New()
	return id.String()
}

func Max(clock int, clock2 int) int {
	if clock > clock2 {
		return clock
	} else {
		return clock2
	}
}

// GetServerName restituisce il nome del server in base alla configurazione
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
