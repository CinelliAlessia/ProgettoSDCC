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

	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	switch message := message.(type) {
	case commonMsg.MessageS:
		if kvs != nil {
			if common.GetDebug() {
				fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
					"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvs.GetClock(), formattedTime)
			} else {
				fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
			}
		} else {
			fmt.Println("ERRORE")
		}
	case commonMsg.MessageC:
		if kvc != nil {
			if common.GetDebug() {
				fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
					"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
			} else {
				fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
			}
		} else {
			fmt.Println("ERRORE")
		}
	default:
		fmt.Println("ERRORE", message)
	}

}

func printGreen(greenString string, message interface{}, kvc *KeyValueStoreCausale, kvs *KeyValueStoreSequential) {
	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	switch message := message.(type) {
	case commonMsg.MessageS:
		if common.GetDebug() {
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
				"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvs.GetClock(), formattedTime)
		} else {
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
		}
	case commonMsg.MessageC:
		if common.GetDebug() {
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
				"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
		} else {
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
		}
	}
}

func printRed(redString string, message interface{}, kvc *KeyValueStoreCausale, kvs *KeyValueStoreSequential) {

	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	switch message := message.(type) {
	case commonMsg.MessageS:
		if common.GetDebug() {
			fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey(), "datastore:", kvs.GetDatastore(),
				"clockMsg:", message.GetClock(), "clockServer:", kvs.GetClock(), formattedTime)
		} else {
			fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey())
		}
	case commonMsg.MessageC:
		if common.GetDebug() {
			fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey(), "datastore:", kvc.GetDatastore(),
				"clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
		} else {
			fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey())
		}
	}
}
