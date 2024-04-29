package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"strings"
)

const (
	put = ".Put"
	get = ".Get"
	del = ".Delete"
)

const (
	sync  = "sync"
	async = "async"
)

const (
	random   = "random"
	specific = "specific"
)

type Operation struct {
	ServerIndex   int
	OperationType string
	Key           string
	Value         string
}

/* ----- FUNZIONI PER PRINT DI DEBUG ----- */

func debugPrintRun(rpcName string, args common.Args) {
	if common.GetDebug() {
		debugName := strings.SplitAfter(rpcName, ".")
		name := "." + debugName[1]

		switch name {
		case put:
			fmt.Println(color.BlueString("RUN Put"), args.Key+":"+args.Value)
		case get:
			fmt.Println(color.BlueString("RUN Get"), args.Key)
		case del:
			fmt.Println(color.BlueString("RUN Delete"), args.Key)
		default:
			fmt.Println(color.BlueString("RUN Unknown"), rpcName, args)
		}
	} else {
		return
	}
}

func debugPrintResponse(rpcName string, args common.Args, response common.Response) {

	debugName := strings.SplitAfter(rpcName, ".")
	name := "." + debugName[1]

	switch name {
	case put:
		fmt.Println(color.GreenString("RISPOSTA Put"), "key:"+args.Key, "value:"+args.Value)
	case get:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Get"), "key:"+args.Key, "response:"+response.Value)
		} else {
			fmt.Println(color.RedString("RISPOSTA Get fallita"), "key:"+args.Key)
		}
	case del:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Delete"), "key:"+args.Key)
		} else {
			fmt.Println(color.RedString("RISPOSTA Delete fallita"), "key:"+args.Key)
		}
	default:
		fmt.Println(color.GreenString("RISPOSTA Unknown"), rpcName, args, response)
	}
}
