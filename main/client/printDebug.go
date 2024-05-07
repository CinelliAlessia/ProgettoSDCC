package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"strings"
)

/* ----- FUNZIONI PER PRINT DI DEBUG ----- */

func debugPrintRun(rpcName string, args common.Args) {

	debugName := strings.SplitAfter(rpcName, ".")
	name := debugName[1]

	switch name {
	case common.Put:
		//if args.GetKey() != common.EndKey {
		fmt.Println(color.BlueString("RUN Put"), args.GetKey()+":"+args.GetValue())
		//}
	case common.Get:
		fmt.Println(color.BlueString("RUN Get"), args.GetKey())
	case common.Del:
		fmt.Println(color.BlueString("RUN Delete"), args.GetKey())
	default:
		fmt.Println(color.BlueString("RUN Unknown"), rpcName, args)
	}

}

func debugPrintResponse(rpcName string, indexServer string, args common.Args, response common.Response) {

	debugName := strings.SplitAfter(rpcName, ".")
	name := debugName[1]

	switch name {
	case common.Put:
		//if args.GetKey() != common.EndKey {
		fmt.Println(color.GreenString("RISPOSTA Put "+indexServer), "key:"+args.GetKey(), "value:"+args.GetValue())
		//}

	case common.Get:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Get "+indexServer), "key:"+args.GetKey(), "response:"+response.GetValue())
		} else {
			fmt.Println(color.RedString("RISPOSTA Get fallita "+indexServer), "key:"+args.GetKey())
		}
	case common.Del:
		if response.Result {
			fmt.Println(color.GreenString("RISPOSTA Delete "+indexServer), "key:"+args.GetKey())
		} else {
			fmt.Println(color.RedString("RISPOSTA Delete fallita "+indexServer), "key:"+args.GetKey())
		}
	default:
		fmt.Println(color.GreenString("RISPOSTA Unknown "+indexServer), rpcName, args, response.GetValue())
	}
}
