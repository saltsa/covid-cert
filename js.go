// +build js,wasm

package main

import (
	"syscall/js"

	log "github.com/sirupsen/logrus"
)

func verifyJSData(this js.Value, args []js.Value) interface{} {
	log.Info("verify JS data")

	if len(args) < 1 {
		log.Info("args len is zero")
		return false
	}
	res := doValidation([]byte(args[0].String()))
	log.Infof("output: %v", res)
	return res
}

func registerCallbacks() bool {
	js.Global().Set("goVerify", js.FuncOf(verifyJSData))
	return true
}
