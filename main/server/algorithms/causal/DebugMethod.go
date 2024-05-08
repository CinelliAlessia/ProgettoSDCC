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
		fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
			"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
	} else {
		fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
	}

}

func printGreen(greenString string, message commonMsg.MessageC, kvc *KeyValueStoreCausale) {
	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	if common.GetDebug() {
		fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
			"clockClient", message.GetSendingFIFO(), "clockMsg:", message.GetClock(), "clockServer:", kvc.GetClock(), formattedTime)
	} else {
		fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
	}
}

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
