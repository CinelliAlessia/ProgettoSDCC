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
			fmt.Println(color.BlueString("RUN Put"), args.GetKey()+":"+args.GetValue())
		case get:
			fmt.Println(color.BlueString("RUN Get"), args.GetKey())
		case del:
			fmt.Println(color.BlueString("RUN Delete"), args.GetKey())
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
		fmt.Println(color.GreenString("RISPOSTA Put"), "key:"+args.GetKey(), "value:"+args.GetValue())
	case get:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Get"), "key:"+args.GetKey(), "response:"+response.GetValue())
		} else {
			fmt.Println(color.RedString("RISPOSTA Get fallita"), "key:"+args.GetKey())
		}
	case del:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Delete"), "key:"+args.GetKey())
		} else {
			fmt.Println(color.RedString("RISPOSTA Delete fallita"), "key:"+args.GetKey())
		}
	default:
		fmt.Println(color.GreenString("RISPOSTA Unknown"), rpcName, args, response.GetValue())
	}
}
