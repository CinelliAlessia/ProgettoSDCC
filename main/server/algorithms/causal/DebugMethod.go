package causal

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"main/server/message"
	"time"
)

// ----- Stampa messaggi ----- //

const layoutTime = "15:04:05.0000"

// PrintDebugBlue Usato per messaggi di ricezione
func printDebugBlue(blueString string, message commonMsg.MessageC, kvc *KeyValueStoreCausale) {

	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	if common.GetDebug() {
		switch message.GetTypeOfMessage() {
		case common.Del:
			fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey(),
				"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
		default:
			fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
				"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
		}
	} else {
		switch message.GetTypeOfMessage() {
		case common.Del:
			fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey())
		default:
			fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
		}
	}

}

// PrintGreen Usato per messaggi di invio
func printGreen(greenString string, message commonMsg.MessageC, kvc *KeyValueStoreCausale) {
	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	if common.GetDebug() {
		switch message.GetTypeOfMessage() {
		case common.Del:
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey(),
				"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)

		default:
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
				"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
		}

	} else {
		switch message.GetTypeOfMessage() {
		case common.Del:
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey(), kvc.GetDatastore())
		default:
			fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(), kvc.GetDatastore())
		}
	}
}

// PrintRed Usato per messaggi di errore
func printRed(redString string, message commonMsg.MessageC, kvc *KeyValueStoreCausale) {

	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	if common.GetDebug() {
		fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey(), "datastore:", kvc.GetDatastore(),
			"clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
	} else {
		fmt.Println(color.RedString(redString), message.GetTypeOfMessage(), message.GetKey())
	}

}
