package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"time"
)

// ----- Stampa messaggi ----- //

const layoutTime = "15:04:05.0000"

func printDebugBlue(blueString string, message Message, clockServer ClockServer) {
	if common.GetDebug() {

		// Ottieni l'orario corrente
		now := time.Now()
		// Formatta l'orario corrente come stringa nel formato desiderato
		formattedTime := now.Format(layoutTime)

		fmt.Println(color.BlueString(blueString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(), "clockClient", message.GetOrderClient(),
			"clockMsg:", message.GetClock(), "clockServer:", clockServer.GetClock(), formattedTime)
	}
}

func printGreen(greenString string, message Message, clockServer ClockServer) {
	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	fmt.Println(color.GreenString(greenString), message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(), "clockClient", message.GetOrderClient(),
		"clockMsg:", message.GetClock(), "clockServer:", clockServer.GetClock(), "idSender", message.GetIdSender(), formattedTime)
}

func printDebugBlueArgs(blueString string, args common.Args) {
	if common.GetDebug() {
		fmt.Println(color.BlueString(blueString), args.Key+":"+args.Value, "msg clock:", args.Timestamp)
	}
}

func printRed(redString string, message Message, clockServer ClockServer) {

	// Ottieni l'orario corrente
	now := time.Now()
	// Formatta l'orario corrente come stringa nel formato desiderato
	formattedTime := now.Format(layoutTime)

	fmt.Println(color.GreenString(redString), message.GetTypeOfMessage(), message.GetKey(), "datastore:", clockServer.GetDatastore(),
		"clockMsg:", message.GetClock(), "clockServer:", clockServer.GetClock(), formattedTime)
}
