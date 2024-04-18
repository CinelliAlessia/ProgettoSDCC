package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"time"
)

func (kvc *KeyValueStoreCausale) printDebugBlue(blueString string, message MessageC) {
	if common.GetDebug() {
		// Ottieni l'orario corrente
		now := time.Now()

		// Formatta l'orario corrente come stringa nel formato desiderato
		formattedTime := now.Format("15:04:05.000")

		fmt.Println(color.BlueString(blueString), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.VectorClock, "my clock:", kvc.VectorClock, formattedTime)
	}
}

func (kvs *KeyValueStoreSequential) printDebugBlue(blueString string, message MessageS) {
	if common.GetDebug() {
		// Ottieni l'orario corrente
		now := time.Now()

		// Formatta l'orario corrente come stringa nel formato desiderato
		formattedTime := now.Format("15:04:05.000")

		fmt.Println(color.BlueString(blueString), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.LogicalClock, formattedTime)
	}
}
func (kvc *KeyValueStoreCausale) printGreen(greenString string, message MessageC) {
	// Ottieni l'orario corrente
	now := time.Now()

	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format("15:04:05.000")

	fmt.Println(color.GreenString(greenString), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.VectorClock, "my clock:", kvc.VectorClock, formattedTime)
	//printDatastore(kvc)
}

func (kvs *KeyValueStoreSequential) printGreen(greenString string, message MessageS) {

	// Ottieni l'orario corrente
	now := time.Now()

	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format("15:04:05.000")
	fmt.Println(color.GreenString(greenString), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.LogicalClock, formattedTime)

	//printDatastore(kvs)
}

func (kvs *KeyValueStoreSequential) printRed(redString string, message MessageS) {

	// Ottieni l'orario corrente
	now := time.Now()

	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format("15:04:05.000")

	fmt.Println(color.RedString(redString), message.TypeOfMessage, message.Args.Key, "datastore:", kvs.Datastore, "msg clock:", message.LogicalClock, "my clock:", kvs.LogicalClock, formattedTime)
	//printDatastore(kvs)
}
