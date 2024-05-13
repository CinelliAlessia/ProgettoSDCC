package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"strconv"
	"strings"
)

/* ----- FUNZIONI PER PRINT DI DEBUG ----- */

// debugPrintRun, stampa a schermo il tipo di operazione che si sta eseguendo
func debugPrintRun(rpcName string, args common.Args) {

	debugName := strings.SplitAfter(rpcName, ".")
	name := debugName[1]

	if common.GetDebug() {
		switch name {
		case common.Put:
			fmt.Println(color.BlueString("RUN Put"), args.GetKey()+":"+args.GetValue(), "ts client", args.GetSendingFIFO())
		case common.Get:
			fmt.Println(color.BlueString("RUN Get"), args.GetKey(), "ts client", args.GetSendingFIFO())
		case common.Del:
			fmt.Println(color.BlueString("RUN Delete"), args.GetKey(), "ts client", args.GetSendingFIFO())
		}
	} else {
		switch name {
		case common.Put:
			fmt.Println(color.BlueString("RUN Put"), args.GetKey()+":"+args.GetValue())
		case common.Get:
			fmt.Println(color.BlueString("RUN Get"), args.GetKey())
		case common.Del:
			fmt.Println(color.BlueString("RUN Delete"), args.GetKey())
		}
	}

}

// debugPrintResponse, stampa a schermo la risposta che il client ha ricevuto dal server
func debugPrintResponse(rpcName string, indexServer int, args common.Args, response common.Response) {
	debugName := strings.SplitAfter(rpcName, ".")
	name := debugName[1]

	index := strconv.Itoa(indexServer)
	stringName := "RISPOSTA " + name + " " + index

	switch name {
	case common.Put:
		fmt.Println(color.GreenString(stringName), "key:"+args.GetKey(), "value:"+args.GetValue())
	case common.Get:
		if response.Result {
			fmt.Println(color.GreenString(stringName), "key:"+args.GetKey(), "response:"+response.GetValue())
		} else {
			fmt.Println(color.RedString("RISPOSTA Get fallita "+index), "key:"+args.GetKey())
		}
	case common.Del:
		fmt.Println(color.GreenString(stringName), "key:"+args.GetKey())
	}
}
