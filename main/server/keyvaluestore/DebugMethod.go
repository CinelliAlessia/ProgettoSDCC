package keyvaluestore

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"main/server/message"
	"time"
)

// ----- Stampa messaggi ----- //

const layoutTime = "15:04:05.0000"

func printDebugBlue(blueString string, message interface{}, kvc *KeyValueStoreCausale, kvs *KeyValueStoreSequential) {
	if common.GetDebug() {

		// Ottieni l'orario corrente
		now := time.Now()
		// Formatta l'orario corrente come stringa nel formato desiderato
		formattedTime := now.Format(layoutTime)

		switch message := message.(type) {
		case msg.MessageS:
			if kvs != nil {
				fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
					"clockClient", message.GetOrderClient(), "clockMsg:", message.GetClock(), "clockServer:", kvs.GetClock(), formattedTime)
			} else {
				fmt.Println("ERRORE")
			}
		case msg.MessageC:
			if kvc != nil {
				fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
					"clockClient", message.GetOrderClient(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
			} else {
				fmt.Println("ERRORE")
			}
		default:
			fmt.Println("ERRORE", message)
		}
	}
}

func printGreen(greenString string, message interface{}, kvc *KeyValueStoreCausale, kvs *KeyValueStoreSequential) {
	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	switch message := message.(type) {
	case msg.MessageS:
		fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
			"clockClient", message.GetOrderClient(), "clockMsg:", message.GetClock(), "clockServer:", kvs.GetClock(), formattedTime)
	case msg.MessageC:
		fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
			"clockClient", message.GetOrderClient(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
	}
}

func printRed(redString string, message interface{}, kvc *KeyValueStoreCausale, kvs *KeyValueStoreSequential) {

	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	switch message := message.(type) {
	case msg.MessageS:
		fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey(), "datastore:", kvs.GetDatastore(),
			"clockMsg:", message.GetClock(), "clockServer:", kvs.GetClock(), formattedTime)
	case msg.MessageC:
		fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey(), "datastore:", kvc.GetDatastore(),
			"clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
	}
}
