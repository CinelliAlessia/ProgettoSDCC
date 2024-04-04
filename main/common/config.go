package common

import "github.com/google/uuid"

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
